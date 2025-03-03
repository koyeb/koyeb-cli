package koyeb

import (
	"github.com/spf13/cobra"

	"github.com/spf13/pflag"
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
			err := h.ParseFlags(ctx, cmd.Flags())
			if err != nil {
				return err
			}
			return h.Create(ctx, cmd, args[0])
		}),
	}
	h.addFlags(createArchiveCmd.Flags())
	archiveCmd.AddCommand(createArchiveCmd)
	return archiveCmd
}

func NewArchiveHandler() *ArchiveHandler {
	return &ArchiveHandler{}
}

type ArchiveHandler struct {
	ignoreDirectories []string
}

// Add the flags for Archive sources
func (a *ArchiveHandler) addFlags(flags *pflag.FlagSet) {
	flags.StringSlice(
		"ignore-dir",
		[]string{".git", "node_modules", "vendor"},
		"Set directories to ignore when building the archive.\n"+
			"To ignore multiple directories, use the flag multiple times.\n"+
			"To include all directories, set the flag to an empty string.",
	)
}

func (a *ArchiveHandler) ParseFlags(ctx *CLIContext, flags *pflag.FlagSet) error {
	ignoreDirectories, err := flags.GetStringSlice("ignore-dir")
	if err != nil {
		return err
	}
	a.ParseIgnoreDirectories(ignoreDirectories)
	return nil
}

func (a *ArchiveHandler) ParseIgnoreDirectories(ignoreDirectories []string) {
	// special case: if the flag is set to an empty string, we  do not ignore any directories
	if len(ignoreDirectories) == 1 && ignoreDirectories[0] == "" {
		ignoreDirectories = []string{}
	}
	a.ignoreDirectories = ignoreDirectories
}
