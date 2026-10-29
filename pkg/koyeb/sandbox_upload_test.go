package koyeb

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	log "github.com/sirupsen/logrus"
)

// The multipart upload session flow, mirrored from the Python SDK's
// upload_file (koyeb/koyeb-python-sdk#106) against httptest fakes.

// uploadCall is one wire call the fake upload executor saw.
type uploadCall struct {
	kind   string // probe, files_put, open, part, complete or abort
	path   string // open: the destination; files_put: the query's path
	size   int64  // open: the declared size
	offset int64  // part: the offset query parameter
	body   []byte // part/files_put: the bytes received
	sha    string // part/files_put: the X-Content-SHA256 header
}

// fakeSpan is a half-open [start, end) range of bytes received.
type fakeSpan struct {
	start, end int64
}

// fakeUploadSession is the executor's in-memory view of one open session.
type fakeUploadSession struct {
	path  string
	size  int64
	want  string // the whole-file digest promised at open
	data  []byte
	spans []fakeSpan
}

// fakeUploadExecutor fakes a v0.0.24+ executor: the /files routes plus the
// multipart session routes over an in-memory store. Parts land at offsets
// into a pre-sized buffer, complete answers 409 with the true missing
// ranges until coverage holds, then verifies the whole-file digest the way
// the executor does. failParts/forgetOffsets/neverRecord script the faults
// the flow must recover from; openReply serves a scripted open answer.
type fakeUploadExecutor struct {
	maxPartSize     int64
	openStatus      int    // when set, open answers this status verbatim
	openReply       string // raw JSON served for open, overriding the live session
	lieCompleteSHA  string // complete echoes this digest instead of the true one
	lieCompleteSize int64  // complete echoes this size when non-zero
	omitCompleteSHA bool
	raw409Body      string        // served verbatim for the 409, overriding the ranges
	completeStatus  int           // when set, complete answers this status verbatim
	partDelay       time.Duration // slept per part before processing
	abortFails      bool          // abort answers 404 like an unknown session
	failParts       map[[2]int64]int
	forgetOffsets   map[int64]bool
	neverRecord     map[int64]bool
	onPart          func() error

	liePutSHA string            // files_put echoes this digest instead of the true one
	putLen    int64             // the Content-Length of the last PUT /files
	files     map[string][]byte // PUT /files store
	session   *fakeUploadSession

	probes      atomic.Int32
	inFlight    atomic.Int32
	maxInFlight atomic.Int32

	mu    sync.Mutex
	calls []uploadCall
}

func (f *fakeUploadExecutor) record(c uploadCall) {
	f.mu.Lock()
	f.calls = append(f.calls, c)
	f.mu.Unlock()
}

func (f *fakeUploadExecutor) kinds() []string {
	f.mu.Lock()
	defer f.mu.Unlock()
	kinds := make([]string, len(f.calls))
	for i, c := range f.calls {
		kinds[i] = c.kind
	}
	return kinds
}

// partPuts is every part PUT as (offset, length), in call order.
func (f *fakeUploadExecutor) partPuts() [][2]int64 {
	f.mu.Lock()
	defer f.mu.Unlock()
	var puts [][2]int64
	for _, c := range f.calls {
		if c.kind == "part" {
			puts = append(puts, [2]int64{c.offset, int64(len(c.body))})
		}
	}
	return puts
}

func (f *fakeUploadExecutor) countKinds(kind string) int {
	f.mu.Lock()
	defer f.mu.Unlock()
	n := 0
	for _, c := range f.calls {
		if c.kind == kind {
			n++
		}
	}
	return n
}

// landed is the assembled file: the session's buffer when one is live,
// the single-PUT store otherwise.
func (f *fakeUploadExecutor) landed() []byte {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.session != nil {
		return f.session.data
	}
	for _, b := range f.files {
		return b
	}
	return nil
}

// addSpan merges [start, end) into the sorted, coalesced spans — the
// executor's addSpan.
func (s *fakeUploadSession) addSpan(start, end int64) {
	ns, ne := start, end
	out := make([]fakeSpan, 0, len(s.spans)+1)
	for _, sp := range s.spans {
		if sp.end >= ns && ne >= sp.start {
			ns, ne = min(ns, sp.start), max(ne, sp.end)
			continue
		}
		out = append(out, sp)
	}
	out = append(out, fakeSpan{ns, ne})
	for i := 1; i < len(out); i++ {
		if out[i].start < out[i-1].start {
			out[i], out[i-1] = out[i-1], out[i]
		}
	}
	s.spans = out
}

// missing is the part of [0, size) no span covers — the executor's
// missingSpans.
func (s *fakeUploadSession) missing() []fakeSpan {
	var missing []fakeSpan
	at := int64(0)
	for _, sp := range s.spans {
		if sp.start > at {
			missing = append(missing, fakeSpan{at, sp.start})
		}
		at = max(at, sp.end)
	}
	if at < s.size {
		missing = append(missing, fakeSpan{at, s.size})
	}
	return missing
}

func writeJSONChecked(w http.ResponseWriter, status int, body string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, _ = w.Write([]byte(body))
}

