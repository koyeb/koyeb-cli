package koyeb

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// writeChecked writes the body, satisfying errcheck in test handlers.
func writeChecked(w http.ResponseWriter, body string) {
	_, _ = w.Write([]byte(body))
}

// executorServer spins up an httptest server standing in for the sandbox
// executor and returns a client pointed at it.
func executorServer(t *testing.T, secret, routingKey string, handler http.HandlerFunc, opts ...SandboxClientOption) (*SandboxClient, *httptest.Server) {
	t.Helper()
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)
	return newSandboxClientFromBaseURL(server.URL, secret, routingKey, opts...), server
}

func TestParseSSE(t *testing.T) {
	run := func(input string) []StreamEvent {
		var events []StreamEvent
		client := &SandboxClient{}
		err := client.parseSSE(strings.NewReader(input), func(e StreamEvent) error {
			events = append(events, e)
			return nil
		})
		require.NoError(t, err)
		return events
	}

	t.Run("single event", func(t *testing.T) {
		events := run("event: output\ndata: {\"stream\":\"stdout\"}\n\n")
		require.Len(t, events, 1)
		assert.Equal(t, "output", events[0].Event)
		assert.Equal(t, `{"stream":"stdout"}`, events[0].Data)
	})

	t.Run("multi-line data joins with newlines (SSE spec)", func(t *testing.T) {
		events := run("event: output\ndata: line one\ndata: line two\n\n")
		require.Len(t, events, 1)
		assert.Equal(t, "line one\nline two", events[0].Data)
	})

	t.Run("comments and unknown fields are ignored", func(t *testing.T) {
		events := run(": keep-alive\nevent: output\ndata: x\n\n")
		require.Len(t, events, 1)
		assert.Equal(t, "output", events[0].Event)
	})

	t.Run("event without trailing blank line still fires", func(t *testing.T) {
		events := run("event: complete\ndata: {}")
		require.Len(t, events, 1)
		assert.Equal(t, "complete", events[0].Event)
	})

	t.Run("handler errors abort the stream", func(t *testing.T) {
		client := &SandboxClient{}
		err := client.parseSSE(strings.NewReader("event: error\ndata: boom\n\n"), func(StreamEvent) error {
			return assert.AnError
		})
		require.ErrorIs(t, err, assert.AnError)
	})
}

func TestSandboxClientRunRequestShape(t *testing.T) {
	var captured struct {
		method  string
		path    string
		auth    string
		routing string
		content string
		body    map[string]any
	}

	client, _ := executorServer(t, "the-secret", "routing-42", func(w http.ResponseWriter, r *http.Request) {
		captured.method = r.Method
		captured.path = r.URL.Path
		captured.auth = r.Header.Get("Authorization")
		captured.routing = r.Header.Get("X-Routing-Key")
		captured.content = r.Header.Get("Content-Type")
		require.NoError(t, json.NewDecoder(r.Body).Decode(&captured.body))
		w.Header().Set("Content-Type", "application/json")
		writeChecked(w, `{"stdout":"hello\n","stderr":"","code":0}`)
	}, WithRetries(0, time.Millisecond))

	res, err := client.Run(context.Background(), &RunRequest{Cmd: "echo hello", Cwd: "/app", Env: map[string]string{"K": "V"}, Timeout: 30})
	require.NoError(t, err)
	assert.Equal(t, "hello\n", res.Stdout)
	assert.Equal(t, 0, res.Code)

	assert.Equal(t, http.MethodPost, captured.method)
	assert.Equal(t, "/run", captured.path)
	assert.Equal(t, "Bearer the-secret", captured.auth)
	assert.Equal(t, "routing-42", captured.routing)
	assert.Equal(t, "application/json", captured.content)
	assert.Equal(t, "echo hello", captured.body["cmd"])
	assert.Equal(t, "/app", captured.body["cwd"])
	assert.Equal(t, map[string]any{"K": "V"}, captured.body["env"])
	assert.Equal(t, float64(30), captured.body["timeout"])
}

