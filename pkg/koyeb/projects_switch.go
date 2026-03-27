package koyeb

import (
	"github.com/spf13/cobra"
)

func (h *ProjectHandler) Switch(ctx *CLIContext, cmd *cobra.Command, args []string) error {
	project, err := ResolveProjectArgs(ctx, args[0])
	if err != nil {
		return err
	}

	return updateOrganizationDefaultProjectID(ctx, &project)
}