func (f *fakeUploadExecutor) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch {
	case r.URL.Path == "/files/stat" && r.URL.Query().Get("path") == "":
		f.probes.Add(1)
		writeJSONChecked(w, http.StatusBadRequest, `{"error":"path is required"}`)

	case r.URL.Path == "/files" && r.Method == http.MethodPut:
		body, _ := io.ReadAll(r.Body)
		sum := sha256.Sum256(body)
		f.record(uploadCall{
			kind: "files_put", path: r.URL.Query().Get("path"),
			body: body, sha: r.Header.Get("X-Content-SHA256"),
		})
		f.putLen = r.ContentLength
		if f.files == nil {
			f.files = map[string][]byte{}
		}
		f.files[r.URL.Query().Get("path")] = body
		replySHA := hex.EncodeToString(sum[:])
		if f.liePutSHA != "" {
			replySHA = f.liePutSHA
		}
		writeJSONChecked(w, http.StatusOK, fmt.Sprintf(`{"size":%d,"sha256":%q}`, len(body), replySHA))

	case r.URL.Path == "/files/uploads" && r.Method == http.MethodPost:
		var req struct {
			Path   string `json:"path"`
			Size   int64  `json:"size"`
			SHA256 string `json:"sha256"`
		}
		body, _ := io.ReadAll(r.Body)
		if err := json.Unmarshal(body, &req); err != nil {
			writeJSONChecked(w, http.StatusBadRequest, `{"error":"invalid request body"}`)
			return
		}
		f.record(uploadCall{kind: "open", path: req.Path, size: req.Size, sha: req.SHA256})
		if f.openStatus != 0 {
			writeJSONChecked(w, f.openStatus, fmt.Sprintf(`{"error":"scripted open failure: status %d"}`, f.openStatus))
			return
		}
		if f.openReply != "" {
			writeJSONChecked(w, http.StatusOK, f.openReply)
			return
		}
		f.mu.Lock()
		f.session = &fakeUploadSession{
			path: req.Path, size: req.Size, want: req.SHA256,
			data: make([]byte, req.Size),
		}
		f.mu.Unlock()
		writeJSONChecked(w, http.StatusOK,
			fmt.Sprintf(`{"upload_id":"u-1","max_part_size":%d}`, f.maxPartSize))

	case strings.HasPrefix(r.URL.Path, "/files/uploads/") && r.Method == http.MethodPut:
		f.partHandler(w, r)

	case strings.HasPrefix(r.URL.Path, "/files/uploads/") && r.Method == http.MethodPost:
		f.completeHandler(w, r)

	case strings.HasPrefix(r.URL.Path, "/files/uploads/") && r.Method == http.MethodDelete:
		f.abortHandler(w, r)

	default:
		http.NotFound(w, r)
	}
}

// partHandler streams one part to its offset, enforcing the executor's
// contract: a live session, a matching part digest, in-bounds writes and
// the part cap. A scripted fault fires before any of it, once per part.
func (f *fakeUploadExecutor) partHandler(w http.ResponseWriter, r *http.Request) {
	f.inFlight.Add(1)
	if m := f.inFlight.Load(); m > f.maxInFlight.Load() {
		f.maxInFlight.Store(m)
	}
	defer f.inFlight.Add(-1)

	if f.partDelay > 0 {
		time.Sleep(f.partDelay)
	}
	if f.onPart != nil {
		if err := f.onPart(); err != nil {
			writeJSONChecked(w, http.StatusInternalServerError, `{"error":"part hook refused the part"}`)
			return
		}
	}

	offset := int64(0)
	_, err := fmt.Sscanf(r.URL.Query().Get("offset"), "%d", &offset)
	if err != nil {
		writeJSONChecked(w, http.StatusBadRequest, `{"error":"offset must be a non-negative integer"}`)
		return
	}
	body, _ := io.ReadAll(r.Body)
	f.record(uploadCall{kind: "part", offset: offset, body: body, sha: r.Header.Get("X-Content-SHA256")})

	if status, ok := f.failParts[[2]int64{offset, int64(len(body))}]; ok {
		delete(f.failParts, [2]int64{offset, int64(len(body))})
		writeJSONChecked(w, status, fmt.Sprintf(`{"error":"scripted failure: status %d"}`, status))
		return
	}

	sum := sha256.Sum256(body)
	got := hex.EncodeToString(sum[:])

	f.mu.Lock()
	defer f.mu.Unlock()
	sess := f.session
	if sess == nil {
		writeJSONChecked(w, http.StatusNotFound, `{"error":"unknown upload session"}`)
		return
	}
	if want := strings.ToLower(r.Header.Get("X-Content-SHA256")); want != "" && want != got {
		writeJSONChecked(w, http.StatusBadRequest, fmt.Sprintf(`{"error":"sha256 mismatch: got %s, expected %s"}`, got, want))
		return
	}
	if offset+int64(len(body)) > sess.size {
		writeJSONChecked(w, http.StatusBadRequest, `{"error":"part writes past the end of the upload"}`)
		return
	}
	if int64(len(body)) > f.maxPartSize {
		writeJSONChecked(w, http.StatusRequestEntityTooLarge, `{"error":"part is larger than max_part_size"}`)
		return
	}
	copy(sess.data[offset:], body)
	switch {
	case f.neverRecord[offset]:
		// The part lands but its span is never recorded.
	case f.forgetOffsets[offset]:
		delete(f.forgetOffsets, offset)
	default:
		sess.addSpan(offset, offset+int64(len(body)))
	}
	w.WriteHeader(http.StatusNoContent)
}

