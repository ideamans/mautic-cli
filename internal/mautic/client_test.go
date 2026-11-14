package mautic

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func testClient(t *testing.T, handler http.HandlerFunc) (*Client, *httptest.Server) {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	c := New(Options{
		BaseURL: srv.URL,
		Auth:    &BasicAuth{Username: "u", Password: "p"},
		Retries: 2,
	})
	c.backoff = time.Millisecond // keep the retry tests quick
	return c, srv
}

func TestRequestPathAndBasicAuth(t *testing.T) {
	var gotPath, gotAuth, gotAccept string
	c, _ := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath, gotAuth, gotAccept = r.URL.Path, r.Header.Get("Authorization"), r.Header.Get("Accept")
		w.Write([]byte(`{"ok":true}`))
	})

	if _, err := c.Get(context.Background(), "/contacts", nil); err != nil {
		t.Fatalf("Get: %v", err)
	}
	if gotPath != "/api/contacts" {
		t.Errorf("path = %q, want /api/contacts", gotPath)
	}
	want := "Basic " + base64.StdEncoding.EncodeToString([]byte("u:p"))
	if gotAuth != want {
		t.Errorf("Authorization = %q", gotAuth)
	}
	if gotAccept != "application/json" {
		t.Errorf("Accept = %q", gotAccept)
	}
}

func TestRetriesOn429ThenSucceeds(t *testing.T) {
	var calls int32
	c, _ := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		if atomic.AddInt32(&calls, 1) == 1 {
			w.Header().Set("Retry-After", "0")
			w.WriteHeader(http.StatusTooManyRequests)
			w.Write([]byte(`{"error":{"message":"slow down","code":429}}`))
			return
		}
		w.Write([]byte(`{"ok":true}`))
	})

	if _, err := c.Get(context.Background(), "/contacts", nil); err != nil {
		t.Fatalf("Get should have succeeded after a retry: %v", err)
	}
	if calls != 2 {
		t.Errorf("calls = %d, want 2", calls)
	}
}

// A 4xx that is not 429 is the caller's fault, so retrying it only wastes the
// user's rate limit.
func TestDoesNotRetryClientErrors(t *testing.T) {
	var calls int32
	c, _ := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&calls, 1)
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte(`{"error":{"message":"You do not have access to the requested area/action.","code":403}}`))
	})

	_, err := c.Get(context.Background(), "/contacts", nil)
	if err == nil {
		t.Fatal("expected an error")
	}
	if calls != 1 {
		t.Errorf("calls = %d, want 1 — a 403 must not be retried", calls)
	}
	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("error type = %T, want *APIError", err)
	}
	if apiErr.Status != 403 || !strings.Contains(apiErr.Message, "do not have access") {
		t.Errorf("unexpected error: %+v", apiErr)
	}
	// The hint is what turns a status code into a next action.
	if !strings.Contains(apiErr.Error(), "Settings -> Roles") {
		t.Errorf("403 should suggest checking the role:\n%v", apiErr)
	}
}

// Mautic uses three different error envelopes depending on which layer
// rejected the request.
func TestParsesEveryErrorEnvelope(t *testing.T) {
	cases := []struct {
		name string
		body string
		want string
	}{
		{"system", `{"error":{"message":"boom","code":500}}`, "boom"},
		{"errors array", `{"errors":[{"message":"not found","code":404}]}`, "not found"},
		{"oauth", `{"error":"invalid_grant","error_description":"The access token provided has expired."}`, "expired"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			e := parseError(400, "GET", "/x", []byte(tc.body))
			if !strings.Contains(e.Message, tc.want) {
				t.Errorf("Message = %q, want it to contain %q", e.Message, tc.want)
			}
		})
	}
}

// An HTML login page is the classic symptom of a wrong URL or a disabled API,
// and it must not surface as a JSON parse error.
func TestNonJSONResponseExplainsItself(t *testing.T) {
	c, _ := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("<!doctype html><html><body>Login</body></html>"))
	})
	_, err := c.Get(context.Background(), "/contacts", nil)
	if err == nil {
		t.Fatal("expected an error")
	}
	if !strings.Contains(err.Error(), "API is enabled") {
		t.Errorf("the error should point at API Settings:\n%v", err)
	}
}

