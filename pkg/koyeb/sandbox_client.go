package koyeb

import (
	"bufio"
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"
)

// SandboxClientInterface defines the interface for sandbox operations
// This allows for easier testing and mocking
type SandboxClientInterface interface {
	Run(ctx context.Context, req *RunRequest) (*RunResponse, error)
	RunStreaming(ctx context.Context, req *RunRequest, onOutput func(stream, data string), onComplete func(code int, hasError bool)) error
	WriteFile(ctx context.Context, path string, content []byte) error
	ReadFile(ctx context.Context, path string) ([]byte, error)
	DeleteFile(ctx context.Context, path string) error
	MakeDir(ctx context.Context, path string) error
	DeleteDir(ctx context.Context, path string) error
	ListDir(ctx context.Context, path string) ([]DirEntry, error)
	BindPort(ctx context.Context, port string) (*PortResponse, error)
	UnbindPort(ctx context.Context) (*PortResponse, error)
	StartProcess(ctx context.Context, req *ProcessRequest) (*StartProcessResponse, error)
	ListProcesses(ctx context.Context) ([]ProcessInfo, error)
	KillProcess(ctx context.Context, processID string) error
	StreamProcessLogs(ctx context.Context, processID string, follow bool, onLog func(timestamp, stream, data string)) error
	Health(ctx context.Context) (*HealthResponse, error)
}

// Ensure SandboxClient implements the interface
var _ SandboxClientInterface = (*SandboxClient)(nil)

// SandboxClient provides HTTP client for sandbox API
type SandboxClient struct {
	baseURL       string
	secret        string
	httpClient    *http.Client
	maxRetries    int
	retryDelay    time.Duration
	streamTimeout time.Duration
}

// SandboxClientOption is a functional option for configuring SandboxClient
type SandboxClientOption func(*SandboxClient)

// WithTimeout sets the default HTTP client timeout
func WithTimeout(timeout time.Duration) SandboxClientOption {
	return func(c *SandboxClient) {
		c.httpClient.Timeout = timeout
	}
}

// WithRetries sets the retry configuration
func WithRetries(maxRetries int, retryDelay time.Duration) SandboxClientOption {
	return func(c *SandboxClient) {
		c.maxRetries = maxRetries
		c.retryDelay = retryDelay
	}
}

// WithStreamTimeout sets the timeout for streaming operations
func WithStreamTimeout(timeout time.Duration) SandboxClientOption {
	return func(c *SandboxClient) {
		c.streamTimeout = timeout
	}
}

// NewSandboxClient creates a new sandbox client
func NewSandboxClient(domain, secret string, opts ...SandboxClientOption) *SandboxClient {
	c := &SandboxClient{
		baseURL: fmt.Sprintf("https://%s/koyeb-sandbox", domain),
		secret:  secret,
		httpClient: &http.Client{
			Timeout: 2 * time.Minute,
		},
		maxRetries:    3,
		retryDelay:    time.Second,
		streamTimeout: 30 * time.Minute,
	}

	for _, opt := range opts {
		opt(c)
	}

	return c
}

// request makes an HTTP request to the sandbox API with retry logic
func (c *SandboxClient) request(ctx context.Context, method, path string, body interface{}) (*http.Response, error) {
	var lastErr error

	for attempt := 0; attempt <= c.maxRetries; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(c.retryDelay * time.Duration(attempt)):
			}
		}

		resp, err := c.doRequest(ctx, method, path, body)
		if err != nil {
			lastErr = err
			// Only retry on network errors or 5xx responses
			if resp != nil && resp.StatusCode < 500 {
				return resp, err
			}
			continue
		}

		// Don't retry on success or client errors (4xx)
		if resp.StatusCode < 500 {
			return resp, nil
		}

		// Retry on server errors (5xx)
		lastErr = fmt.Errorf("server error: status %d", resp.StatusCode)
		resp.Body.Close()
	}

	return nil, fmt.Errorf("request failed after %d retries: %w", c.maxRetries, lastErr)
}

func (c *SandboxClient) doRequest(ctx context.Context, method, path string, body interface{}) (*http.Response, error) {
	var bodyReader io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		bodyReader = bytes.NewReader(jsonBody)
	}

	url := c.baseURL + path
	req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.secret)
	req.Header.Set("Content-Type", "application/json")

	return c.httpClient.Do(req)
}

