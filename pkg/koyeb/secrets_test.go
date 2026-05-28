package koyeb

import (
	"os"
	"strings"
	"testing"

	"github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func secretFlags(t *testing.T, args ...string) *pflag.FlagSet {
	t.Helper()

	flags := pflag.NewFlagSet("test", pflag.ContinueOnError)
	addSecretFlags(flags, nil)
	require.NoError(t, flags.Parse(args))
	return flags
}

func TestGetSecretValueFromStdin(t *testing.T) {
	input, err := os.CreateTemp(t.TempDir(), "secret-stdin")
	require.NoError(t, err)
	_, err = input.WriteString("line 1\n\nline 2\n")
	require.NoError(t, err)
	_, err = input.Seek(0, 0)
	require.NoError(t, err)
	defer input.Close()

	oldStdin := os.Stdin
	os.Stdin = input
	defer func() {
		os.Stdin = oldStdin
	}()

	value, abort, err := getSecretValue(secretFlags(t, "--value-from-stdin"))

	require.NoError(t, err)
	assert.False(t, abort)
	assert.Equal(t, "line 1\n\nline 2", value)
}

func TestGetSecretValueFromStdinPreservesLongLines(t *testing.T) {
	input, err := os.CreateTemp(t.TempDir(), "secret-stdin")
	require.NoError(t, err)
	_, err = input.WriteString(strings.Repeat("a", 64*1024+1))
	require.NoError(t, err)
	_, err = input.Seek(0, 0)
	require.NoError(t, err)
	defer input.Close()

	oldStdin := os.Stdin
	os.Stdin = input
	defer func() {
		os.Stdin = oldStdin
	}()

	value, abort, err := getSecretValue(secretFlags(t, "--value-from-stdin"))

	require.NoError(t, err)
	assert.False(t, abort)
	assert.Equal(t, strings.Repeat("a", 64*1024+1), value)
}
