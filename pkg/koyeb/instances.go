package koyeb

import (
	"context"
	"os"

	"github.com/pkg/errors"
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
	returnCode, err := h.exec(cmd, args)
	if err != nil {
		return err
	}
	if returnCode != 0 {
		os.Exit(returnCode)
	}
	return nil
}

func (h *InstanceHandler) exec(cmd *cobra.Command, args []string) (int, error) {
	// Cobra options ensure we have at least 2 arguments here, but still
	if len(args) < 2 {
		return 0, errors.New("exec needs at least 2 arguments")
	}
	instanceId, userCmd := args[0], args[1:]

	stdStreams, cleanup, err := GetStdStreams()
	if err != nil {
		return 0, errors.Wrap(err, "could not get standard streams")
	}
	defer cleanup()

	e := NewExecutor(stdStreams.Stdin, stdStreams.Stdout, stdStreams.Stderr, userCmd, instanceId)
	return e.Run(context.Background())
}
