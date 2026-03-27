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
	ctx_logs_client
	ctx_exec_client
	ctx_mapper
	ctx_renderer
	ctx_organization
	ctx_project
)

func topLevelCommand(cmd *cobra.Command) *cobra.Command {
	current := cmd
	for current.Parent() != nil && current.Parent().Parent() != nil {
		current = current.Parent()
	}
	return current
}

// SetupCLIContext is called by the root command to setup the context for all subcommands.
// When `organization` is not empty, it should contain the ID of the organization to switch the context to.
func SetupCLIContext(cmd *cobra.Command, organization string, project string) error {
	apiClient, err := getApiClient()
	if err != nil {
		return err
	}

	ctx := cmd.Context()
	ctx = context.WithValue(ctx, koyeb.ContextAccessToken, token)

	if organization != "" {
		token, err := GetOrganizationToken(apiClient.OrganizationApi, ctx, organization)
		if err != nil {
			return err
		}
		ctx = context.WithValue(ctx, koyeb.ContextAccessToken, token)
		// Update command context with the organization token. This is required
		// because the idmapper initialization below will use the token from the
		// context.
		cmd.SetContext(ctx)
	}

	activeProject := ""
	topLevel := topLevelCommand(cmd)
	if flag := cmd.Flags().Lookup("all-projects"); flag != nil {
		allProjects, err := cmd.Flags().GetBool("all-projects")
		if err != nil {
			return err
		}
		if allProjects {
			project = ""
		}
	}
	if topLevel.Name() == "organizations" || topLevel.Name() == "projects" {
		project = ""
	}
	if project == "" {
		project, err = getOrganizationDefaultProjectID(&CLIContext{
			Context:      ctx,
			Client:       apiClient,
			Token:        ctx.Value(koyeb.ContextAccessToken).(string),
			Organization: organization,
		})
		if err != nil {
			return err
		}
	}
	if project != "" {
		projectMapper := idmapper.NewProjectMapper(ctx, apiClient)
		activeProject, err = projectMapper.ResolveID(project)
		if err != nil {
			return err
		}
	}

	ctx = context.WithValue(ctx, ctx_client, apiClient)

	logsApiClient, err := NewLogsAPIClient(apiClient, apiurl, ctx.Value(koyeb.ContextAccessToken).(string))
	if err != nil {
		return err
	}
	ctx = context.WithValue(ctx, ctx_logs_client, logsApiClient)

	execApiClient, err := NewExecAPIClient(apiurl, ctx.Value(koyeb.ContextAccessToken).(string))
	if err != nil {
		return err
	}
	ctx = context.WithValue(ctx, ctx_exec_client, execApiClient)

	ctx = context.WithValue(ctx, ctx_mapper, idmapper.NewMapper(ctx, apiClient, activeProject))
	ctx = context.WithValue(ctx, ctx_renderer, renderer.NewRenderer(outputFormat))
	ctx = context.WithValue(ctx, ctx_organization, organization)
	ctx = context.WithValue(ctx, ctx_project, activeProject)
	cmd.SetContext(ctx)

	return nil
}

type CLIContext struct {
	Context      context.Context
	Client       *koyeb.APIClient
	LogsClient   *LogsAPIClient
	ExecClient   *ExecAPIClient
	Mapper       *idmapper.Mapper
	Token        string
	Renderer     renderer.Renderer
	Organization string
	Project      string
}

// GetCLIContext transforms the untyped context passed to cobra commands into a CLIContext.
func GetCLIContext(ctx context.Context) *CLIContext {
	return &CLIContext{
		Context:      ctx,
		Client:       ctx.Value(ctx_client).(*koyeb.APIClient),
		LogsClient:   ctx.Value(ctx_logs_client).(*LogsAPIClient),
		ExecClient:   ctx.Value(ctx_exec_client).(*ExecAPIClient),
		Mapper:       ctx.Value(ctx_mapper).(*idmapper.Mapper),
		Token:        ctx.Value(koyeb.ContextAccessToken).(string),
		Renderer:     ctx.Value(ctx_renderer).(renderer.Renderer),
		Organization: ctx.Value(ctx_organization).(string),
		Project:      ctx.Value(ctx_project).(string),
	}
}

// WithCLIContext is a decorator that provides a CLIContext to cobra commands.
func WithCLIContext(fn func(ctx *CLIContext, cmd *cobra.Command, args []string) error) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		return fn(GetCLIContext(cmd.Context()), cmd, args)
	}
}
