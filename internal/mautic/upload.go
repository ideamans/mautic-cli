package mautic

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
)

// Upload posts a file as multipart/form-data, which is how Mautic takes files
// and themes: "the file is sent via a regular POST files array, like a browser
// sends it during a file upload".
//
// Retries are deliberately not applied here. Re-sending a multipart body after
// a 5xx risks creating the file twice, and a duplicate asset is worse than a
// failed one the caller can retry knowingly.
func (c *Client) Upload(ctx context.Context, path, field, filePath string, extra map[string]string) (json.RawMessage, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("could not open %s: %w", filePath, err)
	}
	defer f.Close()

	info, err := f.Stat()
	if err != nil {
		return nil, fmt.Errorf("could not stat %s: %w", filePath, err)
	}
	if info.IsDir() {
		return nil, fmt.Errorf("%s is a directory, not a file", filePath)
	}

	endpoint := c.apiURL + path

	if c.dryRun {
		return json.Marshal(DryRunResult{
			DryRun:   true,
			Method:   http.MethodPost,
			URL:      endpoint,
			Instance: c.baseURL,
			Auth:     c.auth.Describe(),
			Body: json.RawMessage(fmt.Sprintf(`{"multipart":{"%s":%q,"bytes":%d}}`,
				field, filepath.Base(filePath), info.Size())),
		})
	}

	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)
	for k, v := range extra {
		if err := writer.WriteField(k, v); err != nil {
			return nil, err
		}
	}
	part, err := writer.CreateFormFile(field, filepath.Base(filePath))
	if err != nil {
		return nil, err
	}
	if _, err := io.Copy(part, f); err != nil {
		return nil, fmt.Errorf("could not read %s: %w", filePath, err)
	}
	if err := writer.Close(); err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", defaultUserAgent)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	if err := c.auth.Apply(req); err != nil {
		return nil, err
	}

	c.tracef("POST %s (multipart, %d bytes)", endpoint, info.Size())
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("POST %s: %w", path, err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if v := resp.Header.Get("Mautic-Version"); v != "" {
		c.version = v
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, parseError(resp.StatusCode, http.MethodPost, path, data)
	}
	if !json.Valid(data) {
		return json.RawMessage(`{}`), nil
	}
	return data, nil
}

// Download streams a binary response, which is how `GET /themes/<name>` returns
// a theme: an application/zip body on success, but a JSON error envelope on
// failure — so the content type has to be checked before writing anything.
func (c *Client) Download(ctx context.Context, path string, w io.Writer) (int64, error) {
	endpoint := c.apiURL + path
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return 0, err
	}
	req.Header.Set("User-Agent", defaultUserAgent)
	if err := c.auth.Apply(req); err != nil {
		return 0, err
	}

	c.tracef("GET %s (binary)", endpoint)
	resp, err := c.http.Do(req)
	if err != nil {
		return 0, fmt.Errorf("GET %s: %w", path, err)
	}
	defer resp.Body.Close()

	if v := resp.Header.Get("Mautic-Version"); v != "" {
		c.version = v
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		data, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return 0, parseError(resp.StatusCode, http.MethodGet, path, data)
	}
	return io.Copy(w, resp.Body)
}
