// Package mautic is a client for the Mautic REST API.
//
// It knows nothing about cobra: it takes already-resolved credentials and
// speaks HTTP. That separation is what lets the command layer stay declarative
// and what makes the raw `mautic api` escape hatch share exactly the same
// request path as the wrapped commands.
package mautic

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// Authenticator applies credentials to an outgoing request.
type Authenticator interface {
	Apply(req *http.Request) error
	// Describe returns a secret-free description for status output.
	Describe() string
}

// Client talks to one Mautic instance.
type Client struct {
	baseURL string
	apiURL  string
	http    *http.Client
	auth    Authenticator
	retries int
	backoff time.Duration
	dryRun  bool
	trace   io.Writer

	// version is the Mautic-Version response header from the last call. Mautic
	// has sent it since 2.4.0, so an empty value means "older than that, or the
	// request never reached Mautic".
	version string
}

// Options configures a Client.
type Options struct {
	// BaseURL is the instance root, without a trailing /api.
	BaseURL string
	// Auth applies credentials. Required.
	Auth Authenticator
	// Timeout bounds a single HTTP request. Zero means 60s.
	Timeout time.Duration
	// Retries is the number of extra attempts for 429 and 5xx. Zero means no
	// retries; a negative value means the default (3). Writes that are not
	// idempotent (POST, PATCH) are retried only on 429 — see retryable.
	Retries int
	// DryRun makes every non-GET request return a description of what would
	// have been sent, without sending it.
	DryRun bool
	// Trace, when non-nil, receives one line per request. Point it at stderr:
	// stdout must carry only the result.
	Trace io.Writer
	// UserAgent overrides the default.
	UserAgent string
}

const defaultUserAgent = "mautic-cli"

// New builds a Client.
func New(o Options) *Client {
	timeout := o.Timeout
	if timeout <= 0 {
		timeout = 60 * time.Second
	}
	retries := o.Retries
	if retries < 0 {
		retries = 3
	}
	base := strings.TrimRight(o.BaseURL, "/")
	return &Client{
		baseURL: base,
		apiURL:  base + "/api",
		http:    &http.Client{Timeout: timeout},
		auth:    o.Auth,
		retries: retries,
		backoff: 500 * time.Millisecond,
		dryRun:  o.DryRun,
		trace:   o.Trace,
	}
}

// BaseURL returns the instance root.
func (c *Client) BaseURL() string { return c.baseURL }

// APIURL returns the REST API root.
func (c *Client) APIURL() string { return c.apiURL }

// MauticVersion returns the Mautic-Version header seen on the last response.
func (c *Client) MauticVersion() string { return c.version }

// DryRun reports whether writes are being suppressed.
func (c *Client) DryRun() bool { return c.dryRun }

// DryRunResult is what a suppressed write returns instead of a response. It is
// deliberately the request itself: an agent reviewing a dry run needs to see
// the body it was about to send, and which instance it was about to hit.
type DryRunResult struct {
	DryRun   bool            `json:"dryRun"`
	Method   string          `json:"method"`
	URL      string          `json:"url"`
	Instance string          `json:"instance"`
	Auth     string          `json:"auth"`
	Body     json.RawMessage `json:"body,omitempty"`
}

// Get performs a GET and decodes the JSON response.
func (c *Client) Get(ctx context.Context, path string, query url.Values) (json.RawMessage, error) {
	return c.Do(ctx, http.MethodGet, path, query, nil)
}