func (f *fakeUploadExecutor) completeHandler(w http.ResponseWriter, r *http.Request) {
	f.record(uploadCall{kind: "complete"})

	if f.completeStatus != 0 {
		writeJSONChecked(w, f.completeStatus, fmt.Sprintf(`{"error":"scripted complete failure: status %d"}`, f.completeStatus))
		return
	}

	f.mu.Lock()
	defer f.mu.Unlock()
	sess := f.session
	if sess == nil {
		writeJSONChecked(w, http.StatusNotFound, `{"error":"unknown upload session"}`)
		return
	}
	if missing := sess.missing(); len(missing) > 0 {
		if f.raw409Body != "" {
			writeJSONChecked(w, http.StatusConflict, f.raw409Body)
			return
		}
		ranges := make([]string, len(missing))
		for i, sp := range missing {
			ranges[i] = fmt.Sprintf(`{"start":%d,"end":%d}`, sp.start, sp.end)
		}
		writeJSONChecked(w, http.StatusConflict,
			fmt.Sprintf(`{"error":"%d range(s) of the upload are still missing","ranges":[%s]}`,
				len(missing), strings.Join(ranges, ",")))
		return
	}
	sum := sha256.Sum256(sess.data)
	digest := hex.EncodeToString(sum[:])
	if sess.want != "" && sess.want != digest {
		f.session = nil // the executor sweeps a corrupted session
		writeJSONChecked(w, http.StatusBadRequest,
			fmt.Sprintf(`{"error":"sha256 mismatch: got %s, expected %s"}`, digest, sess.want))
		return
	}
	if f.lieCompleteSHA != "" {
		digest = f.lieCompleteSHA
	}
	if f.omitCompleteSHA {
		writeJSONChecked(w, http.StatusOK, fmt.Sprintf(`{"size":%d}`, sess.size))
		return
	}
	size := sess.size
	if f.lieCompleteSize != 0 {
		size = f.lieCompleteSize
	}
	writeJSONChecked(w, http.StatusOK, fmt.Sprintf(`{"size":%d,"sha256":%q}`, size, digest))
}

func (f *fakeUploadExecutor) abortHandler(w http.ResponseWriter, r *http.Request) {
	f.record(uploadCall{kind: "abort"})
	if f.abortFails {
		writeJSONChecked(w, http.StatusNotFound, `{"error":"unknown upload session"}`)
		return
	}
	f.mu.Lock()
	f.session = nil
	f.mu.Unlock()
	w.WriteHeader(http.StatusNoContent)
}

// uploadTestClient points a client at the fake executor.
func uploadTestClient(t *testing.T, fake *fakeUploadExecutor, opts ...SandboxClientOption) *SandboxClient {
	t.Helper()
	client, _ := executorServer(t, "the-secret", "", fake.ServeHTTP, opts...)
	return client
}

// uploadTestData is the 12-byte payload the Python tests use.
func uploadTestData() []byte {
	return []byte("abcdefghijkl")
}

// bigTestFile writes n bytes of a deterministic pattern to dir.
func bigTestFile(t *testing.T, dir string, n int64) string {
	t.Helper()
	path := filepath.Join(dir, "big.bin")
	f, err := os.Create(path)
	require.NoError(t, err)
	defer f.Close()
	buf := make([]byte, 1<<20)
	for i := range buf {
		buf[i] = byte(i * 7 / 3)
	}
	for remaining := n; remaining > 0; {
		chunk := buf
		if int64(len(chunk)) > remaining {
			chunk = chunk[:remaining]
		}
		_, err := f.Write(chunk)
		require.NoError(t, err)
		remaining -= int64(len(chunk))
	}
	return path
}

func TestUploadFileSingleShot(t *testing.T) {
	// A file at or below the part size takes the single streamed PUT.
	content := binaryTestPayload()
	sum := sha256.Sum256(content)

	t.Run("streams the whole file through one PUT /files", func(t *testing.T) {
		dir := t.TempDir()
		local := filepath.Join(dir, "small.bin")
		require.NoError(t, os.WriteFile(local, content, 0o644))

		fake := &fakeUploadExecutor{maxPartSize: 1 << 30}
		client := uploadTestClient(t, fake, WithRetries(0, time.Millisecond))

		written, err := client.UploadFile(context.Background(), local, "/r/small.bin")
		require.NoError(t, err)
		assert.Equal(t, int64(len(content)), written)

		f := func() uploadCall {
			fake.mu.Lock()
			defer fake.mu.Unlock()
			return fake.calls[0]
		}()
		assert.Equal(t, "files_put", f.kind, "no session is opened")
		assert.Equal(t, 0, fake.countKinds("open"))
		assert.Equal(t, "/r/small.bin", f.path)
		assert.Equal(t, content, f.body, "every byte arrives byte-exact")
		assert.Equal(t, hex.EncodeToString(sum[:]), strings.ToLower(f.sha),
			"the whole-file digest rides X-Content-SHA256")
		assert.Equal(t, int64(len(content)), fake.putLen,
			"the body is framed with an exact Content-Length, not chunked")
		assert.Equal(t, int32(1), fake.probes.Load())
	})

	t.Run("a zero-byte file takes the single PUT with an empty body", func(t *testing.T) {
		dir := t.TempDir()
		local := filepath.Join(dir, "empty.bin")
		require.NoError(t, os.WriteFile(local, nil, 0o644))

		fake := &fakeUploadExecutor{maxPartSize: 1 << 30}
		client := uploadTestClient(t, fake, WithRetries(0, time.Millisecond))

		written, err := client.UploadFile(context.Background(), local, "/r/empty.bin")
		require.NoError(t, err)
		assert.Equal(t, int64(0), written)
		assert.Equal(t, []string{"files_put"}, func() []string {
			fake.mu.Lock()
			defer fake.mu.Unlock()
			kinds := make([]string, 0, len(fake.calls))
			for _, c := range fake.calls {
				if c.kind != "probe" {
					kinds = append(kinds, c.kind)
				}
			}
			return kinds
		}())
	})

	t.Run("a lying PUT answer raises corrupted", func(t *testing.T) {
		dir := t.TempDir()
		local := filepath.Join(dir, "small.bin")
		require.NoError(t, os.WriteFile(local, content, 0o644))

		fake := &fakeUploadExecutor{maxPartSize: 1 << 30, liePutSHA: strings.Repeat("ee", 32)}
		client := uploadTestClient(t, fake, WithRetries(0, time.Millisecond))

		_, err := client.UploadFile(context.Background(), local, "/r/small.bin")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "corrupted in transit")
	})
}

