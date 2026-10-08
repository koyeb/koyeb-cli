package koyeb

import (
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// ensureCompletionContext rebuilds the CLI context when the command line being
// completed overrides the authentication-related flags (--organization,
// --token, --url). Cobra parses these flags on the final command only, after
// the root's PersistentPreRunE has already set up the context from the
// configuration file, so the override would otherwise be ignored. The flags
// are looked up by Changed rather than by value: an absent flag has its
// default value, which must not be mistaken for an override.
func ensureCompletionContext(cmd *cobra.Command) error {
	orgFlag := cmd.Flags().Lookup("organization")
	tokenFlag := cmd.Flags().Lookup("token")
	urlFlag := cmd.Flags().Lookup("url")
	tokenChanged := tokenFlag != nil && tokenFlag.Changed
	urlChanged := urlFlag != nil && urlFlag.Changed
	if (orgFlag == nil || !orgFlag.Changed) && !tokenChanged && !urlChanged {
		return nil
	}
	// A --organization value equal to the one the context was already built
	// with (from the configuration file) is not an override: rebuilding
	// would only repeat the organization token exchange.
	if !tokenChanged && !urlChanged && orgFlag != nil &&
		orgFlag.Value.String() == GetCLIContext(cmd.Context()).Organization {
		return nil
	}
	// Refresh the credentials from the parsed flags: viper resolves the values
	// of bound flags once they are marked as changed.
	token = viper.GetString("token")
	apiurl = viper.GetString("url")
	return SetupCLIContext(cmd, GetStringFlags(cmd, "organization"))
}

// completeIdentifiers returns a cobra completion function suggesting the
// identifiers (names, short IDs, slugs) of Koyeb objects. The identifiers are
// listed from the API, filtered by the prefix being completed, and file
// completion is disabled.
//
// Cobra does not prefix-filter the results of ValidArgsFunction and flag
// completion functions, so the filtering is done here.
func completeIdentifiers(listIdentifiers func(ctx *CLIContext) ([]string, error)) func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	return func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return completeIdentifiersWithScope(cmd, toComplete, listIdentifiers, false)
	}
}

// completeScopedIdentifiers is completeIdentifiers for commands whose objects
// are scoped by --project/--workspace: the flag is resolved and applied to the
// API client before listing, the same way WithCLIContext does for command
// execution. It must not be used to complete the value of --project/--workspace
// itself, as the value being completed is only partially typed.
func completeScopedIdentifiers(listIdentifiers func(ctx *CLIContext) ([]string, error)) func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	return func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return completeIdentifiersWithScope(cmd, toComplete, listIdentifiers, true)
	}
}

// completeIdentifiersWithScope is the shared implementation of the completion
// functions above. The context is ensured exactly once per completion request,
// then the --project/--workspace scoping is applied if requested, and finally
// the identifiers are listed and filtered by the prefix being completed.
func completeIdentifiersWithScope(cmd *cobra.Command, toComplete string, listIdentifiers func(ctx *CLIContext) ([]string, error), scoped bool) ([]string, cobra.ShellCompDirective) {
	if err := ensureCompletionContext(cmd); err != nil {
		return nil, cobra.ShellCompDirectiveError
	}
	ctx := GetCLIContext(cmd.Context())
	if scoped {
		if err := setProjectHeader(ctx, cmd); err != nil {
			return nil, cobra.ShellCompDirectiveError
		}
	}
	identifiers, err := listIdentifiers(ctx)
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}
	return filterCompletions(identifiers, toComplete), cobra.ShellCompDirectiveNoFileComp
}

func filterCompletions(values []string, toComplete string) []string {
	var filtered []string
	for _, value := range values {
		// Skip empty values: they would produce empty lines, which the
		// completion protocol cannot represent.
		if value == "" {
			continue
		}
		if strings.HasPrefix(value, toComplete) {
			filtered = append(filtered, value)
		}
	}
	return filtered
}

func completeAppIdentifiers(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	return completeScopedIdentifiers(func(ctx *CLIContext) ([]string, error) {
		return ctx.Mapper.App().Complete()
	})(cmd, args, toComplete)
}

func completeDeploymentIdentifiers(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	return completeScopedIdentifiers(func(ctx *CLIContext) ([]string, error) {
		return ctx.Mapper.Deployment().Complete()
	})(cmd, args, toComplete)
}

func completeServiceIdentifiers(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	return completeScopedIdentifiers(func(ctx *CLIContext) ([]string, error) {
		return ctx.Mapper.Service().Complete()
	})(cmd, args, toComplete)
}

func completeProjectIdentifiers(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	return completeIdentifiers(func(ctx *CLIContext) ([]string, error) {
		return ctx.Mapper.Project().Complete()
	})(cmd, args, toComplete)
}

