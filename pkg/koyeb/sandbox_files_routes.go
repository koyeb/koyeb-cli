package koyeb

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	log "github.com/sirupsen/logrus"
)

// The streamed /files routes of executor v0.0.24+, and the automatic
// fallback to the legacy JSON routes on older executors.
//
// Every operation below that can use the streamed routes first asks
// hasFilesRoutes: a body-less GET /files/stat, answered 404 by a
// pre-v0.0.24 executor (the route does not exist) and 400 ("path is
// required") by a new one. The verdict is cached for the client's
// lifetime — one probe and at most one warning per sandbox handle — and
// is never inferred from a failed transfer: an early 404 on a large PUT
// can surface as a 502 on Nomad, so only the probe decides.

// legacyFilesWarning is logged once per client when the probe pins the
// handle to the legacy JSON routes.
const legacyFilesWarning = "sandbox executor has no streamed /files routes (pre-v0.0.24): file transfers use the legacy JSON routes"

// filesRoutePath builds the query for GET /files.
func filesRoutePath(path string) string {
	return "/files?" + url.Values{"path": []string{path}}.Encode()
}

// filesStatPath builds the query for GET /files/stat.
func filesStatPath(path string) string {
	return "/files/stat?" + url.Values{"path": []string{path}}.Encode()
}

// filesRouteError turns a non-2xx /files response into an error, unwrapping
// the executor's {"error": "..."} body when it parses.
func filesRouteError(resp *http.Response) error {
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
	var parsed struct {
		Error string `json:"error"`
	}
	if err := json.Unmarshal(body, &parsed); err == nil && parsed.Error != "" {
		return fmt.Errorf("sandbox API error (status %d): %s", resp.StatusCode, parsed.Error)
	}
	return fmt.Errorf("sandbox API error (status %d): %s", resp.StatusCode, strings.TrimSpace(string(body)))
}

// hasFilesRoutes reports whether the executor serves the streamed /files
// routes, probing once per client: one CLI command builds one client, so
// one sandbox handle probes once. A probe answer that says nothing about
// the routes (network error, 401, ...) surfaces as an error and is not
// cached — the next call tries again.
func (c *SandboxClient) hasFilesRoutes(ctx context.Context) (bool, error) {
	c.filesMu.Lock()
	defer c.filesMu.Unlock()

	if c.filesProbed {
		return c.filesRoutes, nil
	}

	resp, err := c.request(ctx, http.MethodGet, "/files/stat", nil)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 4096))

	switch resp.StatusCode {
	case http.StatusBadRequest:
		// The new executor answering "path is required".
		c.filesRoutes, c.filesProbed = true, true
	case http.StatusNotFound:
		// A pre-v0.0.24 executor: the route does not exist.
		c.filesRoutes, c.filesProbed = false, true
		log.Warn(legacyFilesWarning)
	default:
		// Says nothing about the routes: fail without caching a verdict.
		return false, fmt.Errorf("failed probing the sandbox /files routes (status %d)", resp.StatusCode)
	}
	return c.filesRoutes, nil
}

// ReadFile reads a file's raw bytes: one GET /files on executor v0.0.24+,
// byte-exact with no base64 negotiation. Older executors fall back to the
// legacy JSON route, where non-UTF-8 bytes may not survive.
func (c *SandboxClient) ReadFile(ctx context.Context, path string) ([]byte, error) {
	routed, err := c.hasFilesRoutes(ctx)
	if err != nil {
		return nil, err
	}
	if !routed {
		return c.legacyReadFile(ctx, path)
	}

	resp, err := c.request(ctx, http.MethodGet, filesRoutePath(path), nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, filesRouteError(resp)
	}
	return io.ReadAll(resp.Body)
}

// legacyReadFile reads a file through the legacy JSON route: POST
// /read_file, with the is_base64 negotiation kept as it was.
func (c *SandboxClient) legacyReadFile(ctx context.Context, path string) ([]byte, error) {
	req := &FileRequest{Path: path}
	resp, err := c.request(ctx, http.MethodPost, "/read_file", req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result FileResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if result.Error != "" {
		return nil, fmt.Errorf("read file failed: %s", result.Error)
	}

	// Decode base64 content if the response indicates it
	if result.IsB64 {
		return base64.StdEncoding.DecodeString(result.Content)
	}

	return []byte(result.Content), nil
}
