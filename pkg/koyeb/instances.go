package koyeb

import (
	"context"
	"errors"
	"fmt"

	"github.com/moby/term"

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

	stdin, stdout, stderr := term.StdStreams()
	e := NewExecutor(stdin, stdout, stderr, userCmd, instanceId)
	returnCode, err := e.Run(context.Background())
	fmt.Printf("Return code is %d\n", returnCode)
	return err
}
