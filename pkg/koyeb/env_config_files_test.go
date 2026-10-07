package koyeb

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
	"github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// The env/config-file flag bundle is the single registration and parse
// seam shared by the service, sandbox and pool surfaces. These tests pin
// the bundle contract; the per-surface suites pin the merged behavior.

func TestAddEnvConfigFilesFlagsServiceSurface(t *testing.T) {
	flags := pflag.NewFlagSet("service", pflag.ContinueOnError)
	addEnvConfigFilesFlags(flags, serviceEnvConfigFilesFlagUsage)

	expected := []struct {
		name     string
		defValue string
		usage    string
	}{
		{
			"env", "[]",
			"Update service environment variables using the format KEY=VALUE, for example --env FOO=bar\n" +
				"To use the value of a secret as an environment variable, use the following syntax: --env FOO={{secret.bar}}\n" +
				"To delete an environment variable, prefix its name with '!', for example --env '!FOO'\n",
		},
		{
			"config-file", "[]",
			"Copy a local file to your service container using the format LOCAL_FILE:PATH:[PERMISSIONS]\n" +
				"for example --config-file /etc/data.yaml:/etc/data.yaml:0644\n" +
				"To delete a config file, use !PATH, for example --config-file !/etc/data.yaml\n",
		},
	}
	for _, want := range expected {
		flag := flags.Lookup(want.name)
		require.NotNil(t, flag, "--%s must be registered by the env/config-file bundle", want.name)
		assert.Equal(t, want.defValue, flag.DefValue, "--%s default value", want.name)
		assert.Equal(t, want.usage, flag.Usage, "--%s help text", want.name)
	}
}

func TestAddEnvConfigFilesFlagsSandboxPoolSurface(t *testing.T) {
	flags := pflag.NewFlagSet("sandbox-pool", pflag.ContinueOnError)
	addEnvConfigFilesFlags(flags, sandboxPoolEnvConfigFilesFlagUsage)

	expected := []struct {
		name     string
		defValue string
		usage    string
	}{
		{"env", "[]", "Environment variables (KEY=VALUE)"},
		{"config-file", "[]", "Config files (LOCAL:REMOTE:PERMS)"},
	}
	for _, want := range expected {
		flag := flags.Lookup(want.name)
		require.NotNil(t, flag, "--%s must be registered by the env/config-file bundle", want.name)
		assert.Equal(t, want.defValue, flag.DefValue, "--%s default value", want.name)
		assert.Equal(t, want.usage, flag.Usage, "--%s help text", want.name)
	}
}

// envConfigFilesFlagNames returns the flag names the bundle registers for
// the given skin — the bundle union the surface flag-set tests derive
// their expectations from.
func envConfigFilesFlagNames(usage envConfigFilesFlagUsage) []string {
	flags := pflag.NewFlagSet("env-config-files", pflag.ContinueOnError)
	addEnvConfigFilesFlags(flags, usage)
	names := make([]string, 0, 2)
	flags.VisitAll(func(f *pflag.Flag) { names = append(names, f.Name) })
	return names
}

// envConfigFilesTestFlagSet registers the bundle surface the parse step
// consumes.
func envConfigFilesTestFlagSet() *pflag.FlagSet {
	flags := pflag.NewFlagSet("env-config-files", pflag.ContinueOnError)
	addEnvConfigFilesFlags(flags, serviceEnvConfigFilesFlagUsage)
	return flags
}

// liveEnvConfigFilesDefinition returns a definition carrying a live env
// variable and config file, so the changed-only merge conventions are
// observable.
func liveEnvConfigFilesDefinition() *koyeb.DeploymentDefinition {
	def := koyeb.NewDeploymentDefinitionWithDefaults()

	env := koyeb.NewDeploymentEnvWithDefaults()
	env.SetKey("FOO")
	env.SetValue("live-value")
	def.SetEnv([]koyeb.DeploymentEnv{*env})

	file := koyeb.NewConfigFileWithDefaults()
	file.SetPath("/etc/data.yaml")
	file.SetContent("live-content")
	file.SetPermissions("0644")
	def.SetConfigFiles([]koyeb.ConfigFile{*file})

	return def
}