func TestDryRunSendsNothingButShowsTheRequest(t *testing.T) {
	var calls int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&calls, 1)
		w.Write([]byte(`{"ok":true}`))
	}))
	defer srv.Close()

	c := New(Options{BaseURL: srv.URL, Auth: &BasicAuth{Username: "u", Password: "p"}, DryRun: true})
	raw, err := c.Do(context.Background(), "POST", "/contacts/new", nil, map[string]any{"email": "a@example.com"})
	if err != nil {
		t.Fatalf("Do: %v", err)
	}
	if calls != 0 {
		t.Fatalf("dry run made %d requests, want 0", calls)
	}

	var result DryRunResult
	if err := json.Unmarshal(raw, &result); err != nil {
		t.Fatalf("dry run output is not a DryRunResult: %v", err)
	}
	if !result.DryRun || result.Method != "POST" {
		t.Errorf("unexpected result: %+v", result)
	}
	// The body must be visible: reviewing a dry run means reading what would
	// have been sent.
	if !strings.Contains(string(result.Body), "a@example.com") {
		t.Errorf("body missing from the dry run: %s", result.Body)
	}
	// And so must the instance, so "I would write to production" is legible.
	if result.Instance != srv.URL {
		t.Errorf("Instance = %q, want %q", result.Instance, srv.URL)
	}
}

// A GET is not a write, so --dry-run must not suppress it: an agent inspecting
// before acting still needs to be able to read.
func TestDryRunStillPerformsReads(t *testing.T) {
	var calls int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&calls, 1)
		w.Write([]byte(`{"total":0,"contacts":{}}`))
	}))
	defer srv.Close()

	c := New(Options{BaseURL: srv.URL, Auth: &BasicAuth{Username: "u", Password: "p"}, DryRun: true})
	if _, err := c.Get(context.Background(), "/contacts", nil); err != nil {
		t.Fatalf("Get: %v", err)
	}
	if calls != 1 {
		t.Errorf("calls = %d, want 1", calls)
	}
}

func TestListAllPagesUntilTotal(t *testing.T) {
	c, _ := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		start := r.URL.Query().Get("start")
		switch start {
		case "", "0":
			w.Write([]byte(`{"total":3,"contacts":{"1":{"id":1},"2":{"id":2}}}`))
		default:
			w.Write([]byte(`{"total":3,"contacts":{"3":{"id":3}}}`))
		}
	})

	page, err := c.ListAll(context.Background(), "/contacts", ListParams{}, "contacts", 2, 0)
	if err != nil {
		t.Fatalf("ListAll: %v", err)
	}
	if len(page.Items) != 3 {
		t.Errorf("fetched %d records, want 3", len(page.Items))
	}
	if page.Total != 3 {
		t.Errorf("Total = %d, want 3", page.Total)
	}
}

// An endpoint that keeps reporting more than it returns must not spin forever.
func TestListAllStopsOnAnEmptyPage(t *testing.T) {
	var calls int32
	c, _ := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		if atomic.AddInt32(&calls, 1) == 1 {
			w.Write([]byte(`{"total":999,"contacts":{"1":{"id":1}}}`))
			return
		}
		w.Write([]byte(`{"total":999,"contacts":{}}`))
	})

	page, err := c.ListAll(context.Background(), "/contacts", ListParams{}, "contacts", 10, 0)
	if err != nil {
		t.Fatalf("ListAll: %v", err)
	}
	if len(page.Items) != 1 {
		t.Errorf("got %d records, want 1", len(page.Items))
	}
	if calls != 2 {
		t.Errorf("calls = %d, want 2", calls)
	}
}

