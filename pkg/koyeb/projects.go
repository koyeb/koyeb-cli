package koyeb

import (
	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
	"github.com/spf13/cobra"
)

func NewProjectCmd() *cobra.Command {
	h := NewProjectHandler()

	projectCmd := &cobra.Command{
		Use:     "projects ACTION",
		Aliases: []string{"proj", "project"},
		Short:   "Projects",
	}

	createProjectCmd := &cobra.Command{
		Use:   "create NAME",
		Short: "Create project",
		Args:  cobra.ExactArgs(1),
		RunE: WithCLIContext(func(ctx *CLIContext, cmd *cobra.Command, args []string) error {
			createProject := koyeb.NewCreateProjectWithDefaults()
			createProject.SetName(args[0])

			if cmd.Flags().Changed("description") {
				createProject.SetDescription(GetStringFlags(cmd, "description"))
			}

			return h.Create(ctx, cmd, args, createProject)
		}),
	}
	createProjectCmd.Flags().String("description", "", "Project description")
	projectCmd.AddCommand(createProjectCmd)

	getProjectCmd := &cobra.Command{
		Use:               "get NAME_OR_ID",
		Short:             "Get project",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: CompleteProjectArgs,
		RunE:              WithCLIContext(h.Get),
	}
	projectCmd.AddCommand(getProjectCmd)

	listProjectCmd := &cobra.Command{
		Use:   "list",
		Short: "List projects",
		RunE:  WithCLIContext(h.List),
	}
	projectCmd.AddCommand(listProjectCmd)

	describeProjectCmd := &cobra.Command{
		Use:               "describe NAME_OR_ID",
		Short:             "Describe project",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: CompleteProjectArgs,
		RunE:              WithCLIContext(h.Describe),
	}
	projectCmd.AddCommand(describeProjectCmd)

	updateProjectCmd := &cobra.Command{
		Use:               "update NAME_OR_ID",
		Short:             "Update project",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: CompleteProjectArgs,
		RunE: WithCLIContext(func(ctx *CLIContext, cmd *cobra.Command, args []string) error {
			updateProject := koyeb.NewProjectWithDefaults()

			if cmd.Flags().Changed("name") {
				updateProject.SetName(GetStringFlags(cmd, "name"))
			}
			if cmd.Flags().Changed("description") {
				updateProject.SetDescription(GetStringFlags(cmd, "description"))
			}

			return h.Update(ctx, cmd, args, updateProject)
		}),
	}
	updateProjectCmd.Flags().StringP("name", "n", "", "Change the name of the project")
	updateProjectCmd.Flags().String("description", "", "Change the project description")
	projectCmd.AddCommand(updateProjectCmd)

	deleteProjectCmd := &cobra.Command{
		Use:               "delete NAME_OR_ID",
		Short:             "Delete project",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: CompleteProjectArgs,
		RunE:              WithCLIContext(h.Delete),
	}
	projectCmd.AddCommand(deleteProjectCmd)

	switchProjectCmd := &cobra.Command{
		Use:               "switch NAME_OR_ID",
		Short:             "Switch the CLI context to another project",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: CompleteProjectArgs,
		RunE:              WithCLIContext(h.Switch),
	}
	projectCmd.AddCommand(switchProjectCmd)

	return projectCmd
}

func NewProjectHandler() *ProjectHandler {
	return &ProjectHandler{}
}

type ProjectHandler struct{}
