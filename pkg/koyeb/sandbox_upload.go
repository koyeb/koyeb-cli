package koyeb

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	log "github.com/sirupsen/logrus"
)

// Multipart uploads: a file larger than one request's budget goes up as
// offset-addressed parts into a session at the executor (POST /files/uploads
// opens one, PUT /files/uploads/{id}?offset=N writes a part, POST
// /files/uploads/{id}/complete checks coverage and renames into place,
// DELETE /files/uploads/{id} aborts). Parts may be sent in any order and in
// parallel, and a retried part overwrites the same bytes — every part PUT is
// idempotent at the executor.
//
// The parts ride the untimed streamClient (see sandbox_files_routes.go):
// each attempt reads its range from disk and streams it, so nothing is
// buffered between retries and memory stays constant regardless of the
// file's size. The debug transport dumps those requests headers-only for
// the same reason.

const (
	// uploadPartSize is the client's part size: 50 MiB per the platform
	// contract, capped by the executor's max_part_size in the open reply.
	uploadPartSize = int64(50 << 20)
	// uploadConcurrency is the number of parts in flight.
	uploadConcurrency = 4
	// minSplitPartSize is the floor for splitting a timed-out part in
	// half: below twice this a split cannot help — the connection is
	// stalling, not slow.
	minSplitPartSize = int64(1 << 20)
	// legacyUploadLimit caps uploads to pre-v0.0.24 executors: their
	// single-request JSON uploads cross the platform's proxies, which
	// cap request bodies at 100 MiB (cloud-runtime) or 200 MiB (Nomad).
	legacyUploadLimit = int64(100 << 20)
	// statusCloudflareTimeout is Cloudflare's origin-timeout answer on
	// the cloud-runtime hop — the wire shape a stalled part surfaces as,
	// and the signal to send smaller halves.
	statusCloudflareTimeout = 524
)

// errUploadSessionLost means the executor no longer knows the session —
// it restarted, or the session idled past its expiry. Sessions live in
// the executor's memory; a lost one cannot be resumed, only restarted.
var errUploadSessionLost = errors.New("upload session was lost — the executor restarted or the session expired; restart the upload")

// uploadRange is a half-open [start, end) byte range of the file.
type uploadRange struct {
	start, end int64
}

// uploadSettings sizes a multipart session's mechanics. The public flow
// reads them from the constants above; the tests drive the same flow with
// small values.
type uploadSettings struct {
	partSize    int64 // bytes per part, capped by the executor's max_part_size
	concurrency int   // parts in flight
	minSplit    int64 // a timed-out part splits into halves down to twice this floor
}

// openUploadRequest is the POST /files/uploads body.
type openUploadRequest struct {
	Path   string `json:"path"`
	Size   int64  `json:"size"`
	SHA256 string `json:"sha256,omitempty"` // the whole-file digest, checked at complete
}

// UploadFile uploads a local file to the sandbox, streaming from disk with
// constant memory: one PUT /files for files at or below the 50 MiB part
// size, a multipart session above it on executors v0.0.24+. Older
// executors keep the legacy single-request JSON upload, which the
// platform's proxies cap; files above that cap are refused up front with a
// message naming the cause. It returns the number of bytes uploaded.
func (c *SandboxClient) UploadFile(ctx context.Context, localPath, remotePath string) (int64, error) {
	routed, err := c.hasFilesRoutes(ctx)
	if err != nil {
		return 0, err
	}
	f, err := os.Open(localPath)
	if err != nil {
		return 0, fmt.Errorf("failed to open local file: %w", err)
	}
	defer f.Close()
	info, err := f.Stat()
	if err != nil {
		return 0, fmt.Errorf("failed to stat local file: %w", err)
	}

	size := info.Size()
	if !routed {
		return size, c.legacyUploadFile(ctx, f, size, remotePath)
	}
	if size > uploadPartSize {
		return size, c.uploadMultipart(ctx, f, size, remotePath, uploadSettings{
			partSize:    uploadPartSize,
			concurrency: uploadConcurrency,
			minSplit:    minSplitPartSize,
		})
	}
	return size, c.uploadSingle(ctx, f, size, remotePath)
}

