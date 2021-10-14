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

func (h *InstanceHandler) Exec(*cobra.Command, []string) error {
	return errors.New("unimplemented")
}
