package koyeb

import (
	"context"

	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/idmapper"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/renderer"
	"github.com/spf13/cobra"
)

type ctxkey int

const (
	ctx_client ctxkey = iota
	ctx_mapper
	ctx_renderer
)

// SetupCLIContext is called by the root command to setup the context for all subcommands.
func SetupCLIContext(cmd *cobra.Command) {
	apiClient := getApiClient()
	ctx := cmd.Context()
	ctx = context.WithValue(ctx, koyeb.ContextAccessToken, token)
	ctx = context.WithValue(ctx, ctx_client, apiClient)
	ctx = context.WithValue(ctx, ctx_mapper, idmapper.NewMapper(ctx, apiClient))
	ctx = context.WithValue(ctx, ctx_renderer, renderer.NewRenderer(outputFormat))
	cmd.SetContext(ctx)
}

type CLIContext struct {
	Context  context.Context
	Client   *koyeb.APIClient
	Mapper   *idmapper.Mapper
	Token    string
	Renderer renderer.Renderer
}

// WithCLIContext is a decorator that provides a CLIContext to cobra commands.
func WithCLIContext(fn func(ctx *CLIContext, cmd *cobra.Command, args []string) error) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()
		cliContext := CLIContext{
			Context:  ctx,
			Client:   ctx.Value(ctx_client).(*koyeb.APIClient),
			Mapper:   ctx.Value(ctx_mapper).(*idmapper.Mapper),
			Token:    ctx.Value(koyeb.ContextAccessToken).(string),
			Renderer: ctx.Value(ctx_renderer).(renderer.Renderer),
		}
		return fn(&cliContext, cmd, args)
	}
}