func TestUploadFileRetriesServerErrorsWithTheBodyReplayed(t *testing.T) {
	// Every attempt re-reads the file from disk: 5xx statuses retry like
	// streamPut does, with the full body replayed per attempt.
	content := binaryTestPayload()
	wantSum := sha256.Sum256(content)
	wantSHA := hex.EncodeToString(wantSum[:])
	dir := t.TempDir()
	local := filepath.Join(dir, "small.bin")
	require.NoError(t, os.WriteFile(local, content, 0o644))

	var mu sync.Mutex
	var attempts []string
	var puts atomic.Int32
	client, _ := executorServer(t, "the-secret", "", func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/files/stat" && r.URL.Query().Get("path") == "":
			w.WriteHeader(http.StatusBadRequest)
			writeChecked(w, `{"error":"path is required"}`)
		case r.URL.Path == "/files" && r.Method == http.MethodPut:
			body, _ := io.ReadAll(r.Body)
			got := fmt.Sprintf("%d:%s", len(body), r.Header.Get("X-Content-SHA256"))
			mu.Lock()
			attempts = append(attempts, got)
			mu.Unlock()
			if puts.Add(1) <= 2 {
				w.WriteHeader(http.StatusBadGateway)
				writeChecked(w, `{"error":"waking up"}`)
				return
			}
			writeChecked(w, fmt.Sprintf(`{"size":%d,"sha256":%q}`, len(body), wantSHA))
		default:
			http.NotFound(w, r)
		}
	}, WithRetries(2, time.Millisecond))

	_, uploadErr := client.UploadFile(context.Background(), local, "/r/a.bin")
	require.NoError(t, uploadErr)

	mu.Lock()
	defer mu.Unlock()
	require.Len(t, attempts, 3, "two 502s then the 200")
	for i, a := range attempts {
		assert.Equal(t, fmt.Sprintf("%d:%s", len(content), wantSHA), a,
			"attempt %d must replay the whole body with its digest", i+1)
	}
}

func TestUploadFileDispatchAtThePartSizeConstant(t *testing.T) {
	t.Run("exactly the part size takes the single PUT", func(t *testing.T) {
		local := bigTestFile(t, t.TempDir(), uploadPartSize)
		fake := &fakeUploadExecutor{maxPartSize: 1 << 30}
		client := uploadTestClient(t, fake, WithRetries(0, time.Millisecond))

		_, err := client.UploadFile(context.Background(), local, "/r/big.bin")
		require.NoError(t, err)
		assert.Equal(t, 1, fake.countKinds("files_put"))
		assert.Equal(t, 0, fake.countKinds("open"), "no session at exactly the part size")
	})

	t.Run("one byte past it opens a multipart session", func(t *testing.T) {
		local := bigTestFile(t, t.TempDir(), uploadPartSize+1)
		fake := &fakeUploadExecutor{maxPartSize: 1 << 30}
		client := uploadTestClient(t, fake, WithRetries(0, time.Millisecond))

		written, err := client.UploadFile(context.Background(), local, "/r/big.bin")
		require.NoError(t, err)
		assert.Equal(t, uploadPartSize+1, written)
		assert.Equal(t, 0, fake.countKinds("files_put"))
		assert.Equal(t, 1, fake.countKinds("open"))
		assert.Equal(t, 1, fake.countKinds("complete"))
		puts := fake.partPuts()
		sort.Slice(puts, func(i, j int) bool { return puts[i][0] < puts[j][0] })
		assert.Equal(t, [][2]int64{{0, uploadPartSize}, {uploadPartSize, 1}}, puts,
			"two parts: the full part and the one-byte tail, any order")

		want, err := os.ReadFile(local)
		require.NoError(t, err)
		assert.Equal(t, want, fake.landed())
	})
}

func TestUploadFileLegacyFallback(t *testing.T) {
	t.Run("a small file keeps the legacy JSON upload", func(t *testing.T) {
		dir := t.TempDir()
		local := filepath.Join(dir, "a.txt")
		require.NoError(t, os.WriteFile(local, []byte("legacy bytes"), 0o644))

		fake := &fakeLegacyExecutor{}
		client, _ := executorServer(t, "the-secret", "", fake.ServeHTTP, WithRetries(0, time.Millisecond))

		written, err := client.UploadFile(context.Background(), local, "/r/a.txt")
		require.NoError(t, err)
		assert.Equal(t, int64(len("legacy bytes")), written)
		assert.Equal(t, int32(1), fake.probes.Load(), "the probe pins the handle to legacy")
		assert.Equal(t, int32(1), fake.writes.Load(), "the upload goes to the legacy POST /write_file")
		body, _ := fake.writeBody.Load().(string)
		assert.Contains(t, body, `"path":"/r/a.txt"`)
		assert.Contains(t, body, `"content":"legacy bytes"`)
	})

	t.Run("files above the proxy cap are refused up front", func(t *testing.T) {
		// The 1G client cap is dropped for files-routes executors; the
		// legacy fallback has no multipart, so its single-request
		// uploads are refused at the 100 MiB proxy cap with a message
		// naming the cause — instead of buffering and dying at the
		// proxy with an opaque error.
		dir := t.TempDir()
		local := filepath.Join(dir, "huge.bin")
		f, err := os.Create(local)
		require.NoError(t, err)
		require.NoError(t, f.Truncate(legacyUploadLimit+1)) // sparse, no bytes buffered
		require.NoError(t, f.Close())

		fake := &fakeLegacyExecutor{}
		client, _ := executorServer(t, "the-secret", "", fake.ServeHTTP, WithRetries(0, time.Millisecond))

		_, err = client.UploadFile(context.Background(), local, "/r/huge.bin")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "100 MiB")
		assert.Contains(t, err.Error(), "v0.0.24")
		assert.Contains(t, err.Error(), "multipart")
		assert.Equal(t, int32(1), fake.probes.Load())
		assert.Equal(t, int32(0), fake.writes.Load(), "nothing is uploaded")
	})
}

