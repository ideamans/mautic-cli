package mautic

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// APIError is a non-2xx response from Mautic, decoded into something an agent
// can act on. Mautic returns three different error envelopes depending on which
// layer rejected the request, so all three are parsed into this one type.
type APIError struct {
	// Status is the HTTP status code.
	Status int
	// Code is Mautic's own code, which usually equals Status but not always.
	Code int
	// Message is the human readable reason.
	Message string
	// Type is Mautic's error type, when present.
	Type string
	// Details holds field-level validation errors, keyed by field alias.
	Details map[string]any
	// Body is the raw response, truncated, for when nothing else parsed.
	Body string
	// Method and Path identify the request that failed.
	Method string
	Path   string

	// retryAfter carries the Retry-After header so the retry loop honours the
	// server's own pacing instead of guessing.
	retryAfter time.Duration
}

func (e *APIError) Error() string {
	var b strings.Builder
	fmt.Fprintf(&b, "%s %s failed: HTTP %d", e.Method, e.Path, e.Status)
	if e.Message != "" {
		fmt.Fprintf(&b, ": %s", e.Message)
	} else if e.Body != "" {
		fmt.Fprintf(&b, ": %s", e.Body)
	}
	if len(e.Details) > 0 {
		keys := sortedKeys(e.Details)
		fmt.Fprintf(&b, "\n  rejected fields: %s", strings.Join(keys, ", "))
		for _, k := range keys {
			fmt.Fprintf(&b, "\n    %s: %v", k, e.Details[k])
		}
	}
	if hint := e.hint(); hint != "" {
		fmt.Fprintf(&b, "\n  %s", hint)
	}
	return b.String()
}

// hint turns the statuses an agent actually hits into the next action, rather
// than leaving it to guess.
func (e *APIError) hint() string {
	switch e.Status {
	case 401:
		return "check credentials with `mautic auth status`; for OAuth2 the token may have expired — run `mautic auth login`"
	case 403:
		return "the Mautic user's role lacks permission for this area; grant it under Settings -> Roles, or check `mautic users self`"
	case 404:
		return "the record does not exist, or the API is disabled for this instance (Configuration -> API Settings)"
	case 422:
		return "validation failed — check the field aliases with `mautic fields list`"
	case 429:
		return "rate limited by Mautic's API rate limiter; the request was already retried with backoff, so space the calls out"
	}
	if e.Status >= 500 {
		return "a server-side error in Mautic; check the instance's var/logs/prod.log. Some instances return 500 on /contacts while /stats works"
	}
	return ""
}

// IsNotFound reports whether the error is a 404.
func IsNotFound(err error) bool {
	ae, ok := err.(*APIError)
	return ok && ae.Status == 404
}

// parseError decodes whichever error envelope Mautic used.
//
//	{"error":{"message":"...","code":403}}          system errors
//	{"errors":[{"message":"...","code":404}]}       newer bundles
//	{"error":"invalid_grant","error_description":""} OAuth
func parseError(status int, method, path string, body []byte) *APIError {
	e := &APIError{Status: status, Code: status, Method: method, Path: path, Body: truncate(body, 500)}

	// details is `[]` when empty and an object when populated, because PHP
	// serialises an empty associative array as a JSON array. Decoding it
	// straight into a map fails on every error that has no field detail —
	// which is most of them — so it is decoded loosely and narrowed after.
	type errBody struct {
		Message string `json:"message"`
		Code    int    `json:"code"`
		Type    string `json:"type"`
		Details any    `json:"details"`
	}
	var envelope struct {
		Error            json.RawMessage `json:"error"`
		Errors           []errBody       `json:"errors"`
		ErrorDescription string          `json:"error_description"`
	}
	if err := json.Unmarshal(body, &envelope); err != nil {
		return e
	}

	if len(envelope.Errors) > 0 {
		first := envelope.Errors[0]
		e.Message, e.Type, e.Details = first.Message, first.Type, asDetails(first.Details)
		if first.Code != 0 {
			e.Code = first.Code
		}
		return e
	}

	if len(envelope.Error) > 0 {
		// Either an object (system error) or a string (OAuth error).
		var obj errBody
		if err := json.Unmarshal(envelope.Error, &obj); err == nil && obj.Message != "" {
			e.Message, e.Type, e.Details = obj.Message, obj.Type, asDetails(obj.Details)
			if obj.Code != 0 {
				e.Code = obj.Code
			}
			return e
		}
		var s string
		if err := json.Unmarshal(envelope.Error, &s); err == nil {
			e.Type = s
			e.Message = strings.TrimSpace(s + " " + envelope.ErrorDescription)
		}
	}
	return e
}

// asDetails keeps field-level validation errors and discards the empty array
// Mautic sends when there are none.
func asDetails(v any) map[string]any {
	m, ok := v.(map[string]any)
	if !ok || len(m) == 0 {
		return nil
	}
	return m
}

func truncate(b []byte, n int) string {
	s := strings.TrimSpace(string(b))
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}

func sortedKeys(m map[string]any) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	// sort.Strings without importing sort twice in this file's view.
	for i := 1; i < len(keys); i++ {
		for j := i; j > 0 && keys[j] < keys[j-1]; j-- {
			keys[j], keys[j-1] = keys[j-1], keys[j]
		}
	}
	return keys
}