func TestParseEnvConfigFilesNoFlagsKeepsLiveValues(t *testing.T) {
	h := &ServiceHandler{}
	flags := envConfigFilesTestFlagSet()

	def := liveEnvConfigFilesDefinition()
	require.NoError(t, h.parseEnvConfigFiles(nil, flags, def))

	env := def.GetEnv()
	require.Len(t, env, 1)
	assert.Equal(t, "FOO", env[0].GetKey())
	assert.Equal(t, "live-value", env[0].GetValue())

	files := def.GetConfigFiles()
	require.Len(t, files, 1)
	assert.Equal(t, "/etc/data.yaml", files[0].GetPath())
	assert.Equal(t, "live-content", files[0].GetContent())
}

func TestParseEnvConfigFilesChangedFlagsMergeLiveValues(t *testing.T) {
	h := &ServiceHandler{}
	flags := envConfigFilesTestFlagSet()

	source := filepath.Join(t.TempDir(), "app.conf")
	require.NoError(t, os.WriteFile(source, []byte("new-content"), 0o600))
	require.NoError(t, flags.Parse([]string{
		"--env", "FOO=updated",
		"--env", "BAR=added",
		"--config-file", source + ":/etc/app.conf:0600",
	}))

	def := liveEnvConfigFilesDefinition()
	require.NoError(t, h.parseEnvConfigFiles(nil, flags, def))

	env := def.GetEnv()
	require.Len(t, env, 2)
	// The merge keeps the live order: an existing key is updated in
	// place, a new key is appended.
	assert.Equal(t, "FOO", env[0].GetKey())
	assert.Equal(t, "updated", env[0].GetValue())
	assert.Equal(t, "BAR", env[1].GetKey())
	assert.Equal(t, "added", env[1].GetValue())

	files := def.GetConfigFiles()
	require.Len(t, files, 2)
	assert.Equal(t, "/etc/data.yaml", files[0].GetPath())
	assert.Equal(t, "live-content", files[0].GetContent(), "an untouched file is kept")
	assert.Equal(t, "/etc/app.conf", files[1].GetPath())
	assert.Equal(t, "new-content", files[1].GetContent())
	assert.Equal(t, "0600", files[1].GetPermissions())
}

func TestParseEnvConfigFilesDeletionIdiom(t *testing.T) {
	h := &ServiceHandler{}
	flags := envConfigFilesTestFlagSet()
	require.NoError(t, flags.Parse([]string{
		"--env", "!FOO",
		"--config-file", "!/etc/data.yaml",
	}))

	def := liveEnvConfigFilesDefinition()
	require.NoError(t, h.parseEnvConfigFiles(nil, flags, def))

	assert.Empty(t, def.GetEnv(), "the deletion idiom must remove the live env variable")
	assert.Empty(t, def.GetConfigFiles(), "the deletion idiom must remove the live config file")
}

func TestParseEnvConfigFilesErrorPaths(t *testing.T) {
	h := &ServiceHandler{}

	t.Run("malformed env value", func(t *testing.T) {
		flags := envConfigFilesTestFlagSet()
		require.NoError(t, flags.Parse([]string{"--env", "MALFORMED"}))

		err := h.parseEnvConfigFiles(nil, flags, koyeb.NewDeploymentDefinitionWithDefaults())
		require.Error(t, err)
		assert.Contains(t, err.Error(), `unable to parse the environment variable "MALFORMED"`)
	})

	t.Run("malformed env deletion value", func(t *testing.T) {
		flags := envConfigFilesTestFlagSet()
		require.NoError(t, flags.Parse([]string{"--env", "!FOO=bar"}))

		err := h.parseEnvConfigFiles(nil, flags, koyeb.NewDeploymentDefinitionWithDefaults())
		require.Error(t, err)
		assert.Contains(t, err.Error(), `unable to parse the environment variable "!FOO=bar"`)
	})

	t.Run("malformed config-file value", func(t *testing.T) {
		flags := envConfigFilesTestFlagSet()
		require.NoError(t, flags.Parse([]string{"--config-file", "no-colon"}))

		err := h.parseEnvConfigFiles(nil, flags, koyeb.NewDeploymentDefinitionWithDefaults())
		require.Error(t, err)
		assert.Contains(t, err.Error(), `unable to parse the confi-file flag value "no-colon"`)
	})
}