// Do performs one API call against <base>/api<path>.
//
// path always starts with "/". body is marshalled as JSON when non-nil; Mautic
// accepts JSON bodies on POST/PATCH/PUT and this avoids every caller having to
// think about form encoding.
func (c *Client) Do(ctx context.Context, method, path string, query url.Values, body any) (json.RawMessage, error) {
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	endpoint := c.apiURL + path
	if len(query) > 0 {
		endpoint += "?" + query.Encode()
	}

	var payload []byte
	if body != nil {
		var err error
		payload, err = json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("could not encode the request body: %w", err)
		}
	}

	if c.dryRun && method != http.MethodGet {
		result := DryRunResult{
			DryRun:   true,
			Method:   method,
			URL:      endpoint,
			Instance: c.baseURL,
			Auth:     c.auth.Describe(),
			Body:     json.RawMessage(payload),
		}
		return json.Marshal(result)
	}

	var lastErr error
	for attempt := 0; attempt <= c.retries; attempt++ {
		if attempt > 0 {
			wait := c.retryDelay(attempt, lastErr)
			c.tracef("retry %d/%d in %s: %v", attempt, c.retries, wait, lastErr)
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(wait):
			}
		}

		raw, err := c.attempt(ctx, method, path, endpoint, payload)
		if err == nil {
			return raw, nil
		}
		lastErr = err
		if !retryable(method, err) {
			return nil, err
		}
	}
	return nil, lastErr
}

func (c *Client) attempt(ctx context.Context, method, path, endpoint string, payload []byte) (json.RawMessage, error) {
	var reader io.Reader
	if payload != nil {
		reader = bytes.NewReader(payload)
	}
	req, err := http.NewRequestWithContext(ctx, method, endpoint, reader)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", defaultUserAgent)
	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if err := c.auth.Apply(req); err != nil {
		return nil, err
	}

	c.tracef("%s %s", method, endpoint)
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%s %s: %w", method, path, err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("%s %s: could not read the response: %w", method, path, err)
	}
	if v := resp.Header.Get("Mautic-Version"); v != "" {
		c.version = v
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		apiErr := parseError(resp.StatusCode, method, path, data)
		if ra := resp.Header.Get("Retry-After"); ra != "" {
			if secs, convErr := strconv.Atoi(ra); convErr == nil {
				apiErr.retryAfter = time.Duration(secs) * time.Second
			}
		}
		return nil, apiErr
	}

	// A 204 or an empty 200 is a success with nothing to decode.
	if len(bytes.TrimSpace(data)) == 0 {
		return json.RawMessage(`{}`), nil
	}
	if !json.Valid(data) {
		// Almost always an HTML login page: the instance redirected because the
		// API is off, or the URL points at the site rather than the API.
		return nil, &APIError{
			Status: resp.StatusCode, Code: resp.StatusCode, Method: method, Path: path,
			Message: "the response was not JSON — check that the URL is the Mautic instance root and that the API is enabled under Configuration -> API Settings",
			Body:    truncate(data, 300),
		}
	}
	return json.RawMessage(data), nil
}

func (c *Client) retryDelay(attempt int, err error) time.Duration {
	if ae, ok := err.(*APIError); ok && ae.retryAfter > 0 {
		return ae.retryAfter
	}
	// 500ms, 1s, 2s, ...
	return c.backoff << (attempt - 1)
}

// retryable decides whether a failed attempt is worth repeating.
//
// A 429 means the request was turned away before Mautic did anything, so it is
// safe to repeat for every method. A 5xx or a dropped connection is different:
// Mautic often saves the record and only then fails while building the response
// (seen with POST /forms/new, which created the form and returned 500 — three
// retries left four identical forms). Repeating a GET, PUT or DELETE is
// harmless, but repeating a POST creates the record again and a PATCH may apply
// twice, so those are not retried. The caller gets the error and can check what
// was saved before trying again.
func retryable(method string, err error) bool {
	ae, ok := err.(*APIError)
	if ok && ae.Status == 429 {
		return true
	}
	if method == http.MethodPost || method == http.MethodPatch {
		return false
	}
	if !ok {
		// Network-level failure on an idempotent request: worth one more go.
		return true
	}
	return ae.Status >= 500
}

func (c *Client) tracef(format string, args ...any) {
	if c.trace == nil {
		return
	}
	fmt.Fprintf(c.trace, "mautic: "+format+"\n", args...)
}