// legacyUploadFile keeps the legacy JSON route for pre-v0.0.24 executors:
// the whole file in one buffered POST. The platform's proxies cap request
// bodies at 100 MiB (cloud-runtime) or 200 MiB (Nomad), so a larger file
// is refused here instead of buffering and dying at the proxy with an
// opaque error. The legacy JSON body also inflates control characters
// (up to six bytes each), so files somewhat below the cap can still fail
// there — that is the pre-existing behavior of the legacy route.
func (c *SandboxClient) legacyUploadFile(ctx context.Context, f *os.File, size int64, remotePath string) error {
	if size > legacyUploadLimit {
		return fmt.Errorf(
			"file is %d bytes: this executor predates v0.0.24 and uploads in a single request, "+
				"which the platform proxy caps at 100 MiB (200 MiB on Nomad) — "+
				"upgrade the sandbox image to v0.0.24+ to upload larger files as multipart parts", size)
	}
	content, err := io.ReadAll(f)
	if err != nil {
		return err
	}
	return c.legacyWriteFile(ctx, remotePath, content)
}

// uploadSingle streams the whole file through one PUT /files: the body is
// read from disk on every attempt, the digest rides X-Content-SHA256 so
// the executor verifies while streaming, and the {size, sha256} answer is
// verified against the local file before reporting success.
func (c *SandboxClient) uploadSingle(ctx context.Context, src io.ReaderAt, size int64, remotePath string) error {
	resp, digest, err := c.putRange(ctx, filesRoutePath(remotePath), src, 0, size)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return filesRouteError(resp)
	}
	var reply struct {
		Size   int64  `json:"size"`
		SHA256 string `json:"sha256"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&reply); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}
	// PUT /files always answers both fields; the executor has already
	// matched the digest header, this catches a lie on the way back.
	if reply.Size != size || reply.SHA256 != digest {
		return fmt.Errorf("upload corrupted in transit: executor size %d sha256 %s, local size %d sha256 %s",
			reply.Size, reply.SHA256, size, digest)
	}
	return nil
}

// uploadMultipart uploads src as a multipart session: open, parts,
// complete — aborting the session on any failure after the open. The
// whole-file digest travels at open, so the executor can verify the
// assembled file at complete.
func (c *SandboxClient) uploadMultipart(ctx context.Context, src io.ReaderAt, size int64, remotePath string, s uploadSettings) error {
	digest, err := rangeDigest(src, 0, size)
	if err != nil {
		return err
	}

	resp, err := c.request(ctx, http.MethodPost, "/files/uploads", &openUploadRequest{
		Path: remotePath, Size: size, SHA256: digest,
	})
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		openErr := filesRouteError(resp)
		resp.Body.Close()
		return openErr
	}
	var open struct {
		UploadID    string `json:"upload_id"`
		MaxPartSize int64  `json:"max_part_size"`
	}
	decodeErr := json.NewDecoder(resp.Body).Decode(&open)
	resp.Body.Close()
	if decodeErr != nil {
		return fmt.Errorf("failed to decode response: %w", decodeErr)
	}
	if open.UploadID == "" {
		// No id means no session to abort: if one exists at the executor
		// it is unreachable and left to the idle janitor.
		return fmt.Errorf("upload session reply missing upload_id")
	}

	if err := c.uploadSession(ctx, src, size, digest, open.UploadID, open.MaxPartSize, s); err != nil {
		c.abortUpload(ctx, open.UploadID)
		return err
	}
	log.Debugf("Uploaded %s (%d bytes) as a multipart session", remotePath, size)
	return nil
}

// uploadSession runs the parts and the complete of an open session. A
// valid upload_id means the session lives at the executor: every failure
// from here on is the caller's to abort.
func (c *SandboxClient) uploadSession(ctx context.Context, src io.ReaderAt, size int64, digest, uploadID string, maxPartSize int64, s uploadSettings) error {
	if maxPartSize < 1 {
		return fmt.Errorf("upload session reply missing max_part_size")
	}
	es := s
	es.partSize = min(s.partSize, maxPartSize)
	if err := c.putRanges(ctx, src, uploadID, []uploadRange{{0, size}}, es); err != nil {
		return err
	}
	return c.completeUpload(ctx, src, uploadID, size, digest, es)
}

// completeUpload checks coverage at the executor and verifies the answer
// against the local file before reporting success. A 409 carries the
// missing spans — the session stays live across it, so they are filled and
// complete retried; every fill must shrink the missing bytes, a set that
// does not is the executor not recording them, and that loop is cut here
// rather than spun forever.
func (c *SandboxClient) completeUpload(ctx context.Context, src io.ReaderAt, uploadID string, size int64, digest string, s uploadSettings) error {
	completePath := "/files/uploads/" + url.PathEscape(uploadID) + "/complete"
	prevMissing := int64(-1)
	for {
		resp, err := c.request(ctx, http.MethodPost, completePath, nil)
		if err != nil {
			return err
		}
		switch resp.StatusCode {
		case http.StatusOK:
			var reply struct {
				Size   int64  `json:"size"`
				SHA256 string `json:"sha256"`
			}
			decodeErr := json.NewDecoder(resp.Body).Decode(&reply)
			resp.Body.Close()
			if decodeErr != nil {
				return fmt.Errorf("failed to decode response: %w", decodeErr)
			}
			if reply.Size != size {
				return fmt.Errorf("upload corrupted in transit: executor size %d != local %d", reply.Size, size)
			}
			// The contract marks complete's sha256 optional: a 200 has
			// already matched the open-payload digest at the executor.
			if reply.SHA256 != "" && reply.SHA256 != digest {
				return fmt.Errorf("upload corrupted in transit: executor sha256 %s != %s", reply.SHA256, digest)
			}
			return nil
		case http.StatusNotFound:
			_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 4096))
			resp.Body.Close()
			return errUploadSessionLost
		case http.StatusConflict:
			missing, fillErr := missingRangesFrom409(resp.Body)
			resp.Body.Close()
			if fillErr != nil {
				return fmt.Errorf("upload incomplete and the missing-ranges reply was unusable: %w", fillErr)
			}
			missingBytes := int64(0)
			for _, r := range missing {
				missingBytes += r.end - r.start
			}
			if prevMissing >= 0 && missingBytes >= prevMissing {
				return fmt.Errorf("upload made no progress filling missing ranges: %v", missing)
			}
			prevMissing = missingBytes
			if err := c.putRanges(ctx, src, uploadID, missing, s); err != nil {
				return err
			}
		default:
			completeErr := filesRouteError(resp)
			resp.Body.Close()
			return completeErr
		}
	}
}

// missingRangesFrom409 decodes a complete-409 body's missing spans. A body
// without usable ranges is an error: the resume signal is the contract,
// not a guess.
func missingRangesFrom409(body io.Reader) ([]uploadRange, error) {
	var parsed struct {
		Error  string `json:"error"`
		Ranges []struct {
			Start int64 `json:"start"`
			End   int64 `json:"end"`
		} `json:"ranges"`
	}
	if err := json.NewDecoder(io.LimitReader(body, 1<<20)).Decode(&parsed); err != nil {
		return nil, err
	}
	if len(parsed.Ranges) == 0 {
		return nil, fmt.Errorf("no ranges in the reply")
	}
	ranges := make([]uploadRange, 0, len(parsed.Ranges))
	for _, r := range parsed.Ranges {
		if r.Start < 0 || r.End <= r.Start {
			return nil, fmt.Errorf("invalid range [%d, %d)", r.Start, r.End)
		}
		ranges = append(ranges, uploadRange{r.Start, r.End})
	}
	return ranges, nil
}

// abortUpload aborts the session, freeing its temp file at the executor.
// It rides a context detached from the caller's: a canceled upload (^C)
// must still free the session instead of leaving its temp file to the
// idle janitor. Its own failures are swallowed — never mask the failure
// that caused the abort.
func (c *SandboxClient) abortUpload(ctx context.Context, uploadID string) {
	abortCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 30*time.Second)
	defer cancel()
	resp, err := c.request(abortCtx, http.MethodDelete, "/files/uploads/"+url.PathEscape(uploadID), nil)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 4096))
}

// putRanges writes the given byte ranges as parts, s.concurrency in
// flight, failing fast on the first error: the other workers finish their
// in-flight part and stop instead of draining the plan.
func (c *SandboxClient) putRanges(ctx context.Context, src io.ReaderAt, uploadID string, ranges []uploadRange, s uploadSettings) error {
	plan := planUploadRanges(ranges, s.partSize)
	if len(plan) == 0 {
		return nil
	}
	poolCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	var next atomic.Int64
	var mu sync.Mutex
	var firstErr error
	var wg sync.WaitGroup
	for range min(s.concurrency, len(plan)) {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for poolCtx.Err() == nil {
				i := int(next.Add(1)) - 1
				if i >= len(plan) {
					return
				}
				if err := c.putPartWithSplits(poolCtx, src, uploadID, plan[i], s); err != nil {
					mu.Lock()
					if firstErr == nil {
						firstErr = err
					}
					mu.Unlock()
					cancel()
					return
				}
			}
		}()
	}
	wg.Wait()
	return firstErr
}

// planUploadRanges slices the ranges into parts of at most partSize, in
// ascending order within each range.
func planUploadRanges(ranges []uploadRange, partSize int64) []uploadRange {
	var plan []uploadRange
	for _, r := range ranges {
		for start := r.start; start < r.end; start += partSize {
			plan = append(plan, uploadRange{start, min(start+partSize, r.end)})
		}
	}
	return plan
}

// putPartWithSplits writes one planned range as one or more parts. A
// part that keeps failing with Cloudflare's 524 origin timeout is
// retried as two halves — a smaller part needs less time on the wire —
// recursing down to twice the split floor, below which the stall is
// fatal. A 404 means the executor no longer knows the session.
func (c *SandboxClient) putPartWithSplits(ctx context.Context, src io.ReaderAt, uploadID string, span uploadRange, s uploadSettings) error {
	pending := []uploadRange{span}
	for len(pending) > 0 {
		sp := pending[len(pending)-1]
		pending = pending[:len(pending)-1]
		status, err := c.putPartRetried(ctx, src, uploadID, sp.start, sp.end-sp.start)
		if err == nil {
			continue
		}
		switch {
		case status == http.StatusNotFound:
			return errUploadSessionLost
		case status == statusCloudflareTimeout && sp.end-sp.start >= 2*s.minSplit:
			mid := sp.start + (sp.end-sp.start)/2
			// The lower half is pushed last so it pops first, matching
			// the plan's ascending order.
			pending = append(pending, uploadRange{mid, sp.end}, uploadRange{sp.start, mid})
		default:
			return err
		}
	}
	return nil
}

// streamStatusError is the retry-exhausted failure of a streamed PUT,
// carrying the last HTTP status seen — the split logic tells a 524
// (Cloudflare's origin timeout) from other server errors by it.
type streamStatusError struct {
	status int
	err    error
}

func (e *streamStatusError) Error() string { return e.err.Error() }
func (e *streamStatusError) Unwrap() error { return e.err }

// putPartRetried PUTs one part at its offset, retrying like putRange.
func (c *SandboxClient) putPartRetried(ctx context.Context, src io.ReaderAt, uploadID string, off, length int64) (int, error) {
	partPath := "/files/uploads/" + url.PathEscape(uploadID) + "?" +
		url.Values{"offset": []string{strconv.FormatInt(off, 10)}}.Encode()
	resp, _, err := c.putRange(ctx, partPath, src, off, length)
	if err != nil {
		var se *streamStatusError
		if errors.As(err, &se) {
			return se.status, err
		}
		return 0, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 4096))
		return resp.StatusCode, nil
	}
	return resp.StatusCode, filesRouteError(resp)
}

// putRange PUTs src[off:off+length) at urlPath on the untimed streaming
// client, retrying 5xx statuses and network errors like streamPut does.
// Every attempt re-reads the range from src and re-digests it, so nothing
// is buffered between retries and the X-Content-SHA256 header always
// matches the bytes on the wire. It returns the last attempt's response
// (nil when retries were exhausted), the digest that attempt sent, and
// the error.
func (c *SandboxClient) putRange(ctx context.Context, urlPath string, src io.ReaderAt, off, length int64) (*http.Response, string, error) {
	var lastErr error
	var lastStatus int
	var lastDigest string
	for attempt := 0; attempt <= c.maxRetries; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return nil, "", ctx.Err()
			case <-time.After(c.retryDelay * time.Duration(attempt)):
			}
		}

		digest, err := rangeDigest(src, off, length)
		if err != nil {
			lastErr = err
			lastStatus = 0
			continue
		}
		lastDigest = digest

		req, err := http.NewRequestWithContext(ctx, http.MethodPut, c.baseURL+urlPath,
			io.NewSectionReader(src, off, length))
		if err != nil {
			return nil, "", fmt.Errorf("failed to create request: %w", err)
		}
		req.ContentLength = length
		req.Header.Set("Authorization", "Bearer "+c.secret)
		if c.routingKey != "" {
			req.Header.Set("X-Routing-Key", c.routingKey)
		}
		req.Header.Set("Content-Type", "application/octet-stream")
		req.Header.Set("X-Content-SHA256", digest)

		resp, err := c.streamClient.Do(req)
		if err != nil {
			lastErr = err
			lastStatus = 0
			continue
		}
		if resp.StatusCode >= 500 {
			lastErr = fmt.Errorf("server error: status %d", resp.StatusCode)
			lastStatus = resp.StatusCode
			resp.Body.Close()
			continue
		}
		return resp, lastDigest, nil
	}
	err := fmt.Errorf("request failed after %d retries: %w", c.maxRetries, lastErr)
	if lastStatus != 0 {
		err = &streamStatusError{status: lastStatus, err: err}
	}
	return nil, lastDigest, err
}

// rangeDigest hashes src[off:off+length) in 1 MiB chunks: constant memory.
func rangeDigest(src io.ReaderAt, off, length int64) (string, error) {
	h := sha256.New()
	buf := make([]byte, 1<<20)
	if _, err := io.CopyBuffer(h, io.NewSectionReader(src, off, length), buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}
