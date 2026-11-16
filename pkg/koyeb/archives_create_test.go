package koyeb

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestUploadArchiveSendsContentLength(t *testing.T) {
	content := []byte("archive-content")
	path := filepath.Join(t.TempDir(), "archive.tar.gz")
	if err := os.WriteFile(path, content, 0o600); err != nil {
		t.Fatal(err)
	}
	file, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()

	var gotLength int64
	var gotTransferEncoding []string
	var gotBody []byte
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotLength = r.ContentLength
		gotTransferEncoding = r.TransferEncoding
		gotBody, _ = io.ReadAll(r.Body)
	}))
	defer server.Close()

	if err := NewArchiveHandler().uploadArchive(file, int64(len(content)), server.URL); err != nil {
		t.Fatal(err)
	}

	if gotLength != int64(len(content)) {
		t.Errorf("Content-Length = %d, want %d", gotLength, len(content))
	}
	if len(gotTransferEncoding) != 0 {
		t.Errorf("Transfer-Encoding = %v, want none", gotTransferEncoding)
	}
	if string(gotBody) != string(content) {
		t.Errorf("body = %q, want %q", gotBody, content)
	}
}
