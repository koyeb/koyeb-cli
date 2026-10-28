package koyeb

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	log "github.com/sirupsen/logrus"
)

// binaryTestPayload is a byte-exactness torture payload: every byte value
// once, then a run of invalid UTF-8 and NULs that the legacy JSON routes
// corrupted to U+FFFD.
func binaryTestPayload() []byte {
	b := make([]byte, 256)
	for i := range b {
		b[i] = byte(i)
	}
	return append(b, 0xC3, 0x28, 0x00, 0x80, 0xFF, 0xFE, 0x00, 0x00, 0xF0, 0x9F)
}

// captureLog redirects the global logrus output into a buffer for the
// duration of one test, returning the buffer and a read function.
func captureLog(t *testing.T) func() string {
	t.Helper()
	var buf bytes.Buffer
	log.SetOutput(&buf)
	t.Cleanup(func() { log.SetOutput(os.Stderr) })
	return func() string { return buf.String() }
}

// fakeFilesExecutor fakes a v0.0.24+ executor: the streamed /files routes
// over an in-memory store, counting each route's hits. Unknown paths 404
// like the real mux.
type fakeFilesExecutor struct {
	files      map[string][]byte
	probes     atomic.Int32 // GET /files/stat without a path (the probe)
	gets       atomic.Int32 // GET /files with a path
	legacyHits atomic.Int32 // any legacy JSON route (must stay 0 here)
}

func (f *fakeFilesExecutor) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch {
	case r.URL.Path == "/files/stat" && r.URL.Query().Get("path") == "":
		f.probes.Add(1)
		w.WriteHeader(http.StatusBadRequest)
		writeChecked(w, `{"error":"path is required"}`)
	case r.URL.Path == "/files" && r.Method == http.MethodGet:
		f.gets.Add(1)
		content, ok := f.files[r.URL.Query().Get("path")]
		if !ok {
			w.WriteHeader(http.StatusNotFound)
			writeChecked(w, `{"error":"file does not exist"}`)
			return
		}
		w.Header().Set("Content-Type", "application/octet-stream")
		writeChecked(w, string(content))
	default:
		f.legacyHits.Add(1)
		http.NotFound(w, r)
	}
}

// fakeLegacyExecutor fakes a pre-v0.0.24 executor: only the legacy JSON
// routes; every other path — /files/stat included — answers 404 like the
// real mux, which is exactly what the fallback probe keys on.
type fakeLegacyExecutor struct {
	probes   atomic.Int32 // any /files hit (the probe)
	reads    atomic.Int32 // POST /read_file
	stats    atomic.Int32 // POST /stat_file
	content  string       // /read_file response content
	b64      bool         // respond is_base64 with base64 content
	statBody string       // /stat_file response body
}

func (f *fakeLegacyExecutor) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch {
	case strings.HasPrefix(r.URL.Path, "/files"):
		f.probes.Add(1)
		http.NotFound(w, r)
	case r.URL.Path == "/read_file" && r.Method == http.MethodPost:
		f.reads.Add(1)
		if f.b64 {
			writeChecked(w, fmt.Sprintf(`{"content":%q,"is_base64":true}`, f.content))
		} else {
			writeChecked(w, fmt.Sprintf(`{"content":%q}`, f.content))
		}
	case r.URL.Path == "/stat_file" && r.Method == http.MethodPost:
		f.stats.Add(1)
		writeChecked(w, f.statBody)
	default:
		http.NotFound(w, r)
	}
}

func TestReadFileStreamsRawBytesOnFilesRoutes(t *testing.T) {
	content := binaryTestPayload()
	fake := &fakeFilesExecutor{files: map[string][]byte{"/tmp/data.bin": content}}
	client, _ := executorServer(t, "the-secret", "", fake.ServeHTTP, WithRetries(0, time.Millisecond))

	got, err := client.ReadFile(context.Background(), "/tmp/data.bin")
	require.NoError(t, err)
	assert.Equal(t, content, got, "every byte value and invalid UTF-8 must round-trip byte-exact")
	assert.Equal(t, int32(1), fake.probes.Load(), "the probe runs once")
	assert.Equal(t, int32(1), fake.gets.Load(), "the bytes come from one GET /files")
	assert.Equal(t, int32(0), fake.legacyHits.Load(), "the legacy JSON routes are never touched")
}

