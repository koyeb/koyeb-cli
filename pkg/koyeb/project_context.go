package koyeb

import (
	"context"
	"strings"

	apiv1 "github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/idmapper"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func ResolveProjectArgs(ctx *CLIContext, val string) (string, error) {
	return ctx.Mapper.Project().ResolveID(val)
}

func CompleteProjectArgs(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	mapper, err := newProjectCompletionMapper(cmd)
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	names, err := mapper.Names()
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	completions := make([]string, 0, len(names))
	for _, name := range names {
		if strings.HasPrefix(name, toComplete) {
			completions = append(completions, name)
		}
	}

	return completions, cobra.ShellCompDirectiveNoFileComp
}

func SwitchProjectConfig(project string) error {
	// TODO: Persist this through Organization.default_project once the API schema
	// exposes it in the generated client. The pinned client currently has no
	// default_project field on Organization, so the CLI stores the active project
	// locally.
	viper.Set("project", project)
	return viper.WriteConfig()
}

func ClearProjectConfig() error {
	viper.Set("project", "")
	return viper.WriteConfig()
}

type projectSetter interface {
	SetProjectId(string)
}

func applyProjectID(payload projectSetter, project string) {
	if project != "" {
		payload.SetProjectId(project)
	}
}

func newProjectCompletionMapper(cmd *cobra.Command) (*idmapper.ProjectMapper, error) {
	if err := initConfig(cmd.Root()); err != nil {
		return nil, err
	}

	apiClient, err := getApiClient()
	if err != nil {
		return nil, err
	}

	ctx := context.Background()
	ctx = context.WithValue(ctx, apiv1.ContextAccessToken, token)

	if organization != "" {
		token, err := GetOrganizationToken(apiClient.OrganizationApi, ctx, organization)
		if err != nil {
			return nil, err
		}
		ctx = context.WithValue(ctx, apiv1.ContextAccessToken, token)
	}

	return idmapper.NewProjectMapper(ctx, apiClient), nil
}
