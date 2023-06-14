package koyeb

import (
	"context"

	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/idmapper"
	"github.com/spf13/cobra"
)

// SetupCLIContext is called by the root command to setup the context for all subcommands.
func SetupCLIContext(cmd *cobra.Command) {
	apiClient := getApiClient()
	ctx := cmd.Context()
	ctx = context.WithValue(ctx, koyeb.ContextAccessToken, token)
	ctx = context.WithValue(ctx, "client", apiClient)
	ctx = context.WithValue(ctx, "mapper", idmapper.NewMapper(ctx, apiClient))
	cmd.SetContext(ctx)
}

type CLIContext struct {
	context context.Context
	client  *koyeb.APIClient
	mapper  *idmapper.Mapper
	token   string
}

// WithCLIContext is a decorator that provides a CLIContext to cobra commands.
func WithCLIContext(fn func(ctx *CLIContext, cmd *cobra.Command, args []string) error) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()
		cliContext := CLIContext{
			context: ctx,
			client:  ctx.Value("client").(*koyeb.APIClient),
			mapper:  ctx.Value("mapper").(*idmapper.Mapper),
			token:   ctx.Value(koyeb.ContextAccessToken).(string),
		}
		return fn(&cliContext, cmd, args)
	}
}