func TestUploadFileMissingLocalFile(t *testing.T) {
	fake := &fakeUploadExecutor{maxPartSize: 1 << 30}
	client := uploadTestClient(t, fake, WithRetries(0, time.Millisecond))

	_, err := client.UploadFile(context.Background(), "/no/such/file.bin", "/r/x")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to open local file")
}

func TestUploadErrorMapping(t *testing.T) {
	data := uploadTestData()
	settings := uploadSettings{partSize: 4, concurrency: 1, minSplit: 1}

	t.Run("an open error surfaces with its status", func(t *testing.T) {
		fake := &fakeUploadExecutor{maxPartSize: 1 << 20, openStatus: http.StatusConflict}
		client := uploadTestClient(t, fake, WithRetries(0, time.Millisecond))

		err := client.uploadMultipart(context.Background(),
			strings.NewReader(string(data)), int64(len(data)), "/r/dir", settings)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "status 409")
		assert.Contains(t, err.Error(), "scripted open failure")
		assert.Equal(t, 0, fake.countKinds("abort"), "nothing to abort before a session exists")
	})

	t.Run("a complete error surfaces with its status", func(t *testing.T) {
		fake := &fakeUploadExecutor{maxPartSize: 1 << 20, completeStatus: http.StatusBadRequest}
		client := uploadTestClient(t, fake, WithRetries(0, time.Millisecond))

		err := client.uploadMultipart(context.Background(),
			strings.NewReader(string(data)), int64(len(data)), "/r/big.bin", settings)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "status 400")
		assert.Equal(t, 1, fake.countKinds("abort"))
	})
}

func TestUploadPartsNotBoundByTheClientTimeout(t *testing.T) {
	// The parts ride the untimed streamClient like slice 2's PUT: the
	// default 2-minute httpClient timeout would kill any transfer
	// longer than 2 minutes. The probe and the open answer instantly, so
	// only the parts see the delay.
	data := uploadTestData()
	fake := &fakeUploadExecutor{maxPartSize: 1 << 20, partDelay: 150 * time.Millisecond}
	client := uploadTestClient(t, fake, WithRetries(0, time.Millisecond), WithTimeout(50*time.Millisecond))

	require.NoError(t, client.uploadMultipart(context.Background(),
		strings.NewReader(string(data)), int64(len(data)), "/r/big.bin",
		uploadSettings{partSize: 4, concurrency: 1, minSplit: 1}))
	assert.Equal(t, data, fake.landed())
}

func TestUploadPartsRideTheStreamDebugTransport(t *testing.T) {
	// The part PUTs ride streamClient whose debug transport renders
	// body-carrying requests headers-only: a wire dump buffers the whole
	// body — or a same-length dummy — which would double the RSS of a
	// large upload.
	data := uploadTestData()
	fake := &fakeUploadExecutor{maxPartSize: 1 << 20}
	client := uploadTestClient(t, fake, WithRetries(0, time.Millisecond))

	oldLevel := log.GetLevel()
	log.SetLevel(log.DebugLevel)
	t.Cleanup(func() { log.SetLevel(oldLevel) })
	read := captureLog(t)

	require.NoError(t, client.uploadMultipart(context.Background(),
		strings.NewReader(string(data)), int64(len(data)), "/r/big.bin",
		uploadSettings{partSize: 4, concurrency: 1, minSplit: 1}))

	dump := read()
	assert.Contains(t, dump, "PUT /files/uploads/")
	assert.Contains(t, dump, "X-Content-Sha256:")
	assert.Contains(t, dump, "Authorization: <HIDDEN", "the token stays redacted")
	assert.NotContains(t, dump, "abcdefghijkl", "the body bytes never enter the log")
	assert.Less(t, len(dump), 16*1024, "the dump stays headers-sized, not body-sized")
}

func TestUploadDefaults(t *testing.T) {
	// The documented defaults, pinned where the flow reads them.
	assert.Equal(t, int64(50<<20), uploadPartSize, "50 MiB parts")
	assert.Equal(t, 4, uploadConcurrency, "4 parts in flight")
	assert.Equal(t, int64(1<<20), minSplitPartSize, "halves stop above the 1 MiB floor")
	assert.Equal(t, int64(100<<20), legacyUploadLimit, "the legacy single-request cap")
}