func TestSandboxClientRetryOnServerErrors(t *testing.T) {
	t.Run("5xx responses are retried until success", func(t *testing.T) {
		var calls int32
		client, _ := executorServer(t, "secret", "", func(w http.ResponseWriter, r *http.Request) {
			if atomic.AddInt32(&calls, 1) <= 2 {
				w.WriteHeader(http.StatusServiceUnavailable)
				return
			}
			writeChecked(w, `{"stdout":"","code":0}`)
		}, WithRetries(3, time.Millisecond))

		res, err := client.Run(context.Background(), &RunRequest{Cmd: "true"})
		require.NoError(t, err)
		assert.Equal(t, 0, res.Code)
		assert.Equal(t, int32(3), atomic.LoadInt32(&calls))
	})

	t.Run("4xx responses fail immediately without retry", func(t *testing.T) {
		var calls int32
		client, _ := executorServer(t, "secret", "", func(w http.ResponseWriter, r *http.Request) {
			atomic.AddInt32(&calls, 1)
			w.WriteHeader(http.StatusNotFound)
		}, WithRetries(3, time.Millisecond))

		_, err := client.Run(context.Background(), &RunRequest{Cmd: "true"})
		require.Error(t, err, "404 must surface as an API error")
		assert.Equal(t, int32(1), atomic.LoadInt32(&calls))
	})

	t.Run("persistent 5xx exhausts retries", func(t *testing.T) {
		client, _ := executorServer(t, "secret", "", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}, WithRetries(2, time.Millisecond))

		_, err := client.Run(context.Background(), &RunRequest{Cmd: "true"})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "after 2 retries")
	})
}

func TestSandboxClientStatFileFormats(t *testing.T) {
	formats := map[string]string{
		"nested entry": `{"entry":{"name":"app.py","is_dir":false,"size":128,"mode":"0644"},"error":""}`,
		"direct":       `{"name":"app.py","is_dir":false,"size":128,"mode":"0644"}`,
		"flat":         `{"name":"app.py","is_dir":false,"size":128,"mode":"0644"}`,
	}
	for name, body := range formats {
		t.Run(name, func(t *testing.T) {
			client, _ := executorServer(t, "secret", "", func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "/stat_file", r.URL.Path)
				writeChecked(w, body)
			})

			entry, err := client.StatFile(context.Background(), "/app/app.py")
			require.NoError(t, err)
			assert.Equal(t, "app.py", entry.Name)
			assert.False(t, entry.IsDir)
			assert.Equal(t, int64(128), entry.Size)
			assert.Equal(t, "0644", entry.Mode)
		})
	}

	t.Run("stat errors surface", func(t *testing.T) {
		client, _ := executorServer(t, "secret", "", func(w http.ResponseWriter, r *http.Request) {
			writeChecked(w, `{"error":"NO_SUCH_FILE"}`)
		})

		_, err := client.StatFile(context.Background(), "/missing")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "NO_SUCH_FILE")
	})
}

func TestSandboxClientRunStreamingRoundTrip(t *testing.T) {
	client, _ := executorServer(t, "secret", "", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/run_streaming", r.URL.Path)
		w.Header().Set("Content-Type", "text/event-stream")
		flusher := w.(http.Flusher)
		writeChecked(w, "event: output\ndata: {\"stream\":\"stdout\",\"data\":\"chunk one\"}\n\n")
		flusher.Flush()
		writeChecked(w, "event: output\ndata: {\"stream\":\"stderr\",\"data\":\"chunk two\"}\n\n")
		flusher.Flush()
		writeChecked(w, "event: complete\ndata: {\"code\":0,\"error\":false}\n\n")
	})

	var outputs []string
	var streams []string
	code := -1
	err := client.RunStreaming(context.Background(), &RunRequest{Cmd: "build"}, func(stream, data string) {
		streams = append(streams, stream)
		outputs = append(outputs, data)
	}, func(exit int, hasError bool) {
		code = exit
	})
	require.NoError(t, err)
	assert.Equal(t, []string{"stdout", "stderr"}, streams)
	assert.Equal(t, []string{"chunk one", "chunk two"}, outputs)
	assert.Equal(t, 0, code)
}
