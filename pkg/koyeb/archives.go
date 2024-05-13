package koyeb

import (
	"github.com/spf13/cobra"
)

func NewArchiveCmd() *cobra.Command {
	h := NewArchiveHandler()

	archiveCmd := &cobra.Command{
		Use:     "archives ACTION",
		Aliases: []string{"archive"},
		Short:   "Archives",
	}

	createArchiveCmd := &cobra.Command{
		Use:   "create NAME",
		Short: "Create archive",
		Args:  cobra.ExactArgs(1),
		RunE: WithCLIContext(func(ctx *CLIContext, cmd *cobra.Command, args []string) error {
			return h.Create(ctx, cmd, args[0])
		}),
	}
	archiveCmd.AddCommand(createArchiveCmd)
	return archiveCmd
}

func NewArchiveHandler() *ArchiveHandler {
	return &ArchiveHandler{}
}

type ArchiveHandler struct {
}
