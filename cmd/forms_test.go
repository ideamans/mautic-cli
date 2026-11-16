package cmd

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ideamans/mautic-cli/internal/mautic"
)

// Mautic resets a form's postAction to "return" when a PATCH body omits it, so
// forms update reads the current value and sends it back.
func TestFormPatchKeepsPostAction(t *testing.T) {
	var gets int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/api/forms/17" {
			t.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
			http.NotFound(w, r)
			return
		}
		gets++
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"form":{"id":17,"postAction":"redirect","postActionProperty":"https://example.com/thanks"}}`))
	}))
	defer srv.Close()
	client := mautic.New(mautic.Options{BaseURL: srv.URL, Auth: &mautic.BasicAuth{Username: "u", Password: "p"}, DryRun: true})
	ctx := context.Background()

	payload := map[string]any{"fields": []any{map[string]any{"alias": "lp_variant", "type": "hidden"}}}
	note, err := preserveFormPostAction(ctx, client, "17", payload)
	if err != nil {
		t.Fatal(err)
	}
	if payload["postAction"] != "redirect" || payload["postActionProperty"] != "https://example.com/thanks" {
		t.Fatalf("postAction was not carried over: %#v", payload)
	}
	if note == "" {
		t.Error("expected a note for stderr")
	}

	// A property the caller set is left alone; postAction is still filled in.
	payload = map[string]any{"postActionProperty": "https://example.com/new"}
	if _, err := preserveFormPostAction(ctx, client, "17", payload); err != nil {
		t.Fatal(err)
	}
	if payload["postAction"] != "redirect" || payload["postActionProperty"] != "https://example.com/new" {
		t.Fatalf("caller's postActionProperty was overwritten: %#v", payload)
	}

	// When the caller sets postAction, nothing is fetched or changed.
	before := gets
	payload = map[string]any{"postAction": "message"}
	note, err = preserveFormPostAction(ctx, client, "17", payload)
	if err != nil || note != "" || gets != before {
		t.Fatalf("explicit postAction should skip the lookup: note=%q err=%v gets=%d->%d", note, err, before, gets)
	}
	if _, ok := payload["postActionProperty"]; ok {
		t.Fatalf("postActionProperty should not be added when postAction was given: %#v", payload)
	}
}
