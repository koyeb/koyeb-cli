package koyeb

import (
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type completionFunc func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective)

// findCommand resolves a command path (e.g. "services", "scale", "update")
// against the root command, the same way cobra does at execution time.
func findCommand(t *testing.T, path ...string) *cobra.Command {
	t.Helper()
	rootCmd := GetRootCommand()
	cmd, _, err := rootCmd.Find(path)
	require.NoError(t, err, "command not found: %v", path)
	require.NotSame(t, rootCmd, cmd, "command not found: %v", path)
	return cmd
}

// assertValidArgsFunction checks that the command's ValidArgsFunction is the
// expected one: not only that completion is wired, but that the right
// identifiers are suggested for the command's positional arguments.
func assertValidArgsFunction(t *testing.T, cmd *cobra.Command, expected completionFunc) {
	t.Helper()
	if expected == nil {
		assert.Nil(t, cmd.ValidArgsFunction, "unexpected dynamic completion: the argument is not an existing object")
		return
	}
	require.NotNil(t, cmd.ValidArgsFunction, "expected dynamic completion for positional arguments")
	assert.Equal(t,
		reflect.ValueOf(expected).Pointer(),
		reflect.ValueOf(cmd.ValidArgsFunction).Pointer(),
		"wrong completion function wired",
	)
}

// commandsWithCompletion pins, for every command taking existing Koyeb objects
// as positional arguments, the exact completion function wired. Commands
// creating new objects (create, init) or taking no object argument (list) are
// absent.
var commandsWithCompletion = []struct {
	path     []string
	expected completionFunc
}{
	{[]string{"apps", "get"}, completeAppIdentifiers},
	{[]string{"apps", "describe"}, completeAppIdentifiers},
	{[]string{"apps", "update"}, completeAppIdentifiers},
	{[]string{"apps", "delete"}, completeAppIdentifiers},
	{[]string{"apps", "pause"}, completeAppIdentifiers},
	{[]string{"apps", "resume"}, completeAppIdentifiers},

	{[]string{"deployments", "get"}, completeDeploymentIdentifiers},
	{[]string{"deployments", "describe"}, completeDeploymentIdentifiers},
	{[]string{"deployments", "cancel"}, completeDeploymentIdentifiers},
	{[]string{"deployments", "logs"}, completeDeploymentIdentifiers},

	{[]string{"services", "get"}, completeServiceIdentifiers},
	{[]string{"services", "unapplied-changes"}, completeServiceIdentifiers},
	{[]string{"services", "logs"}, completeServiceIdentifiers},
	{[]string{"services", "describe"}, completeServiceIdentifiers},
	{[]string{"services", "exec"}, completeServiceExecArgs},
	{[]string{"services", "update"}, completeServiceIdentifiers},
	{[]string{"services", "redeploy"}, completeServiceIdentifiers},
	{[]string{"services", "delete"}, completeServiceIdentifiers},
	{[]string{"services", "pause"}, completeServiceIdentifiers},
	{[]string{"services", "resume"}, completeServiceIdentifiers},
	{[]string{"services", "scale"}, completeServiceIdentifiers},
	{[]string{"services", "scale", "update"}, completeServiceIdentifiers},
	{[]string{"services", "scale", "get"}, completeServiceIdentifiers},
	{[]string{"services", "scale", "delete"}, completeServiceIdentifiers},

	{[]string{"instances", "get"}, completeInstanceIdentifiers},
	{[]string{"instances", "describe"}, completeInstanceIdentifiers},
	{[]string{"instances", "exec"}, completeInstanceExecArgs},
	{[]string{"instances", "logs"}, completeInstanceIdentifiers},

	{[]string{"domains", "get"}, completeDomainIdentifiers},
	{[]string{"domains", "describe"}, completeDomainIdentifiers},
	{[]string{"domains", "delete"}, completeDomainIdentifiers},
	{[]string{"domains", "refresh"}, completeDomainIdentifiers},
	{[]string{"domains", "attach"}, completeDomainAttachArgs},
	{[]string{"domains", "detach"}, completeDomainIdentifiers},

	{[]string{"secrets", "get"}, completeSecretIdentifiers},
	{[]string{"secrets", "describe"}, completeSecretIdentifiers},
	{[]string{"secrets", "update"}, completeSecretIdentifiers},
	{[]string{"secrets", "delete"}, completeSecretIdentifiers},
	{[]string{"secrets", "reveal"}, completeSecretIdentifiers},

	{[]string{"databases", "get"}, completeDatabaseIdentifiers},
	{[]string{"databases", "update"}, completeDatabaseIdentifiers},
	{[]string{"databases", "delete"}, completeDatabaseIdentifiers},

	{[]string{"volumes", "get"}, completeVolumeIdentifiers},
	{[]string{"volumes", "update"}, completeVolumeIdentifiers},
	{[]string{"volumes", "delete"}, completeVolumeIdentifiers},

	{[]string{"snapshots", "create"}, completeSnapshotCreateArgs},
	{[]string{"snapshots", "get"}, completeSnapshotIdentifiers},
	{[]string{"snapshots", "update"}, completeSnapshotIdentifiers},
	{[]string{"snapshots", "delete"}, completeSnapshotIdentifiers},

	{[]string{"regional-deployments", "get"}, completeRegionalDeploymentIdentifiers},

	{[]string{"organizations", "switch"}, completeOrganizationIdentifiers},

	{[]string{"pool", "get"}, completePoolIdentifiers},
	{[]string{"pool", "describe"}, completePoolIdentifiers},
	{[]string{"pool", "delete"}, completePoolIdentifiers},
	{[]string{"pool", "update"}, completePoolIdentifiers},
	{[]string{"pool", "claim"}, completePoolIdentifiers},
}