func TestUploadMultipartSessionWire(t *testing.T) {
	data := uploadTestData()
	settings := uploadSettings{partSize: 4, concurrency: 1, minSplit: 1}
	sum := sha256.Sum256(data)

	t.Run("open sends path, size and the whole-file digest", func(t *testing.T) {
		fake := &fakeUploadExecutor{maxPartSize: 1 << 20}
		client := uploadTestClient(t, fake, WithRetries(0, time.Millisecond))

		require.NoError(t, client.uploadMultipart(context.Background(),
			strings.NewReader(string(data)), int64(len(data)), "/r/big.bin", settings))

		fake.mu.Lock()
		open := fake.calls[0]
		fake.mu.Unlock()
		assert.Equal(t, "open", open.kind)
		assert.Equal(t, "/r/big.bin", open.path)
		assert.Equal(t, int64(len(data)), open.size)
		assert.Equal(t, hex.EncodeToString(sum[:]), open.sha,
			"the digest travels at open, so the executor can verify at complete")
	})

	t.Run("parts carry their bytes at offsets with per-part digests", func(t *testing.T) {
		fake := &fakeUploadExecutor{maxPartSize: 1 << 20}
		client := uploadTestClient(t, fake, WithRetries(0, time.Millisecond))

		require.NoError(t, client.uploadMultipart(context.Background(),
			strings.NewReader(string(data)), int64(len(data)), "/r/big.bin", settings))

		assert.Equal(t, [][2]int64{{0, 4}, {4, 4}, {8, 4}}, fake.partPuts(),
			"parts of part_size at ascending offsets")
		for _, c := range func() []uploadCall {
			fake.mu.Lock()
			defer fake.mu.Unlock()
			return fake.calls
		}() {
			if c.kind != "part" {
				continue
			}
			partSum := sha256.Sum256(c.body)
			assert.Equal(t, hex.EncodeToString(partSum[:]), strings.ToLower(c.sha),
				"each part carries its own digest")
		}
		assert.Equal(t, data, fake.landed(), "the parts assemble into the file")
		assert.Equal(t, "complete", fake.kinds()[len(fake.kinds())-1])
	})

	t.Run("part size is capped by the executor's max_part_size", func(t *testing.T) {
		fake := &fakeUploadExecutor{maxPartSize: 3}
		client := uploadTestClient(t, fake, WithRetries(0, time.Millisecond))

		require.NoError(t, client.uploadMultipart(context.Background(),
			strings.NewReader(string(data)), int64(len(data)), "/r/big.bin", settings))

		assert.Equal(t, [][2]int64{{0, 3}, {3, 3}, {6, 3}, {9, 3}}, fake.partPuts())
		assert.Equal(t, data, fake.landed())
	})
}

func TestUploadCompleteReplyVerification(t *testing.T) {
	data := uploadTestData()
	settings := uploadSettings{partSize: 4, concurrency: 1, minSplit: 1}
	sum := sha256.Sum256(data)

	t.Run("a lying complete digest raises corrupted and aborts", func(t *testing.T) {
		fake := &fakeUploadExecutor{maxPartSize: 1 << 20, lieCompleteSHA: strings.Repeat("ff", 32)}
		client := uploadTestClient(t, fake, WithRetries(0, time.Millisecond))

		err := client.uploadMultipart(context.Background(),
			strings.NewReader(string(data)), int64(len(data)), "/r/big.bin", settings)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "corrupted in transit")
		assert.Contains(t, err.Error(), hex.EncodeToString(sum[:]))
		assert.Equal(t, 1, fake.countKinds("abort"), "the session is freed, not stranded")
	})

	t.Run("a lying complete size raises corrupted and aborts", func(t *testing.T) {
		fake := &fakeUploadExecutor{maxPartSize: 1 << 20, lieCompleteSize: 999}
		client := uploadTestClient(t, fake, WithRetries(0, time.Millisecond))

		err := client.uploadMultipart(context.Background(),
			strings.NewReader(string(data)), int64(len(data)), "/r/big.bin", settings)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "corrupted in transit")
		assert.Contains(t, err.Error(), "999")
		assert.Equal(t, 1, fake.countKinds("abort"))
	})

	t.Run("a complete reply without the optional digest still succeeds", func(t *testing.T) {
		// The contract marks complete's sha256 optional: a 200 has
		// already matched the open-payload digest at the executor.
		fake := &fakeUploadExecutor{maxPartSize: 1 << 20, omitCompleteSHA: true}
		client := uploadTestClient(t, fake, WithRetries(0, time.Millisecond))

		require.NoError(t, client.uploadMultipart(context.Background(),
			strings.NewReader(string(data)), int64(len(data)), "/r/big.bin", settings))
		assert.Equal(t, data, fake.landed())
		assert.Equal(t, 0, fake.countKinds("abort"))
	})
}

func TestUploadConcurrency(t *testing.T) {
	data := bytesSeq(32)

	t.Run("four parts are in flight at once", func(t *testing.T) {
		// The barrier only passes when four parts are truly
		// simultaneous; a smaller pool breaks it and the upload fails.
		barrier := make(chan struct{})
		var arrivals atomic.Int32
		fake := &fakeUploadExecutor{
			maxPartSize: 1 << 20,
			onPart: func() error {
				if arrivals.Add(1) == 4 {
					close(barrier)
					return nil
				}
				select {
				case <-barrier:
					return nil
				case <-time.After(2 * time.Second):
					return fmt.Errorf("barrier timeout: fewer than 4 parts in flight")
				}
			},
		}
		client := uploadTestClient(t, fake, WithRetries(0, time.Millisecond))

		require.NoError(t, client.uploadMultipart(context.Background(),
			strings.NewReader(string(data)), int64(len(data)), "/r/big.bin",
			uploadSettings{partSize: 4, concurrency: 4, minSplit: 1}))

		assert.Equal(t, int32(4), fake.maxInFlight.Load(), "4 parts in flight")
		assert.Equal(t, data, fake.landed())
	})

	t.Run("concurrency one runs serially in plan order", func(t *testing.T) {
		fake := &fakeUploadExecutor{maxPartSize: 1 << 20}
		client := uploadTestClient(t, fake, WithRetries(0, time.Millisecond))

		require.NoError(t, client.uploadMultipart(context.Background(),
			strings.NewReader(string(data)), int64(len(data)), "/r/big.bin",
			uploadSettings{partSize: 4, concurrency: 1, minSplit: 1}))

		assert.Equal(t, int32(1), fake.maxInFlight.Load())
		puts := fake.partPuts()
		want := make([][2]int64, 0, 8)
		for off := int64(0); off < 32; off += 4 {
			want = append(want, [2]int64{off, 4})
		}
		assert.Equal(t, want, puts)
	})
}

