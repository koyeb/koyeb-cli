package koyeb

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestShellQuote(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{"simple", "'simple'"},
		{"with space", "'with space'"},
		{"it's", "'it'\\''s'"},
		{"$(rm -rf /)", "'$(rm -rf /)'"},
		{"back`tick", "'back`tick'"},
		{"", "''"},
	}

	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			assert.Equal(t, tt.want, shellQuote(tt.in))
		})
	}
}

func TestMvCommand(t *testing.T) {
	assert.Equal(t, "mv '/tmp/old name' '/tmp/new name'", mvCommand("/tmp/old name", "/tmp/new name"))
	assert.Equal(t, "mv '/a' '/b'", mvCommand("/a", "/b"))
}

func TestTestCommand(t *testing.T) {
	tests := []struct {
		expression string
		want       string
	}{
		{"-e", "test -e '/some/path'"},
		{"-f", "test -f '/some/path'"},
		{"-d", "test -d '/some/path'"},
	}

	for _, tt := range tests {
		t.Run(tt.expression, func(t *testing.T) {
			assert.Equal(t, tt.want, testCommand(tt.expression, "/some/path"))
		})
	}
}

func TestSandboxFsSubcommands(t *testing.T) {
	cmd := NewSandboxCmd()

	for _, sub := range []string{"rename", "move", "exists", "is-file", "is-dir"} {
		fsCmd, _, err := cmd.Find([]string{"fs"})
		require.NoError(t, err)
		_, _, err = fsCmd.Find([]string{sub})
		require.NoError(t, err, "sandbox fs must register the %q subcommand", sub)
	}
}

func TestSandboxFsSubcommandArgShapes(t *testing.T) {
	cmd := NewSandboxCmd()
	fsCmd, _, err := cmd.Find([]string{"fs"})
	require.NoError(t, err)

	tests := []struct {
		sub   string
		use   string
		args  []string
		valid bool
	}{
		{"rename", "rename NAME OLD_PATH NEW_PATH", []string{"sb", "/a", "/b"}, true},
		{"rename", "rename NAME OLD_PATH NEW_PATH", []string{"sb", "/a"}, false},
		{"move", "move NAME SOURCE_PATH DESTINATION_PATH", []string{"sb", "/a", "/b"}, true},
		{"move", "move NAME SOURCE_PATH DESTINATION_PATH", []string{"sb"}, false},
		{"exists", "exists NAME PATH", []string{"sb", "/a"}, true},
		{"exists", "exists NAME PATH", []string{"sb"}, false},
		{"is-file", "is-file NAME PATH", []string{"sb", "/a"}, true},
		{"is-dir", "is-dir NAME PATH", []string{"sb", "/a"}, true},
	}

	for _, tt := range tests {
		t.Run(tt.sub, func(t *testing.T) {
			subCmd, _, err := fsCmd.Find([]string{tt.sub})
			require.NoError(t, err)
			assert.Equal(t, tt.use, subCmd.Use)
			assert.Equal(t, tt.valid, subCmd.Args(subCmd, tt.args) == nil)
		})
	}
}

// fakeSandboxClient is the in-memory adapter for SandboxClientInterface.
type fakeSandboxClient struct {
	runReqs []*RunRequest
	runResp *RunResponse
	runErr  error
}

func (f *fakeSandboxClient) Run(_ context.Context, req *RunRequest) (*RunResponse, error) {
	f.runReqs = append(f.runReqs, req)
	if f.runErr != nil {
		return nil, f.runErr
	}
	if f.runResp != nil {
		return f.runResp, nil
	}
	return &RunResponse{}, nil
}

func (f *fakeSandboxClient) RunStreaming(
	_ context.Context,
	_ *RunRequest,
	_ func(string, string),
	_ func(int, bool),
) error {
	return nil
}

func (f *fakeSandboxClient) WriteFile(context.Context, string, []byte) error { return nil }
func (f *fakeSandboxClient) ReadFile(context.Context, string) ([]byte, error) {
	return nil, nil
}
func (f *fakeSandboxClient) DeleteFile(context.Context, string) error { return nil }
func (f *fakeSandboxClient) MakeDir(context.Context, string) error    { return nil }
func (f *fakeSandboxClient) DeleteDir(context.Context, string) error  { return nil }
func (f *fakeSandboxClient) ListDir(context.Context, string) ([]DirEntry, error) {
	return nil, nil
}
func (f *fakeSandboxClient) StatFile(context.Context, string) (*DirEntry, error) {
	return nil, nil
}
func (f *fakeSandboxClient) BindPort(context.Context, string) (*PortResponse, error) {
	return nil, nil
}
func (f *fakeSandboxClient) UnbindPort(context.Context) (*PortResponse, error) {
	return nil, nil
}
func (f *fakeSandboxClient) StartProcess(context.Context, *ProcessRequest) (*StartProcessResponse, error) {
	return nil, nil
}
func (f *fakeSandboxClient) ListProcesses(context.Context) ([]ProcessInfo, error) {
	return nil, nil
}
func (f *fakeSandboxClient) KillProcess(context.Context, string) error { return nil }
func (f *fakeSandboxClient) StreamProcessLogs(context.Context, string, bool, func(string, string, string)) error {
	return nil
}
func (f *fakeSandboxClient) Health(context.Context) (*HealthResponse, error) {
	return &HealthResponse{Healthy: true}, nil
}

func TestFsMovePathWiring(t *testing.T) {
	ctx := &CLIContext{Context: context.Background()}
	h := &SandboxHandler{}

	t.Run("runs the quoted mv command", func(t *testing.T) {
		fake := &fakeSandboxClient{}
		require.NoError(t, h.fsMovePath(ctx, fake, "renaming", "/tmp/old name", "/tmp/new name"))
		require.Len(t, fake.runReqs, 1)
		assert.Equal(t, "mv '/tmp/old name' '/tmp/new name'", fake.runReqs[0].Cmd)
	})

	t.Run("maps non-zero exits to the operation error", func(t *testing.T) {
		fake := &fakeSandboxClient{runResp: &RunResponse{Code: 1, Stderr: "No such file"}}
		err := h.fsMovePath(ctx, fake, "moving", "/a", "/b")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "moving")
		assert.Contains(t, err.Error(), "No such file")
	})

	t.Run("api errors surface as connection errors", func(t *testing.T) {
		fake := &fakeSandboxClient{runErr: fmt.Errorf("unreachable")}
		err := h.fsMovePath(ctx, fake, "renaming", "/a", "/b")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "sandbox API request failed")
	})
}

func TestFsTestPathWiring(t *testing.T) {
	ctx := &CLIContext{Context: context.Background()}

	for _, tt := range []struct {
		expression string
		wantCmd    string
	}{
		{"-e", "test -e '/some/path'"},
		{"-f", "test -f '/some/path'"},
		{"-d", "test -d '/some/path'"},
	} {
		t.Run(tt.expression, func(t *testing.T) {
			fake := &fakeSandboxClient{}
			require.NoError(t, fsTest(ctx, fake, "/some/path", tt.expression))
			require.Len(t, fake.runReqs, 1)
			assert.Equal(t, tt.wantCmd, fake.runReqs[0].Cmd)
		})
	}
}