// commandsWithoutCompletion lists commands taking a NAME argument that is a
// new object to create, a file path, or no object argument at all: they must
// not offer dynamic completion.
var commandsWithoutCompletion = [][]string{
	{"apps", "create"},
	{"apps", "init"},
	{"apps", "list"},
	{"deployments", "list"},
	{"services", "create"},
	{"services", "list"},
	{"instances", "list"},
	{"instances", "cp"},
	{"domains", "create"},
	{"domains", "list"},
	{"secrets", "create"},
	{"secrets", "list"},
	{"databases", "create"},
	{"databases", "list"},
	{"volumes", "create"},
	{"volumes", "list"},
	{"snapshots", "list"},
	{"pool", "list"},
}

func TestCompletionWiring(t *testing.T) {
	for _, c := range commandsWithCompletion {
		t.Run("with completion: "+strings.Join(c.path, " "), func(t *testing.T) {
			assertValidArgsFunction(t, findCommand(t, c.path...), c.expected)
		})
	}
	for _, path := range commandsWithoutCompletion {
		t.Run("without completion: "+strings.Join(path, " "), func(t *testing.T) {
			assertValidArgsFunction(t, findCommand(t, path...), nil)
		})
	}
}

// scopedFlagDeclarations returns, for every flag name in
// scopedFlagCompletions, the paths of the commands declaring the flag
// locally. This is exactly the registration surface of
// registerScopedFlagCompletions: inherited persistent flags resolve to the
// declaring command's registration, as cobra keys completions by flag pointer.
func scopedFlagDeclarations() map[string][]string {
	found := map[string][]string{}
	var walk func(cmd *cobra.Command, path string)
	walk = func(cmd *cobra.Command, path string) {
		if path != "" {
			path += " "
		}
		path += cmd.Name()
		for name := range scopedFlagCompletions {
			if cmd.LocalFlags().Lookup(name) != nil {
				found[name] = append(found[name], path)
			}
		}
		for _, subcmd := range cmd.Commands() {
			walk(subcmd, path)
		}
	}
	walk(GetRootCommand(), "")
	return found
}

// expectedScopedFlagDeclarations pins the registration surface: every command
// declaring a flag whose value is a Koyeb object identifier. When a new
// command or flag is added to the tree, this list must be extended
// consciously, together with scopedFlagCompletions.
var expectedScopedFlagDeclarations = map[string][]string{
	"project": {
		"koyeb apps", "koyeb archives", "koyeb compose", "koyeb deploy", "koyeb deployments",
		"koyeb domains", "koyeb instances", "koyeb metrics", "koyeb organizations", "koyeb pool",
		"koyeb regional-deployments", "koyeb sandbox", "koyeb services", "koyeb snapshots", "koyeb volumes",
	},
	"workspace": {
		"koyeb apps", "koyeb archives", "koyeb compose", "koyeb deploy", "koyeb deployments",
		"koyeb domains", "koyeb instances", "koyeb metrics", "koyeb organizations", "koyeb pool",
		"koyeb regional-deployments", "koyeb sandbox", "koyeb services", "koyeb snapshots", "koyeb volumes",
	},
	"organization": {"koyeb"},
	"app": {
		"koyeb databases create", "koyeb databases get", "koyeb databases update", "koyeb deploy",
		"koyeb deployments list", "koyeb instances list", "koyeb sandbox create", "koyeb sandbox list",
		"koyeb services create", "koyeb services delete", "koyeb services describe", "koyeb services exec",
		"koyeb services get", "koyeb services list", "koyeb services logs", "koyeb services pause",
		"koyeb services redeploy", "koyeb services resume", "koyeb services scale",
		"koyeb services scale delete", "koyeb services scale get", "koyeb services scale update",
		"koyeb services unapplied-changes", "koyeb services update",
	},
	"service":    {"koyeb deployments list", "koyeb instances list", "koyeb metrics get"},
	"instance":   {"koyeb metrics get", "koyeb services logs"},
	"attach-to":  {"koyeb domains create"},
	"snapshot":   {"koyeb sandbox create", "koyeb volumes create"},
	"deployment": {"koyeb regional-deployments list"},
}

func TestScopedFlagCompletions(t *testing.T) {
	found := scopedFlagDeclarations()

	require.Equal(t, len(expectedScopedFlagDeclarations), len(found),
		"scopedFlagCompletions and expectedScopedFlagDeclarations disagree on flag names")

	for name, expectedCommands := range expectedScopedFlagDeclarations {
		sort.Strings(expectedCommands)
		sort.Strings(found[name])
		assert.Equal(t, expectedCommands, found[name],
			"unexpected registration surface for flag --%s", name)
	}
}