// bytesSeq returns n bytes of an ascending pattern.
func bytesSeq(n int) []byte {
	b := make([]byte, n)
	for i := range b {
		b[i] = byte(i)
	}
	return b
}

func TestUploadTimeoutSplit(t *testing.T) {
	data := uploadTestData()

	t.Run("a 524 part is split into halves", func(t *testing.T) {
		// Cloudflare's 524 origin timeout is the CLI's "timed-out
		// part": the stream client has no idle bound of its own, so the
		// proxy's answer is the only split signal. The Python SDK's
		// other trigger (its idle-bound SandboxTimeoutError) has no CLI
		// equivalent.
		fake := &fakeUploadExecutor{
			maxPartSize: 1 << 20,
			failParts:   map[[2]int64]int{{0, 8}: statusCloudflareTimeout},
		}
		client := uploadTestClient(t, fake, WithRetries(0, time.Millisecond))

		require.NoError(t, client.uploadMultipart(context.Background(),
			strings.NewReader(string(data)), int64(len(data)), "/r/big.bin",
			uploadSettings{partSize: 8, concurrency: 1, minSplit: 2}))

		assert.Equal(t, [][2]int64{{0, 8}, {0, 4}, {4, 4}, {8, 4}}, fake.partPuts(),
			"the failed (0,8) attempt, its two halves, then the untouched tail")
		assert.Equal(t, data, fake.landed())
	})

	t.Run("the split recurses until a part lands", func(t *testing.T) {
		fake := &fakeUploadExecutor{
			maxPartSize: 1 << 20,
			failParts:   map[[2]int64]int{{0, 8}: statusCloudflareTimeout, {0, 4}: statusCloudflareTimeout},
		}
		client := uploadTestClient(t, fake, WithRetries(0, time.Millisecond))

		require.NoError(t, client.uploadMultipart(context.Background(),
			strings.NewReader(string(data)), int64(len(data)), "/r/big.bin",
			uploadSettings{partSize: 8, concurrency: 1, minSplit: 2}))

		assert.Equal(t, [][2]int64{{0, 8}, {0, 4}, {0, 2}, {2, 2}, {4, 4}, {8, 4}}, fake.partPuts())
		assert.Equal(t, data, fake.landed())
	})

	t.Run("a part below the split floor fails the upload", func(t *testing.T) {
		// Below twice the floor a split cannot help — the connection
		// is stalling, not slow.
		fake := &fakeUploadExecutor{
			maxPartSize: 1 << 20,
			failParts:   map[[2]int64]int{{0, 4}: statusCloudflareTimeout},
		}
		client := uploadTestClient(t, fake, WithRetries(0, time.Millisecond))

		err := client.uploadMultipart(context.Background(),
			strings.NewReader(string(data)), int64(len(data)), "/r/big.bin",
			uploadSettings{partSize: 4, concurrency: 1, minSplit: 4})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "524")
		assert.Equal(t, 1, fake.countKinds("abort"))
	})
}

func TestUploadAbortOnFailure(t *testing.T) {
	data := uploadTestData()
	settings := uploadSettings{partSize: 4, concurrency: 1, minSplit: 1}

	t.Run("a hard part failure aborts and raises", func(t *testing.T) {
		fake := &fakeUploadExecutor{
			maxPartSize: 1 << 20,
			failParts:   map[[2]int64]int{{4, 4}: http.StatusBadRequest},
		}
		client := uploadTestClient(t, fake, WithRetries(0, time.Millisecond))

		err := client.uploadMultipart(context.Background(),
			strings.NewReader(string(data)), int64(len(data)), "/r/big.bin", settings)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "status 400")
		assert.Equal(t, 1, fake.countKinds("abort"))
		fake.mu.Lock()
		defer fake.mu.Unlock()
		assert.Nil(t, fake.session, "the session is gone")
	})

	t.Run("an abort failure does not mask the original", func(t *testing.T) {
		fake := &fakeUploadExecutor{
			maxPartSize: 1 << 20,
			abortFails:  true,
			failParts:   map[[2]int64]int{{0, 4}: http.StatusBadRequest},
		}
		client := uploadTestClient(t, fake, WithRetries(0, time.Millisecond))

		err := client.uploadMultipart(context.Background(),
			strings.NewReader(string(data)), int64(len(data)), "/r/big.bin", settings)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "status 400", "the original failure surfaces")
		assert.NotContains(t, err.Error(), "unknown upload session")
		assert.Equal(t, 1, fake.countKinds("abort"))
	})

	t.Run("a lost session names the restart", func(t *testing.T) {
		fake := &fakeUploadExecutor{
			maxPartSize: 1 << 20,
			failParts:   map[[2]int64]int{{0, 4}: http.StatusNotFound},
		}
		client := uploadTestClient(t, fake, WithRetries(0, time.Millisecond))

		err := client.uploadMultipart(context.Background(),
			strings.NewReader(string(data)), int64(len(data)), "/r/big.bin", settings)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "session was lost")
		assert.Contains(t, err.Error(), "restart the upload")
		assert.Equal(t, 1, fake.countKinds("abort"))
	})

	t.Run("a canceled upload aborts without draining the queue", func(t *testing.T) {
		// A ^C mid-upload must surface promptly: the workers finish
		// their in-flight part and stop instead of draining the plan,
		// and the abort rides a detached context so the executor's
		// session is still freed.
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		var canceled atomic.Bool
		fake := &fakeUploadExecutor{
			maxPartSize: 1 << 20,
			onPart: func() error {
				if canceled.CompareAndSwap(false, true) {
					cancel()
				}
				return nil
			},
		}
		client := uploadTestClient(t, fake, WithRetries(0, time.Millisecond))

		data := bytesSeq(80)
		err := client.uploadMultipart(ctx,
			strings.NewReader(string(data)), int64(len(data)), "/r/big.bin",
			uploadSettings{partSize: 4, concurrency: 2, minSplit: 1})
		require.Error(t, err)
		assert.ErrorIs(t, err, context.Canceled)
		assert.Equal(t, 1, fake.countKinds("abort"),
			"the abort reaches the executor despite the canceled context")
		assert.Less(t, len(fake.partPuts()), 20,
			"the queue is not drained: in-flight parts finish, no more are pulled")
	})
}

