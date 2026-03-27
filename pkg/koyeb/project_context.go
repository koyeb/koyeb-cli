package koyeb

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	apiv1 "github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/errors"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/idmapper"
	"github.com/spf13/cobra"
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

func getCurrentOrganizationID(ctx *CLIContext) (string, error) {
	if ctx.Organization != "" {
		return ctx.Organization, nil
	}

	orgRes, resp, err := ctx.Client.ProfileApi.GetCurrentOrganization(ctx.Context).Execute()
	if err == nil {
		org := orgRes.GetOrganization()
		return org.GetId(), nil
	}
	if resp != nil && resp.StatusCode != http.StatusUnauthorized {
		return "", errors.NewCLIErrorFromAPIError("Unable to fetch the current organization", err, resp)
	}

	appsRes, appResp, appErr := ctx.Client.AppsApi.ListApps(ctx.Context).Limit("1").Execute()
	if appErr == nil {
		apps := appsRes.GetApps()
		if len(apps) > 0 && apps[0].GetOrganizationId() != "" {
			return apps[0].GetOrganizationId(), nil
		}
	} else if appResp != nil && appResp.StatusCode != http.StatusUnauthorized {
		return "", errors.NewCLIErrorFromAPIError("Unable to discover the current organization", appErr, appResp)
	}

	projectsRes, projectResp, projectErr := ctx.Client.ProjectsApi.ListProjects(ctx.Context).Limit("1").Execute()
	if projectErr == nil {
		projects := projectsRes.GetProjects()
		if len(projects) > 0 && projects[0].GetOrganizationId() != "" {
			return projects[0].GetOrganizationId(), nil
		}
	} else if projectResp != nil && projectResp.StatusCode != http.StatusUnauthorized {
		return "", errors.NewCLIErrorFromAPIError("Unable to discover the current organization", projectErr, projectResp)
	}

	return "", &errors.CLIError{
		What: "Unable to determine the current organization",
		Why:  "the CLI could not infer which organization is currently active",
		Additional: []string{
			"Specify the organization explicitly with --organization, or switch to one with `koyeb organizations switch`.",
		},
		Solution: errors.SolutionFixConfig,
	}
}

type rawOrganizationReply struct {
	Organization rawOrganization `json:"organization"`
}

type rawOrganization struct {
	Id               string  `json:"id,omitempty"`
	DefaultProjectID *string `json:"default_project_id,omitempty"`
}

func getOrganizationDefaultProjectID(ctx *CLIContext) (string, error) {
	orgID, err := getCurrentOrganizationID(ctx)
	if err != nil {
		return "", err
	}

	reply, err := getRawOrganization(ctx, orgID)
	if err != nil {
		return "", err
	}

	if reply.Organization.DefaultProjectID == nil {
		return "", nil
	}
	return *reply.Organization.DefaultProjectID, nil
}

func updateOrganizationDefaultProjectID(ctx *CLIContext, projectID *string) error {
	orgID, err := getCurrentOrganizationID(ctx)
	if err != nil {
		return err
	}

	payload, err := json.Marshal(map[string]*string{
		"default_project_id": projectID,
	})
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(
		ctx.Context,
		http.MethodPatch,
		fmt.Sprintf("%s/v1/organizations/%s?update_mask=default_project_id", strings.TrimRight(apiurl, "/"), orgID),
		bytes.NewReader(payload),
	)
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Bearer "+ctx.Token)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "*/*")

	resp, err := (&http.Client{Transport: &DebugTransport{http.DefaultTransport}}).Do(req)
	if err != nil {
		return &errors.CLIError{
			What:     "Unable to update the default project of the organization",
			Why:      "the CLI was unable to query the Koyeb API",
			Orig:     err,
			Solution: errors.SolutionFixConfig,
		}
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return &errors.CLIError{
			What: "Unable to update the default project of the organization",
			Why:  fmt.Sprintf("the Koyeb API returned HTTP/%d", resp.StatusCode),
			Additional: []string{
				strings.TrimSpace(string(body)),
			},
			Solution: errors.SolutionTryAgainOrUpdateOrIssue,
		}
	}

	return nil
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

func getRawOrganization(ctx *CLIContext, orgID string) (*rawOrganizationReply, error) {
	req, err := http.NewRequestWithContext(
		ctx.Context,
		http.MethodGet,
		fmt.Sprintf("%s/v1/organizations/%s", strings.TrimRight(apiurl, "/"), orgID),
		nil,
	)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+ctx.Token)
	req.Header.Set("Accept", "*/*")

	resp, err := (&http.Client{Transport: &DebugTransport{http.DefaultTransport}}).Do(req)
	if err != nil {
		return nil, &errors.CLIError{
			What:     "Unable to fetch the current organization",
			Why:      "the CLI was unable to query the Koyeb API",
			Orig:     err,
			Solution: errors.SolutionFixConfig,
		}
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode >= 300 {
		return nil, &errors.CLIError{
			What: "Unable to fetch the current organization",
			Why:  fmt.Sprintf("the Koyeb API returned HTTP/%d", resp.StatusCode),
			Additional: []string{
				strings.TrimSpace(string(body)),
			},
			Solution: errors.SolutionTryAgainOrUpdateOrIssue,
		}
	}

	var reply rawOrganizationReply
	if err := json.Unmarshal(body, &reply); err != nil {
		return nil, err
	}

	return &reply, nil
}
