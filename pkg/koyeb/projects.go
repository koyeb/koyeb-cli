package koyeb

import (
	"fmt"

	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func NewProjectCmd() *cobra.Command {
	h := NewProjectHandler()

	projectCmd := &cobra.Command{
		Use:     "projects ACTION",
		Aliases: []string{"project", "projs", "proj"},
		Short:   "Projects",
	}

	createProjectCmd := &cobra.Command{
		Use:   "create NAME",
		Short: "Create project",
		Args:  cobra.ExactArgs(1),
		RunE: WithCLIContext(func(ctx *CLIContext, cmd *cobra.Command, args []string) error {
			createProject := koyeb.NewCreateProjectWithDefaults()
			SyncFlags(cmd, args, createProject)
			return h.Create(ctx, cmd, args, createProject)
		}),
	}
	createProjectCmd.Flags().String("description", "", "Project description")
	projectCmd.AddCommand(createProjectCmd)

	getProjectCmd := &cobra.Command{
		Use:   "get NAME",
		Short: "Get project",
		Args:  cobra.ExactArgs(1),
		RunE:  WithCLIContext(h.Get),
	}
	projectCmd.AddCommand(getProjectCmd)

	listProjectCmd := &cobra.Command{
		Use:   "list",
		Short: "List projects",
		RunE:  WithCLIContext(h.List),
	}
	projectCmd.AddCommand(listProjectCmd)

	describeProjectCmd := &cobra.Command{
		Use:   "describe NAME",
		Short: "Describe project",
		Args:  cobra.ExactArgs(1),
		RunE:  WithCLIContext(h.Describe),
	}
	projectCmd.AddCommand(describeProjectCmd)

	updateProjectCmd := &cobra.Command{
		Use:   "update NAME",
		Short: "Update project",
		Args:  cobra.ExactArgs(1),
		RunE: WithCLIContext(func(ctx *CLIContext, cmd *cobra.Command, args []string) error {
			updateProject := koyeb.NewProjectWithDefaults()
			SyncFlags(cmd, args, updateProject)
			return h.Update(ctx, cmd, args, updateProject)
		}),
	}
	updateProjectCmd.Flags().StringP("name", "n", "", "Project name")
	updateProjectCmd.Flags().String("description", "", "Project description")
	projectCmd.AddCommand(updateProjectCmd)

	deleteProjectCmd := &cobra.Command{
		Use:   "delete NAME",
		Short: "Delete project",
		Args:  cobra.ExactArgs(1),
		RunE:  WithCLIContext(h.Delete),
	}
	projectCmd.AddCommand(deleteProjectCmd)

	switchProjectCmd := &cobra.Command{
		Use:   "switch NAME",
		Short: "Switch the default project for the CLI context",
		Args:  cobra.ExactArgs(1),
		RunE:  WithCLIContext(h.Switch),
	}
	projectCmd.AddCommand(switchProjectCmd)

	return projectCmd
}

func NewProjectHandler() *ProjectHandler {
	return &ProjectHandler{}
}

type ProjectHandler struct {
}

func (h *ProjectHandler) ResolveProjectArgs(ctx *CLIContext, val string) (string, error) {
	projectMapper := ctx.Mapper.Project()
	id, err := projectMapper.ResolveID(val)
	if err != nil {
		return "", err
	}
	return id, nil
}

func (h *ProjectHandler) CreateProject(ctx *CLIContext, payload *koyeb.CreateProject) (*koyeb.CreateProjectReply, error) {
	res, resp, err := ctx.Client.ProjectsApi.CreateProject(ctx.Context).Project(*payload).Execute()
	if err != nil {
		return nil, errors.NewCLIErrorFromAPIError(
			fmt.Sprintf("Error while creating the project `%s`", payload.GetName()),
			err,
			resp,
		)
	}
	return res, nil
}

func (h *ProjectHandler) Switch(ctx *CLIContext, cmd *cobra.Command, args []string) error {
	project, err := h.ResolveProjectArgs(ctx, args[0])
	if err != nil {
		return err
	}
	viper.Set("project", project)
	if err := viper.WriteConfig(); err != nil {
		return &errors.CLIError{
			What: "Unable to switch the current project",
			Why:  "we were unable to write the configuration file",
			Additional: []string{
				"The command `koyeb project switch` needs to update your configuration file, usually located in $HOME/.koyeb.yaml",
				"If you do not have write access to this file, you can use the --config flag to specify a different location.",
				"Alternatively, you can manually edit the configuration file and set the project field to the project ID you want to use.",
				"You can also provide the project UUID with the --project flag.",
			},
			Orig:     err,
			Solution: "Fix the issue preventing the CLI to write the configuration file, or manually edit the configuration file",
		}
	}
	return nil
}