func TestUpload409Resume(t *testing.T) {
	// complete's 409 body carries the missing spans — fill them and
	// complete again; the session stays live across the 409.
	data := uploadTestData()
	settings := uploadSettings{partSize: 4, concurrency: 1, minSplit: 1}

	t.Run("missing ranges from a 409 are filled", func(t *testing.T) {
		fake := &fakeUploadExecutor{maxPartSize: 1 << 20, forgetOffsets: map[int64]bool{4: true}}
		client := uploadTestClient(t, fake, WithRetries(0, time.Millisecond))

		require.NoError(t, client.uploadMultipart(context.Background(),
			strings.NewReader(string(data)), int64(len(data)), "/r/big.bin", settings))

		assert.Equal(t, 2, fake.countKinds("complete"), "the 409 then the completing 200")
		puts := fake.partPuts()
		assert.Len(t, puts, 4, "3 parts plus the one fill")
		assert.Equal(t, [2]int64{4, 4}, puts[len(puts)-1], "the fill re-sends the missing span")
		assert.Equal(t, data, fake.landed())
	})

	t.Run("two missing ranges are both filled", func(t *testing.T) {
		fake := &fakeUploadExecutor{maxPartSize: 1 << 20, forgetOffsets: map[int64]bool{4: true, 8: true}}
		client := uploadTestClient(t, fake, WithRetries(0, time.Millisecond))

		require.NoError(t, client.uploadMultipart(context.Background(),
			strings.NewReader(string(data)), int64(len(data)), "/r/big.bin", settings))

		assert.Equal(t, 2, fake.countKinds("complete"))
		assert.Len(t, fake.partPuts(), 5, "3 parts plus the two fills")
		assert.Equal(t, data, fake.landed())
	})

	t.Run("no progress raises instead of looping", func(t *testing.T) {
		// A fill that does not shrink the missing bytes is the
		// executor not recording them — that loop is cut here rather
		// than spun forever.
		fake := &fakeUploadExecutor{maxPartSize: 1 << 20, neverRecord: map[int64]bool{4: true}}
		client := uploadTestClient(t, fake, WithRetries(0, time.Millisecond))

		err := client.uploadMultipart(context.Background(),
			strings.NewReader(string(data)), int64(len(data)), "/r/big.bin", settings)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "no progress")
		assert.Equal(t, 2, fake.countKinds("complete"))
		assert.Equal(t, 1, fake.countKinds("abort"))
	})

	t.Run("an unparseable 409 raises", func(t *testing.T) {
		fake := &fakeUploadExecutor{
			maxPartSize: 1 << 20,
			neverRecord: map[int64]bool{4: true},
			raw409Body:  `not json`,
		}
		client := uploadTestClient(t, fake, WithRetries(0, time.Millisecond))

		err := client.uploadMultipart(context.Background(),
			strings.NewReader(string(data)), int64(len(data)), "/r/big.bin", settings)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "missing-ranges reply was unusable")
	})

	t.Run("a complete against a lost session names the restart", func(t *testing.T) {
		// The session dies between the parts and the complete — the
		// executor restarted or the session idled out.
		fake := &fakeUploadExecutor{maxPartSize: 1 << 20, completeStatus: http.StatusNotFound}
		client := uploadTestClient(t, fake, WithRetries(0, time.Millisecond))

		err := client.uploadMultipart(context.Background(),
			strings.NewReader(string(data)), int64(len(data)), "/r/big.bin", settings)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "session was lost")
		assert.Contains(t, err.Error(), "restart the upload")
		assert.Equal(t, 1, fake.countKinds("abort"))
	})
}

func TestUploadOpenReplyValidation(t *testing.T) {
	data := uploadTestData()
	settings := uploadSettings{partSize: 4, concurrency: 1, minSplit: 1}

	t.Run("an open reply without upload_id raises without aborting", func(t *testing.T) {
		fake := &fakeUploadExecutor{maxPartSize: 1 << 20, openReply: `{"max_part_size":4}`}
		client := uploadTestClient(t, fake, WithRetries(0, time.Millisecond))

		err := client.uploadMultipart(context.Background(),
			strings.NewReader(string(data)), int64(len(data)), "/r/big.bin", settings)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "upload_id")
		assert.Equal(t, 0, fake.countKinds("abort"), "no id means no session to abort")
		assert.Equal(t, 0, fake.countKinds("part"))
	})

	t.Run("an open reply without max_part_size aborts the live session", func(t *testing.T) {
		// A valid upload_id means the session exists at the executor;
		// failing on the cap must abort it, not strand its temp file.
		fake := &fakeUploadExecutor{maxPartSize: 1 << 20, openReply: `{"upload_id":"u-1"}`}
		client := uploadTestClient(t, fake, WithRetries(0, time.Millisecond))

		err := client.uploadMultipart(context.Background(),
			strings.NewReader(string(data)), int64(len(data)), "/r/big.bin", settings)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "max_part_size")
		assert.Equal(t, 1, fake.countKinds("abort"))
		assert.Equal(t, 0, fake.countKinds("part"))
	})
}
