package koyeb

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/koyeb/koyeb-cli/pkg/koyeb/errors"
)

// fakeExecClient fakes the ExecClient seam: the first call validates the
// remote path, the second runs the tar exec. Tests configure which call
// fails and which pipe end to close so the tar goroutine can finish.
type fakeExecClient struct {
	calls       int
	failOn      int
	drainOn     int
	closeStdout bool
}

func (f *fakeExecClient) Exec(ctx context.Context, id ExecId, cmd []string) (int, error) {
	return 0, nil
}

func (f *fakeExecClient) ExecWithStreams(ctx context.Context, stdStreams *StdStreams, id ExecId, cmd []string) (int, error) {
	f.calls++
	if f.calls == f.drainOn {
		_, _ = io.Copy(io.Discard, stdStreams.Stdin)
	}
	if f.calls == f.failOn {
		if c, ok := stdStreams.Stdin.(io.Closer); ok {
			_ = c.Close()
		}
		if f.closeStdout {
			if c, ok := stdStreams.Stdout.(io.Closer); ok {
				_ = c.Close()
			}
		}
		return 1, fmt.Errorf("remote exec failed")
	}
	return 0, nil
}

// The exec failure path reads the tar goroutine's error: run under the race
// detector (go test -race) to pin that the read is synchronized with the
// goroutine that writes it.
func TestCopyToInstanceExecFailureReturnsCLIError(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "file.txt")
	require.NoError(t, os.WriteFile(src, []byte("data"), 0o644))

	fake := &fakeExecClient{failOn: 2} // call 1: remote validation, call 2: the tar exec
	ctx := &CLIContext{Context: context.Background(), ExecClient: fake}

	manager, err := NewCopyManager(
		&FileSpec{FilePath: src},
		&FileSpec{InstanceID: "instance-id", FilePath: "/remote"},
	)
	require.NoError(t, err)

	cpErr := manager.Copy(ctx)
	require.Error(t, cpErr)
	var cliErr *errors.CLIError
	require.ErrorAs(t, cpErr, &cliErr)
	assert.Contains(t, cliErr.Error(), "Error while copying")
}

func TestCopyFromInstanceExecFailureReturnsCLIError(t *testing.T) {
	dir := t.TempDir()
	dst := filepath.Join(dir, "file.txt")
	require.NoError(t, os.WriteFile(dst, []byte("data"), 0o644))

	fake := &fakeExecClient{failOn: 2, closeStdout: true}
	ctx := &CLIContext{Context: context.Background(), ExecClient: fake}

	manager, err := NewCopyManager(
		&FileSpec{InstanceID: "instance-id", FilePath: "/remote/file.txt"},
		&FileSpec{FilePath: dst},
	)
	require.NoError(t, err)

	cpErr := manager.Copy(ctx)
	require.Error(t, cpErr)
	var cliErr *errors.CLIError
	require.ErrorAs(t, cpErr, &cliErr)
}

func TestCopyToInstanceSuccess(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "file.txt")
	require.NoError(t, os.WriteFile(src, []byte("data"), 0o644))

	fake := &fakeExecClient{drainOn: 2} // drains the tar stream, reports success
	ctx := &CLIContext{Context: context.Background(), ExecClient: fake}

	manager, err := NewCopyManager(
		&FileSpec{FilePath: src},
		&FileSpec{InstanceID: "instance-id", FilePath: "/remote"},
	)
	require.NoError(t, err)

	require.NoError(t, manager.Copy(ctx))
}