func TestListAllHonoursTheMaximum(t *testing.T) {
	c, _ := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"total":100,"contacts":{"1":{"id":1},"2":{"id":2},"3":{"id":3}}}`))
	})
	page, err := c.ListAll(context.Background(), "/contacts", ListParams{}, "contacts", 3, 2)
	if err != nil {
		t.Fatalf("ListAll: %v", err)
	}
	if len(page.Items) != 2 {
		t.Errorf("got %d records, want the requested maximum of 2", len(page.Items))
	}
}

func TestCapturesMauticVersionHeader(t *testing.T) {
	c, _ := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Mautic-Version", "7.0.1")
		w.Write([]byte(`{"ok":true}`))
	})
	if _, err := c.Get(context.Background(), "/x", nil); err != nil {
		t.Fatal(err)
	}
	if c.MauticVersion() != "7.0.1" {
		t.Errorf("MauticVersion = %q", c.MauticVersion())
	}
}

// Mautic serialises an empty `details` as `[]` rather than `{}`, because PHP
// does. Decoding it as a map fails, and the whole message was being lost.
func TestErrorDetailsAsEmptyArray(t *testing.T) {
	e := parseError(404, "GET", "/contacts/1",
		[]byte(`{"errors":[{"code":404,"message":"Item was not found.","details":[]}]}`))
	if e.Message != "Item was not found." {
		t.Errorf("Message = %q, want the decoded message", e.Message)
	}
	if e.Details != nil {
		t.Errorf("Details = %v, want nil for an empty array", e.Details)
	}
	if strings.Contains(e.Error(), `{"errors"`) {
		t.Errorf("the raw body should not appear once the message parsed:\n%v", e)
	}
}

// A populated details object is field-level validation feedback and must
// survive, because it names which field was rejected.
func TestErrorDetailsAsObject(t *testing.T) {
	e := parseError(422, "POST", "/contacts/new",
		[]byte(`{"errors":[{"code":422,"message":"invalid","details":{"email":["not valid"]}}]}`))
	if len(e.Details) != 1 {
		t.Fatalf("Details = %v, want the email entry", e.Details)
	}
	if !strings.Contains(e.Error(), "email") {
		t.Errorf("the rejected field should be named:\n%v", e)
	}
}

// Mautic can save a record and then fail with 500 while building the response.
// Repeating a POST would create the record again, so writes are not retried on
// 5xx (a 429 is still retried: the request was refused before anything ran).
func TestDoesNotRetryPostOn5xx(t *testing.T) {
	var calls int32
	c, _ := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&calls, 1)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"errors":[{"message":"Looks like I encountered an error","code":500}]}`))
	})
	if _, err := c.Do(context.Background(), "POST", "/forms/new", nil, []byte(`{"name":"x"}`)); err == nil {
		t.Fatal("POST should have failed")
	}
	if _, err := c.Do(context.Background(), "PATCH", "/forms/1/edit", nil, []byte(`{"name":"x"}`)); err == nil {
		t.Fatal("PATCH should have failed")
	}
	if calls != 2 {
		t.Errorf("calls = %d, want 2 (one each, no retries)", calls)
	}
}

func TestRetriesPostOn429(t *testing.T) {
	var calls int32
	c, _ := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		if atomic.AddInt32(&calls, 1) == 1 {
			w.Header().Set("Retry-After", "0")
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		w.Write([]byte(`{"form":{"id":1}}`))
	})
	if _, err := c.Do(context.Background(), "POST", "/forms/new", nil, []byte(`{"name":"x"}`)); err != nil {
		t.Fatalf("POST should have succeeded after a 429: %v", err)
	}
	if calls != 2 {
		t.Errorf("calls = %d, want 2", calls)
	}
}

func TestZeroRetriesMeansNone(t *testing.T) {
	var calls int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&calls, 1)
		w.WriteHeader(http.StatusBadGateway)
	}))
	t.Cleanup(srv.Close)
	c := New(Options{BaseURL: srv.URL, Auth: &BasicAuth{Username: "u", Password: "p"}, Retries: 0})
	c.backoff = time.Millisecond
	if _, err := c.Get(context.Background(), "/contacts", nil); err == nil {
		t.Fatal("GET should have failed")
	}
	if calls != 1 {
		t.Errorf("calls = %d, want 1 (--retries 0 must not retry)", calls)
	}
}
