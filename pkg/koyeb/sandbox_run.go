package koyeb

import (
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/koyeb/koyeb-cli/pkg/koyeb/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// Run executes a command in the sandbox
func (h *SandboxHandler) Run(ctx *CLIContext, cmd *cobra.Command, args []string) error {
	sandboxName := args[0]
	command := strings.Join(args[1:], " ")

	// Get and validate flags
	cwd, err := cmd.Flags().GetString("cwd")
	if err != nil {
		return &errors.CLIError{
			What:     "Error parsing flags",
			Why:      "failed to parse --cwd flag",
			Orig:     err,
			Solution: "Check the flag syntax",
		}
	}

	envSlice, err := cmd.Flags().GetStringSlice("env")
	if err != nil {
		return &errors.CLIError{
			What:     "Error parsing flags",
			Why:      "failed to parse --env flag",
			Orig:     err,
			Solution: "Check the flag syntax (use KEY=VALUE format)",
		}
	}

	timeout, err := cmd.Flags().GetInt("timeout")
	if err != nil {
		return &errors.CLIError{
			What:     "Error parsing flags",
			Why:      "failed to parse --timeout flag",
			Orig:     err,
			Solution: "Provide a valid integer for timeout",
		}
	}

	stream, err := cmd.Flags().GetBool("stream")
	if err != nil {
		return &errors.CLIError{
			What:     "Error parsing flags",
			Why:      "failed to parse --stream flag",
			Orig:     err,
			Solution: "Check the flag syntax",
		}
	}

	// Validate timeout
	timeout = ValidateTimeout(timeout)

	// Parse environment variables with validation
	env, warnings := ParseEnvVars(envSlice)
	for _, w := range warnings {
		log.Warn(w)
	}

	// Get sandbox info and create client
	info, err := h.GetSandboxInfo(ctx, sandboxName)
	if err != nil {
		return err
	}

	client := NewSandboxClient(
		info.Domain,
		info.SandboxSecret,
		WithTimeout(time.Duration(timeout)*time.Second),
	)

	req := &RunRequest{
		Cmd:     command,
		Cwd:     cwd,
		Timeout: timeout,
	}
	if len(env) > 0 {
		req.Env = env
	}

	if stream {
		return h.runStreaming(ctx, client, req)
	}

	return h.runBuffered(ctx, client, req)
}

// runBuffered executes command with buffered output
func (h *SandboxHandler) runBuffered(ctx *CLIContext, client *SandboxClient, req *RunRequest) error {
	result, err := client.Run(ctx.Context, req)
	if err != nil {
		return &errors.CLIError{
			What:       "Error while executing command in sandbox",
			Why:        "the command execution failed",
			Additional: nil,
			Orig:       err,
			Solution:   "Check that the sandbox is running and the command is valid",
		}
	}

	if result.Stdout != "" {
		fmt.Print(result.Stdout)
	}
	if result.Stderr != "" {
		fmt.Fprint(os.Stderr, result.Stderr)
	}

	if result.Code != 0 {
		os.Exit(result.Code)
	}

	return nil
}

// runStreaming executes command with streaming output
// Uses mutex to safely capture exit code from callback
func (h *SandboxHandler) runStreaming(ctx *CLIContext, client *SandboxClient, req *RunRequest) error {
	var (
		exitCode int
		mu       sync.Mutex
		done     = make(chan struct{})
	)

	err := client.RunStreaming(ctx.Context, req,
		func(stream, data string) {
			if stream == StreamStdout {
				fmt.Print(data)
			} else {
				fmt.Fprint(os.Stderr, data)
			}
		},
		func(code int, hasError bool) {
			mu.Lock()
			exitCode = code
			mu.Unlock()
			close(done)
		},
	)

	if err != nil {
		return &errors.CLIError{
			What:       "Error while executing command in sandbox",
			Why:        "the streaming command execution failed",
			Additional: nil,
			Orig:       err,
			Solution:   "Check that the sandbox is running and the command is valid",
		}
	}

	// Wait for completion callback to be processed
	<-done

	// Safely read exit code
	mu.Lock()
	code := exitCode
	mu.Unlock()

	if code != 0 {
		os.Exit(code)
	}

	return nil
}