func TestReadFileFallsBackToLegacyExecutor(t *testing.T) {
	fake := &fakeLegacyExecutor{content: "plain text"}
	client, _ := executorServer(t, "the-secret", "", fake.ServeHTTP, WithRetries(0, time.Millisecond))

	got, err := client.ReadFile(context.Background(), "/tmp/a.txt")
	require.NoError(t, err)
	assert.Equal(t, []byte("plain text"), got)
	assert.Equal(t, int32(1), fake.probes.Load(), "the 404 probe pins the handle to legacy")
	assert.Equal(t, int32(1), fake.reads.Load(), "the content comes from the legacy /read_file")
}

func TestReadFileLegacyBase64StillDecoded(t *testing.T) {
	// The is_base64 negotiation stays on the legacy path: some executors
	// answer base64 there, and the fallback must keep parsing it.
	raw := binaryTestPayload()
	fake := &fakeLegacyExecutor{b64: true, content: base64.StdEncoding.EncodeToString(raw)}
	client, _ := executorServer(t, "the-secret", "", fake.ServeHTTP, WithRetries(0, time.Millisecond))

	got, err := client.ReadFile(context.Background(), "/tmp/data.bin")
	require.NoError(t, err)
	assert.Equal(t, raw, got)
}

func TestFilesRoutesProbeVerdictCachedPerClient(t *testing.T) {
	t.Run("new executor: one probe across reads", func(t *testing.T) {
		fake := &fakeFilesExecutor{files: map[string][]byte{
			"/tmp/a": []byte("a"),
			"/tmp/b": []byte("b"),
		}}
		client, _ := executorServer(t, "the-secret", "", fake.ServeHTTP, WithRetries(0, time.Millisecond))

		for _, p := range []string{"/tmp/a", "/tmp/b", "/tmp/a"} {
			_, err := client.ReadFile(context.Background(), p)
			require.NoError(t, err)
		}
		assert.Equal(t, int32(1), fake.probes.Load(), "the verdict is cached for the client's lifetime")
		assert.Equal(t, int32(3), fake.gets.Load())
	})

	t.Run("legacy executor: one probe across reads", func(t *testing.T) {
		fake := &fakeLegacyExecutor{content: "x"}
		client, _ := executorServer(t, "the-secret", "", fake.ServeHTTP, WithRetries(0, time.Millisecond))

		for range 3 {
			_, err := client.ReadFile(context.Background(), "/tmp/a")
			require.NoError(t, err)
		}
		assert.Equal(t, int32(1), fake.probes.Load())
		assert.Equal(t, int32(3), fake.reads.Load())
	})
}

func TestLegacyFallbackWarnsOncePerClient(t *testing.T) {
	fake := &fakeLegacyExecutor{content: "x"}
	client, _ := executorServer(t, "the-secret", "", fake.ServeHTTP, WithRetries(0, time.Millisecond))
	readLog := captureLog(t)

	for range 2 {
		_, err := client.ReadFile(context.Background(), "/tmp/a")
		require.NoError(t, err)
	}

	logs := readLog()
	assert.Equal(t, 1, strings.Count(logs, "pre-v0.0.24"), "exactly one legacy warning per handle, got: %s", logs)
	assert.Contains(t, logs, "legacy JSON routes")
}

func TestFilesRoutesVerdictNeverInferredFromFailedTransfer(t *testing.T) {
	// A 404 on GET /files (the file is missing) must neither re-run the
	// probe nor flip the handle: only the body-less probe decides.
	fake := &fakeFilesExecutor{files: map[string][]byte{}}
	client, _ := executorServer(t, "the-secret", "", fake.ServeHTTP, WithRetries(0, time.Millisecond))

	for range 2 {
		_, err := client.ReadFile(context.Background(), "/missing")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "file does not exist")
	}
	assert.Equal(t, int32(1), fake.probes.Load(), "a failed read never re-probes")
	assert.Equal(t, int32(2), fake.gets.Load(), "the handle stays on the new routes")
}