func completeOrganizationIdentifiers(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	return completeIdentifiers(func(ctx *CLIContext) ([]string, error) {
		return ctx.Mapper.Organization().Complete()
	})(cmd, args, toComplete)
}

func completeInstanceIdentifiers(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	return completeScopedIdentifiers(func(ctx *CLIContext) ([]string, error) {
		return ctx.Mapper.Instance().Complete()
	})(cmd, args, toComplete)
}

func completeDomainIdentifiers(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	return completeScopedIdentifiers(func(ctx *CLIContext) ([]string, error) {
		return ctx.Mapper.Domain().Complete()
	})(cmd, args, toComplete)
}

func completeSecretIdentifiers(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	return completeScopedIdentifiers(func(ctx *CLIContext) ([]string, error) {
		return ctx.Mapper.Secret().Complete()
	})(cmd, args, toComplete)
}

func completeDatabaseIdentifiers(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	return completeScopedIdentifiers(func(ctx *CLIContext) ([]string, error) {
		return ctx.Mapper.Database().Complete()
	})(cmd, args, toComplete)
}

func completeVolumeIdentifiers(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	return completeScopedIdentifiers(func(ctx *CLIContext) ([]string, error) {
		return ctx.Mapper.Volume().Complete()
	})(cmd, args, toComplete)
}

func completeSnapshotIdentifiers(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	return completeScopedIdentifiers(func(ctx *CLIContext) ([]string, error) {
		return ctx.Mapper.Snapshot().Complete()
	})(cmd, args, toComplete)
}

func completePoolIdentifiers(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	return completeScopedIdentifiers(func(ctx *CLIContext) ([]string, error) {
		return ctx.Mapper.Pool().Complete()
	})(cmd, args, toComplete)
}

func completeRegionalDeploymentIdentifiers(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	return completeScopedIdentifiers(func(ctx *CLIContext) ([]string, error) {
		return ctx.Mapper.RegionalDeployment().Complete()
	})(cmd, args, toComplete)
}

// completeInstanceExecArgs completes the first positional argument of
// "instances exec NAME CMD -- [args...]": the instance. The following arguments
// are the command to run in the instance, there is nothing to complete.
func completeInstanceExecArgs(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	if len(args) == 0 {
		return completeInstanceIdentifiers(cmd, args, toComplete)
	}
	return nil, cobra.ShellCompDirectiveNoFileComp
}

// completeServiceExecArgs completes the first positional argument of
// "services exec NAME CMD -- [args...]": the service. The following arguments
// are the command to run in the instance, there is nothing to complete.
func completeServiceExecArgs(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	if len(args) == 0 {
		return completeServiceIdentifiers(cmd, args, toComplete)
	}
	return nil, cobra.ShellCompDirectiveNoFileComp
}

// completeDomainAttachArgs completes the positional arguments of
// "domains attach NAME APP": a domain for the first argument, an app for the
// second.
func completeDomainAttachArgs(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	switch len(args) {
	case 0:
		return completeDomainIdentifiers(cmd, args, toComplete)
	case 1:
		return completeAppIdentifiers(cmd, args, toComplete)
	}
	return nil, cobra.ShellCompDirectiveNoFileComp
}

// completeSnapshotCreateArgs completes the positional arguments of
// "snapshots create NAME PARENT_VOLUME": the first argument is the name of the
// new snapshot (nothing to complete), the second is the parent volume.
func completeSnapshotCreateArgs(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	if len(args) == 1 {
		return completeVolumeIdentifiers(cmd, args, toComplete)
	}
	return nil, cobra.ShellCompDirectiveNoFileComp
}

// scopedFlagCompletions maps the names of the flags whose values are Koyeb
// object identifiers to their completion functions.
var scopedFlagCompletions = map[string]func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective){
	"project":      completeProjectIdentifiers,
	"workspace":    completeProjectIdentifiers,
	"organization": completeOrganizationIdentifiers,
	"app":          completeAppIdentifiers,
	"service":      completeServiceIdentifiers,
	"instance":     completeInstanceIdentifiers,
	"attach-to":    completeAppIdentifiers,
	"snapshot":     completeSnapshotIdentifiers,
	"deployment":   completeDeploymentIdentifiers,
}

// registerScopedFlagCompletions registers shell completion for the flags whose
// values are Koyeb object identifiers, on every command of the tree that
// declares them. Registration is keyed by flag pointer in cobra, so only the
// commands declaring the flag locally register a completion function; commands
// inheriting a persistent flag from a parent resolve to the same pointer.
func registerScopedFlagCompletions(cmd *cobra.Command) {
	for _, subcmd := range cmd.Commands() {
		registerScopedFlagCompletions(subcmd)
	}
	for name, completion := range scopedFlagCompletions {
		if cmd.LocalFlags().Lookup(name) != nil {
			cmd.RegisterFlagCompletionFunc(name, completion) //nolint:errcheck
		}
	}
}
