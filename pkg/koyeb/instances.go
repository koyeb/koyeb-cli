package koyeb

import (
	"errors"

	"github.com/spf13/cobra"
)

func NewInstanceCmd() *cobra.Command {
	h := NewInstanceHandler()
	instanceCmd := &cobra.Command{
		Use:     "instances [action]",
		Aliases: []string{"i", "instance"},
		Short:   "Instances",
	}

	execInstanceCmd := &cobra.Command{
		Use:   "exec [name] [cmd] [cmd...]",
		Short: "Execute command in instance's context",
		Args:  cobra.MinimumNArgs(2),
		RunE:  h.Exec,
	}
	instanceCmd.AddCommand(execInstanceCmd)

	return instanceCmd
}

func NewInstanceHandler() *InstanceHandler {
	return &InstanceHandler{}
}

type InstanceHandler struct {
}

func (h *InstanceHandler) Exec(cmd *cobra.Command, args []string) error {
	// Cobra options ensure we have at least 2 arguments here, but still
	if len(args) < 2 {
		return errors.New("exec needs at least 2 arguments")
	}
	instanceId, userCmd := args[0], args[1:]

	return h.exec(instanceId, userCmd, []string{"Client hello", "Client bye"})
}