func TestFilesRoutesProbeFailureUncachedAndSurfaced(t *testing.T) {
	// A probe answer that says nothing about the routes (401 here) fails
	// the operation without caching a verdict: the next call tries again.
	var probes atomic.Int32
	handler := func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/files/stat" {
			probes.Add(1)
			w.WriteHeader(http.StatusUnauthorized)
			writeChecked(w, `{"error":"unauthorized"}`)
			return
		}
		http.NotFound(w, r)
	}
	client, _ := executorServer(t, "the-secret", "", handler, WithRetries(0, time.Millisecond))

	for range 2 {
		_, err := client.ReadFile(context.Background(), "/tmp/a")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "files")
	}
	assert.Equal(t, int32(2), probes.Load(), "an inconclusive probe is retried on the next call")
}

// signalingWriter closes first on its first Write, then keeps the bytes.
type signalingWriter struct {
	first chan struct{}
	buf   bytes.Buffer
	once  sync.Once
}

func (w *signalingWriter) Write(p []byte) (int, error) {
	w.once.Do(func() { close(w.first) })
	return w.buf.Write(p)
}

func TestDownloadFileStreamsRawBytesOnFilesRoutes(t *testing.T) {
	// 300 KiB of binary: spans io.Copy's 32 KiB buffer many times.
	content := make([]byte, 300*1024)
	for i := range content {
		content[i] = byte(i * 7)
	}
	fake := &fakeFilesExecutor{files: map[string][]byte{"/tmp/data.bin": content}}
	client, _ := executorServer(t, "the-secret", "", fake.ServeHTTP, WithRetries(0, time.Millisecond))

	var out bytes.Buffer
	n, err := client.DownloadFile(context.Background(), "/tmp/data.bin", &out)
	require.NoError(t, err)
	assert.Equal(t, int64(len(content)), n)
	assert.Equal(t, content, out.Bytes(), "multi-chunk binary must arrive byte-exact")
	assert.Equal(t, int32(1), fake.probes.Load())
	assert.Equal(t, int32(1), fake.gets.Load())
	assert.Equal(t, int32(0), fake.legacyHits.Load())
}

func TestDownloadFileDeliversBytesWhileExecutorMidBody(t *testing.T) {
	// A true stream: bytes reach the caller while the executor is still
	// writing the body. A client that buffers the response whole (an eager
	// debug body dump, a buffered legacy read) never delivers in time.
	firstChunk := bytes.Repeat([]byte{0xAB}, 64*1024)
	secondChunk := []byte("the tail")
	firstFlushed := make(chan struct{})
	release := make(chan struct{})

	handler := func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/files/stat" {
			w.WriteHeader(http.StatusBadRequest)
			writeChecked(w, `{"error":"path is required"}`)
			return
		}
		writeChecked(w, string(firstChunk))
		if f, ok := w.(http.Flusher); ok {
			f.Flush()
		}
		close(firstFlushed)
		<-release
		writeChecked(w, string(secondChunk))
	}
	client, _ := executorServer(t, "the-secret", "", handler, WithRetries(0, time.Millisecond))

	writer := &signalingWriter{first: make(chan struct{})}
	done := make(chan error, 1)
	go func() {
		_, err := client.DownloadFile(context.Background(), "/tmp/data.bin", writer)
		done <- err
	}()

	select {
	case <-writer.first:
	case <-time.After(5 * time.Second):
		close(release)
		<-done
		t.Fatal("bytes only arrived after the executor finished the body: the response was buffered whole")
	}
	close(release)
	require.NoError(t, <-done)
	want := append(append([]byte{}, firstChunk...), secondChunk...)
	assert.Equal(t, want, writer.buf.Bytes())
}

func TestDownloadFileFallsBackToLegacyExecutor(t *testing.T) {
	fake := &fakeLegacyExecutor{content: "legacy bytes"}
	client, _ := executorServer(t, "the-secret", "", fake.ServeHTTP, WithRetries(0, time.Millisecond))

	var out bytes.Buffer
	n, err := client.DownloadFile(context.Background(), "/tmp/a.txt", &out)
	require.NoError(t, err)
	assert.Equal(t, int64(len("legacy bytes")), n)
	assert.Equal(t, []byte("legacy bytes"), out.Bytes())
	assert.Equal(t, int32(1), fake.probes.Load())
	assert.Equal(t, int32(1), fake.reads.Load())
}