// requestWithTimeout makes a request with a specific timeout
func (c *SandboxClient) requestWithTimeout(ctx context.Context, timeout time.Duration, method, path string, body interface{}) (*http.Response, error) {
	timeoutCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	return c.request(timeoutCtx, method, path, body)
}

// RunRequest is the request body for /run endpoint
type RunRequest struct {
	Cmd     string            `json:"cmd"`
	Cwd     string            `json:"cwd,omitempty"`
	Env     map[string]string `json:"env,omitempty"`
	Timeout int               `json:"timeout,omitempty"` // Timeout in seconds
}

// RunResponse is the response from /run endpoint
type RunResponse struct {
	Stdout string `json:"stdout"`
	Stderr string `json:"stderr"`
	Code   int    `json:"code"`
	Error  string `json:"error,omitempty"`
}

// Run executes a command and returns the result
func (c *SandboxClient) Run(ctx context.Context, req *RunRequest) (*RunResponse, error) {
	// Use request timeout if specified, otherwise use default
	timeout := 2 * time.Minute
	if req.Timeout > 0 {
		timeout = time.Duration(req.Timeout) * time.Second
	}

	resp, err := c.requestWithTimeout(ctx, timeout, "POST", "/run", req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("sandbox API error (status %d): %s", resp.StatusCode, string(body))
	}

	var result RunResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

// StreamEvent represents a server-sent event
type StreamEvent struct {
	Event string
	Data  string
}

// StreamOutputEvent is the data for output events
type StreamOutputEvent struct {
	Stream string `json:"stream"` // "stdout" or "stderr"
	Data   string `json:"data"`
}

// StreamCompleteEvent is the data for complete events
type StreamCompleteEvent struct {
	Code  int  `json:"code"`
	Error bool `json:"error"`
}

// StreamResult holds the result from streaming execution
type StreamResult struct {
	ExitCode int
	HasError bool
	mu       sync.Mutex
}

// SetResult safely sets the result
func (r *StreamResult) SetResult(code int, hasError bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.ExitCode = code
	r.HasError = hasError
}

// GetResult safely gets the result
func (r *StreamResult) GetResult() (int, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.ExitCode, r.HasError
}

// RunStreaming executes a command with streaming output
func (c *SandboxClient) RunStreaming(ctx context.Context, req *RunRequest, onOutput func(stream, data string), onComplete func(code int, hasError bool)) error {
	// Use stream timeout for long-running commands
	streamCtx, cancel := context.WithTimeout(ctx, c.streamTimeout)
	defer cancel()

	resp, err := c.doRequest(streamCtx, "POST", "/run_streaming", req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("sandbox API error (status %d): %s", resp.StatusCode, string(body))
	}

	return c.parseSSE(resp.Body, func(event StreamEvent) error {
		switch event.Event {
		case "output":
			var output StreamOutputEvent
			if err := json.Unmarshal([]byte(event.Data), &output); err != nil {
				return fmt.Errorf("failed to parse output event: %w", err)
			}
			if onOutput != nil {
				onOutput(output.Stream, output.Data)
			}
		case "complete":
			var complete StreamCompleteEvent
			if err := json.Unmarshal([]byte(event.Data), &complete); err != nil {
				return fmt.Errorf("failed to parse complete event: %w", err)
			}
			if onComplete != nil {
				onComplete(complete.Code, complete.Error)
			}
		case "error":
			return fmt.Errorf("command error: %s", event.Data)
		}
		return nil
	})
}

// parseSSE parses server-sent events from a reader
// Properly handles multi-line data according to SSE spec
func (c *SandboxClient) parseSSE(r io.Reader, handler func(StreamEvent) error) error {
	scanner := bufio.NewScanner(r)
	var currentEvent StreamEvent
	var dataLines []string

	for scanner.Scan() {
		line := scanner.Text()

		if line == "" {
			// Empty line means end of event
			if currentEvent.Event != "" || len(dataLines) > 0 {
				// Concatenate data lines with newlines (SSE spec)
				currentEvent.Data = strings.Join(dataLines, "\n")
				if err := handler(currentEvent); err != nil {
					return err
				}
				currentEvent = StreamEvent{}
				dataLines = nil
			}
			continue
		}

		if strings.HasPrefix(line, ":") {
			// Comment line, ignore
			continue
		}

		if strings.HasPrefix(line, "event:") {
			currentEvent.Event = strings.TrimSpace(strings.TrimPrefix(line, "event:"))
		} else if strings.HasPrefix(line, "data:") {
			// Accumulate data lines instead of overwriting
			dataLines = append(dataLines, strings.TrimPrefix(line, "data:"))
		}
		// id: and retry: lines are part of the SSE spec but not used by this client
	}

	// Handle any remaining event
	if currentEvent.Event != "" || len(dataLines) > 0 {
		currentEvent.Data = strings.Join(dataLines, "\n")
		if err := handler(currentEvent); err != nil {
			return err
		}
	}

	return scanner.Err()
}

// FileRequest is the request body for file operations
type FileRequest struct {
	Path    string `json:"path"`
	Content string `json:"content,omitempty"` // Base64 encoded for binary safety
	IsB64   bool   `json:"is_base64,omitempty"`
}

// FileResponse is the response from file operations
type FileResponse struct {
	Success bool       `json:"success,omitempty"`
	Content string     `json:"content,omitempty"` // Base64 encoded if is_base64 is true
	IsB64   bool       `json:"is_base64,omitempty"`
	Entries []DirEntry `json:"entries,omitempty"`
	Error   string     `json:"error,omitempty"`
}

// DirEntry represents a directory entry with metadata
type DirEntry struct {
	Name    string `json:"name"`
	IsDir   bool   `json:"is_dir"`
	Size    int64  `json:"size,omitempty"`
	ModTime string `json:"mod_time,omitempty"`
	Mode    string `json:"mode,omitempty"`
}

// WriteFile writes content to a file
func (c *SandboxClient) WriteFile(ctx context.Context, path string, content []byte) error {
	req := &FileRequest{
		Path:    path,
		Content: string(content),
	}

	resp, err := c.request(ctx, "POST", "/write_file", req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var result FileResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	if !result.Success {
		return fmt.Errorf("write file failed: %s", result.Error)
	}

	return nil
}

// ReadFile reads content from a file (handles binary data via base64)
func (c *SandboxClient) ReadFile(ctx context.Context, path string) ([]byte, error) {
	req := &FileRequest{Path: path}
	resp, err := c.request(ctx, "POST", "/read_file", req)
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

// DeleteFile deletes a file
func (c *SandboxClient) DeleteFile(ctx context.Context, path string) error {
	resp, err := c.request(ctx, "POST", "/delete_file", &FileRequest{Path: path})
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var result FileResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	if !result.Success {
		return fmt.Errorf("delete file failed: %s", result.Error)
	}

	return nil
}

// MakeDir creates a directory
func (c *SandboxClient) MakeDir(ctx context.Context, path string) error {
	resp, err := c.request(ctx, "POST", "/make_dir", &FileRequest{Path: path})
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var result FileResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	if !result.Success {
		return fmt.Errorf("make directory failed: %s", result.Error)
	}

	return nil
}

// DeleteDir deletes a directory recursively
func (c *SandboxClient) DeleteDir(ctx context.Context, path string) error {
	resp, err := c.request(ctx, "POST", "/delete_dir", &FileRequest{Path: path})
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var result FileResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	if !result.Success {
		return fmt.Errorf("delete directory failed: %s", result.Error)
	}

	return nil
}

// ListDir lists directory contents with metadata
func (c *SandboxClient) ListDir(ctx context.Context, path string) ([]DirEntry, error) {
	resp, err := c.request(ctx, "POST", "/list_dir", &FileRequest{Path: path})
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Try FileResponse with DirEntry entries first
	var result FileResponse
	if err := json.Unmarshal(body, &result); err == nil && len(result.Entries) > 0 {
		if result.Error != "" {
			return nil, fmt.Errorf("list directory failed: %s", result.Error)
		}
		return result.Entries, nil
	}

	// Try response with string entries (just filenames)
	var stringResult struct {
		Entries []string `json:"entries"`
		Error   string   `json:"error,omitempty"`
	}
	if err := json.Unmarshal(body, &stringResult); err == nil {
		if stringResult.Error != "" {
			return nil, fmt.Errorf("list directory failed: %s", stringResult.Error)
		}
		// Convert string entries to DirEntry
		entries := make([]DirEntry, len(stringResult.Entries))
		for i, name := range stringResult.Entries {
			entries[i] = DirEntry{Name: name}
		}
		return entries, nil
	}

	// Try direct array of strings
	var names []string
	if err := json.Unmarshal(body, &names); err == nil {
		entries := make([]DirEntry, len(names))
		for i, name := range names {
			entries[i] = DirEntry{Name: name}
		}
		return entries, nil
	}

	return nil, fmt.Errorf("failed to decode response: %s", string(body))
}

// StatFile gets file/directory metadata
func (c *SandboxClient) StatFile(ctx context.Context, path string) (*DirEntry, error) {
	resp, err := c.request(ctx, "POST", "/stat_file", &FileRequest{Path: path})
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
		return nestedResult.Entry, nil
	}
	if nestedResult.Error != "" {
		return nil, fmt.Errorf("stat file failed: %s", nestedResult.Error)
	}

	// Try direct DirEntry format
	var entry DirEntry
	if err := json.Unmarshal(body, &entry); err == nil && entry.Name != "" {
		return &entry, nil
	}

	// Try flat format with fields at top level
	var flatResult struct {
		Name    string `json:"name"`
		IsDir   bool   `json:"is_dir"`
		Size    int64  `json:"size"`
		ModTime string `json:"mod_time"`
		Mode    string `json:"mode"`
		Error   string `json:"error,omitempty"`
	}
	if err := json.Unmarshal(body, &flatResult); err == nil {
		if flatResult.Error != "" {
			return nil, fmt.Errorf("stat file failed: %s", flatResult.Error)
		}
		return &DirEntry{
			Name:    flatResult.Name,
			IsDir:   flatResult.IsDir,
			Size:    flatResult.Size,
			ModTime: flatResult.ModTime,
			Mode:    flatResult.Mode,
		}, nil
	}

	return nil, fmt.Errorf("failed to decode response: %s", string(body))
}

// PortRequest is the request body for port operations
type PortRequest struct {
	Port string `json:"port,omitempty"`
}

// PortResponse is the response from port operations
type PortResponse struct {
	Success     bool   `json:"success"`
	Message     string `json:"message,omitempty"`
	Port        string `json:"port,omitempty"`
	ProxyPort   string `json:"proxy_port,omitempty"` // The actual port exposed externally
	CurrentPort string `json:"current_port,omitempty"`
	Error       string `json:"error,omitempty"`
}

// BindPort binds a port to the TCP proxy
func (c *SandboxClient) BindPort(ctx context.Context, port string) (*PortResponse, error) {
	resp, err := c.request(ctx, "POST", "/bind_port", &PortRequest{Port: port})
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result PortResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

// UnbindPort unbinds the current port from the TCP proxy
func (c *SandboxClient) UnbindPort(ctx context.Context) (*PortResponse, error) {
	resp, err := c.request(ctx, "POST", "/unbind_port", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result PortResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

// GetPortStatus returns the currently bound port
func (c *SandboxClient) GetPortStatus(ctx context.Context) (*PortResponse, error) {
	resp, err := c.request(ctx, "GET", "/port_status", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Try to decode as PortResponse struct first
	var result PortResponse
	if err := json.Unmarshal(body, &result); err == nil {
		return &result, nil
	}

	// API might return just a port number or null
	var portNum int
	if err := json.Unmarshal(body, &portNum); err == nil {
		return &PortResponse{
			Port:    fmt.Sprintf("%d", portNum),
			Success: true,
		}, nil
	}

	// Check for null (no port bound)
	if string(body) == "null" || string(body) == "" {
		return &PortResponse{Success: true}, nil
	}

	return nil, fmt.Errorf("failed to decode response: unexpected format: %s", string(body))
}

// ProcessRequest is the request body for starting a process
type ProcessRequest struct {
	Cmd string            `json:"cmd"`
	Cwd string            `json:"cwd,omitempty"`
	Env map[string]string `json:"env,omitempty"`
}

// ProcessInfo is the information about a process
type ProcessInfo struct {
	ID        string `json:"id"`
	PID       int    `json:"pid,omitempty"`
	Status    string `json:"status"`
	Command   string `json:"command"`
	StartTime string `json:"start_time,omitempty"`
}

// StartProcessResponse is the response from starting a process
type StartProcessResponse struct {
	ID     string `json:"id"`
	PID    int    `json:"pid"`
	Status string `json:"status"`
}

// StartProcess starts a background process
func (c *SandboxClient) StartProcess(ctx context.Context, req *ProcessRequest) (*StartProcessResponse, error) {
	resp, err := c.request(ctx, "POST", "/start_process", req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("start process failed (status %d): %s", resp.StatusCode, string(body))
	}

	var result StartProcessResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

// ListProcessesResponse is the response from listing processes
type ListProcessesResponse struct {
	Processes []ProcessInfo `json:"processes"`
}

// ListProcesses lists all background processes
func (c *SandboxClient) ListProcesses(ctx context.Context) ([]ProcessInfo, error) {
	resp, err := c.request(ctx, "GET", "/list_processes", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("list processes failed (status %d): %s", resp.StatusCode, string(body))
	}

	var result ListProcessesResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return result.Processes, nil
}

// KillProcessRequest is the request body for killing a process
type KillProcessRequest struct {
	ID     string `json:"id"`
	Signal string `json:"signal,omitempty"` // Optional signal (default: SIGTERM)
}

// KillProcessResponse is the response from killing a process
type KillProcessResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
	Error   string `json:"error,omitempty"`
}

// KillProcess kills a background process
func (c *SandboxClient) KillProcess(ctx context.Context, processID string) error {
	resp, err := c.request(ctx, "POST", "/kill_process", &KillProcessRequest{ID: processID})
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var result KillProcessResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	if !result.Success {
		return fmt.Errorf("kill process failed: %s", result.Error)
	}

	return nil
}

// ProcessLogEvent is the data for log events
type ProcessLogEvent struct {
	Timestamp string `json:"timestamp"`
	Stream    string `json:"stream"`
	Data      string `json:"data"`
}

// StreamProcessLogs streams logs from a background process
func (c *SandboxClient) StreamProcessLogs(ctx context.Context, processID string, follow bool, onLog func(timestamp, stream, data string)) error {
	path := fmt.Sprintf("/process_logs_streaming?id=%s", processID)
	if follow {
		path += "&follow=true"
	}

	// Use stream timeout for long-running log streams
	streamCtx, cancel := context.WithTimeout(ctx, c.streamTimeout)
	defer cancel()

	resp, err := c.doRequest(streamCtx, "GET", path, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("stream logs failed (status %d): %s", resp.StatusCode, string(body))
	}

	return c.parseSSE(resp.Body, func(event StreamEvent) error {
		switch event.Event {
		case "log":
			var log ProcessLogEvent
			if err := json.Unmarshal([]byte(event.Data), &log); err != nil {
				return fmt.Errorf("failed to parse log event: %w", err)
			}
			if onLog != nil {
				onLog(log.Timestamp, log.Stream, log.Data)
			}
		case "complete":
			// Stream ended
			return nil
		case "error":
			return fmt.Errorf("log stream error: %s", event.Data)
		}
		return nil
	})
}

// HealthResponse contains sandbox health information
type HealthResponse struct {
	Healthy   bool   `json:"healthy"`
	Status    string `json:"status"`
	Uptime    int64  `json:"uptime,omitempty"`
	Version   string `json:"version,omitempty"`
	ProxyPort string `json:"proxy_port,omitempty"` // The TCP proxy port
}

// Health checks if the sandbox is healthy and returns detailed info
func (c *SandboxClient) Health(ctx context.Context) (*HealthResponse, error) {
	resp, err := c.request(ctx, "GET", "/health", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return &HealthResponse{Healthy: false, Status: fmt.Sprintf("unhealthy (status %d)", resp.StatusCode)}, nil
	}

	var result HealthResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		// If we can't decode, but got 200, assume healthy
		return &HealthResponse{Healthy: true, Status: "healthy"}, nil
	}

	result.Healthy = true
	return &result, nil
}
