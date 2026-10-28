package koyeb

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"

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

// DownloadFile streams a file's bytes into w with constant memory on
// executor v0.0.24+: one GET /files, the body copied straight to w. There
// is no resume (ST-101): an interrupted download is the caller's to rerun.
// Older executors fall back to the buffered legacy read.
func (c *SandboxClient) DownloadFile(ctx context.Context, path string, w io.Writer) (int64, error) {
	routed, err := c.hasFilesRoutes(ctx)
	if err != nil {
		return 0, err
	}
	if !routed {
		content, err := c.legacyReadFile(ctx, path)
		if err != nil {
			return 0, err
		}
		return io.Copy(w, bytes.NewReader(content))
	}

	resp, err := c.streamGet(ctx, filesRoutePath(path))
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return 0, filesRouteError(resp)
	}
	return io.Copy(w, resp.Body)
}

// streamGet issues the download GET on the streaming client, retrying 5xx
// statuses like request does. A read that fails mid-body is not retried:
// the bytes already written to the caller cannot be taken back, and there
// is no resume — the caller reruns the download.
func (c *SandboxClient) streamGet(ctx context.Context, path string) (*http.Response, error) {
	var lastErr error
	for attempt := 0; attempt <= c.maxRetries; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(c.retryDelay * time.Duration(attempt)):
			}
		}

		req, err := c.newStreamRequest(ctx, http.MethodGet, path)
		if err != nil {
			return nil, err
		}
		resp, err := c.streamClient.Do(req)
		if err != nil {
			lastErr = err
			continue
		}
		if resp.StatusCode >= 500 {
			lastErr = fmt.Errorf("server error: status %d", resp.StatusCode)
			resp.Body.Close()
			continue
		}
		return resp, nil
	}
	return nil, fmt.Errorf("request failed after %d retries: %w", c.maxRetries, lastErr)
}

// newStreamRequest builds a body-less request for the streaming client,
// carrying the same headers as doRequest.
func (c *SandboxClient) newStreamRequest(ctx context.Context, method, path string) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.secret)
	if c.routingKey != "" {
		req.Header.Set("X-Routing-Key", c.routingKey)
	}
	return req, nil
}

// streamDebugTransport is DebugTransport for streamed transfers: the
// request and the response headers are dumped the same way, but the
// response body is never drained — a streamed download's bytes belong to
// the caller, and buffering them would defeat the constant memory.
type streamDebugTransport struct {
	http.RoundTripper
}

func (t *streamDebugTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if reqData, dumpErr := httputil.DumpRequestOut(req, true); dumpErr == nil {
		var safeReqData string
		if !debugFull {
			safeReqData = authorizationHeaderRegexp.ReplaceAllString(string(reqData), "Authorization: <HIDDEN, add --debug-full to show the token>")
		} else {
			safeReqData = string(reqData)
		}
		log.Debug(fmt.Sprintf("========== HTTP request ==========\n%s\n========== end of request ==========\n", safeReqData))
	}

	resp, err := t.RoundTripper.RoundTrip(req)
	if err != nil {
		return nil, err
	}
	// Headers only: dumping the body would buffer the whole stream.
	if respData, dumpErr := httputil.DumpResponse(resp, false); dumpErr == nil {
		log.Debug(fmt.Sprintf("========== HTTP response ==========\n%s\n========== end of response ==========\n", respData))
	}
	return resp, nil
}

// FileStat is a file's metadata: the /files/stat answer on executor
// v0.0.24+ (type, size, octal permission bits, RFC 3339 UTC mtime with
// nanoseconds), with the legacy /stat_file formats mapped onto the same
// shape on older executors.
type FileStat struct {
	Type  string `json:"type"` // file, dir, symlink or other
	Size  int64  `json:"size"`
	Mode  string `json:"mode"`            // octal permission bits, e.g. "0644"
	MTime string `json:"mtime,omitempty"` // RFC 3339 UTC, nanoseconds
}

// StatFile returns a file's metadata: GET /files/stat on executor
// v0.0.24+; the legacy POST /stat_file on older ones.
func (c *SandboxClient) StatFile(ctx context.Context, path string) (*FileStat, error) {
	routed, err := c.hasFilesRoutes(ctx)
	if err != nil {
		return nil, err
	}
	if !routed {
		return c.legacyStatFile(ctx, path)
	}

	resp, err := c.request(ctx, http.MethodGet, filesStatPath(path), nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, filesRouteError(resp)
	}

	var stat FileStat
	if err := json.NewDecoder(resp.Body).Decode(&stat); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	return &stat, nil
}

// statFromDirEntry maps a legacy DirEntry onto FileStat.
func statFromDirEntry(e *DirEntry) *FileStat {
	return &FileStat{
		Type:  dirEntryType(e.IsDir),
		Size:  e.Size,
		Mode:  e.Mode,
		MTime: e.ModTime,
	}
}

// dirEntryType renders the legacy is_dir flag as a type name.
func dirEntryType(isDir bool) string {
	if isDir {
		return "dir"
	}
	return "file"
}

// legacyStatFile stats a file through the legacy JSON route: POST
// /stat_file, in any of the formats older deployments answered with.
func (c *SandboxClient) legacyStatFile(ctx context.Context, path string) (*FileStat, error) {
	resp, err := c.request(ctx, http.MethodPost, "/stat_file", &FileRequest{Path: path})
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Try nested entry format first
	var nestedResult struct {
		Entry *DirEntry `json:"entry"`
		Error string    `json:"error,omitempty"`
	}
	if err := json.Unmarshal(body, &nestedResult); err == nil && nestedResult.Entry != nil {
		return statFromDirEntry(nestedResult.Entry), nil
	}
	if nestedResult.Error != "" {
		return nil, fmt.Errorf("stat file failed: %s", nestedResult.Error)
	}

	// Try direct DirEntry format
	var entry DirEntry
	if err := json.Unmarshal(body, &entry); err == nil && entry.Name != "" {
		return statFromDirEntry(&entry), nil
	}

	// Try flat format with fields at top level. Its Error field is handled
	// above: any string "error" is caught by the nested check, which runs
	// unconditionally — a flat error check here could never fire.
	var flatResult struct {
		Name    string `json:"name"`
		IsDir   bool   `json:"is_dir"`
		Size    int64  `json:"size"`
		ModTime string `json:"mod_time"`
		Mode    string `json:"mode"`
	}
	if err := json.Unmarshal(body, &flatResult); err == nil {
		return &FileStat{
			Type:  dirEntryType(flatResult.IsDir),
			Size:  flatResult.Size,
			Mode:  flatResult.Mode,
			MTime: flatResult.ModTime,
		}, nil
	}

	return nil, fmt.Errorf("failed to decode response: %s", string(body))
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