func TestDownloadFileFilesRoutesErrorMapping(t *testing.T) {
	fake := &fakeFilesExecutor{files: map[string][]byte{}}
	client, _ := executorServer(t, "the-secret", "", fake.ServeHTTP, WithRetries(0, time.Millisecond))

	var out bytes.Buffer
	_, err := client.DownloadFile(context.Background(), "/missing", &out)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "status 404")
	assert.Contains(t, err.Error(), "file does not exist")
}

func TestDownloadFileRetriesServerErrorsBeforeStreaming(t *testing.T) {
	// 5xx statuses are retried like request does; a mid-body failure is
	// not (no resume — the caller reruns).
	var gets atomic.Int32
	content := []byte("after retries")
	handler := func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/files/stat" {
			w.WriteHeader(http.StatusBadRequest)
			writeChecked(w, `{"error":"path is required"}`)
			return
		}
		if gets.Add(1) <= 2 {
			w.WriteHeader(http.StatusBadGateway)
			writeChecked(w, `{"error":"waking up"}`)
			return
		}
		w.Header().Set("Content-Type", "application/octet-stream")
		writeChecked(w, string(content))
	}
	client, _ := executorServer(t, "the-secret", "", handler, WithRetries(2, time.Millisecond))

	var out bytes.Buffer
	n, err := client.DownloadFile(context.Background(), "/tmp/data.bin", &out)
	require.NoError(t, err)
	assert.Equal(t, content, out.Bytes())
	assert.Equal(t, int64(len(content)), n)
	assert.Equal(t, int32(3), gets.Load(), "two 502s then the 200")
}

func TestFsDownloadWritesLocalFile(t *testing.T) {
	ctx := &CLIContext{Context: context.Background()}

	t.Run("streams the remote file to the local path", func(t *testing.T) {
		fake := &fakeSandboxClient{downloadContent: []byte("remote bytes")}
		local := filepath.Join(t.TempDir(), "out.bin")

		require.NoError(t, downloadFile(ctx.Context, fake, "/remote/file.bin", local))

		got, err := os.ReadFile(local)
		require.NoError(t, err)
		assert.Equal(t, []byte("remote bytes"), got)
		info, err := os.Stat(local)
		require.NoError(t, err)
		assert.Equal(t, fs.FileMode(0o644), info.Mode().Perm())
		assert.Equal(t, "/remote/file.bin", fake.downloadPath)
	})

	t.Run("client errors map to the download CLI error", func(t *testing.T) {
		fake := &fakeSandboxClient{downloadErr: fmt.Errorf("connection reset")}
		err := downloadFile(ctx.Context, fake, "/remote/file.bin", filepath.Join(t.TempDir(), "x"))
		require.Error(t, err)
		assert.Contains(t, err.Error(), "downloading file from sandbox")
		assert.Contains(t, err.Error(), "connection reset")
	})
}

func TestReadFileFilesRoutesErrorMapping(t *testing.T) {
	for _, tt := range []struct {
		name   string
		status int
		body   string
	}{
		{"missing file", http.StatusNotFound, `{"error":"file does not exist"}`},
		{"directory", http.StatusBadRequest, `{"error":"/tmp is not a regular file (it is a dir)"}`},
		{"permission", http.StatusForbidden, `{"error":"permission denied"}`},
	} {
		t.Run(tt.name, func(t *testing.T) {
			handler := func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path == "/files/stat" && r.URL.Query().Get("path") == "" {
					w.WriteHeader(http.StatusBadRequest)
					writeChecked(w, `{"error":"path is required"}`)
					return
				}
				w.WriteHeader(tt.status)
				writeChecked(w, tt.body)
			}
			client, _ := executorServer(t, "the-secret", "", handler, WithRetries(0, time.Millisecond))

			_, err := client.ReadFile(context.Background(), "/tmp/a")
			require.Error(t, err)
			assert.Contains(t, err.Error(), fmt.Sprintf("status %d", tt.status))
			assert.NotContains(t, err.Error(), `"error"`, "the JSON body is unwrapped, not echoed")
		})
	}
}
