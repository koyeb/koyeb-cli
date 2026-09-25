package koyeb

import (
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
