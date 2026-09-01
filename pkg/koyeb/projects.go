package koyeb

import (
	"github.com/spf13/cobra"
)

func NewProjectCmd() *cobra.Command {
	h := NewProjectHandler()

	projectCmd := &cobra.Command{
		Use:     "projects ACTION",
		Aliases: []string{"project"},
		Short:   "Projects",
	}

	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List projects",
		RunE:  WithCLIContext(h.List),
	}
	projectCmd.AddCommand(listCmd)

	return projectCmd
}

func NewProjectHandler() *ProjectHandler {
	return &ProjectHandler{}
}

type ProjectHandler struct{}

func (h *ProjectHandler) ResolveProjectArgs(ctx *CLIContext, val string) (string, error) {
	projectMapper := ctx.Mapper.Project()
	id, err := projectMapper.ResolveID(val)
	if err != nil {
		return "", err
	}
	return id, nil
}

// setProjectHeader resolves the --project flag (if set) and sets the
// x-koyeb-project-id header on the API client config so all subsequent
// requests are scoped to that project.
func setProjectHeader(ctx *CLIContext, cmd *cobra.Command) error {
	projectFlag := GetStringFlags(cmd, "project")
	if projectFlag == "" {
		return nil
	}
	projectHandler := NewProjectHandler()
	projectId, err := projectHandler.ResolveProjectArgs(ctx, projectFlag)
	if err != nil {
		return err
	}
	ctx.Client.GetConfig().AddDefaultHeader("x-koyeb-project-id", projectId)
	return nil
}
