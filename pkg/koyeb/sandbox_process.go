package koyeb

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/koyeb/koyeb-cli/pkg/koyeb/errors"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/renderer"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// StartProcess starts a background process in the sandbox
func (h *SandboxHandler) StartProcess(ctx *CLIContext, cmd *cobra.Command, args []string) error {
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

	// Parse environment variables with validation
	env, warnings := ParseEnvVars(envSlice)
	for _, w := range warnings {
		log.Warn(w)
	}

	info, err := h.GetSandboxInfo(ctx, sandboxName)
	if err != nil {
		return err
	}

	client := NewSandboxClient(info.Domain, info.SandboxSecret)

	req := &ProcessRequest{
		Cmd: command,
		Cwd: cwd,
	}
	if len(env) > 0 {
		req.Env = env
	}

	result, err := client.StartProcess(ctx.Context, req)
	if err != nil {
		return &errors.CLIError{
			What:       "Error while starting process in sandbox",
			Why:        "the process failed to start",
			Additional: nil,
			Orig:       err,
			Solution:   "Check that the sandbox is running and the command is valid",
		}
	}

	log.Infof("Process started with ID: %s (PID: %d, Status: %s)", result.ID, result.PID, result.Status)
	fmt.Printf("Process ID: %s\n", result.ID)

	return nil
}

// ListProcesses lists background processes in the sandbox
func (h *SandboxHandler) ListProcesses(ctx *CLIContext, cmd *cobra.Command, args []string) error {
	sandboxName := args[0]

	info, err := h.GetSandboxInfo(ctx, sandboxName)
	if err != nil {
		return err
	}

	client := NewSandboxClient(info.Domain, info.SandboxSecret)

	processes, err := client.ListProcesses(ctx.Context)
	if err != nil {
		return &errors.CLIError{
			What:       "Error while listing processes in sandbox",
			Why:        "failed to retrieve process list",
			Additional: nil,
			Orig:       err,
			Solution:   "Check that the sandbox is running",
		}
	}

	if len(processes) == 0 {
		fmt.Println("No background processes running")
		return nil
	}

	full := GetBoolFlags(cmd, "full")
	reply := NewListProcessesReply(processes, full)
	ctx.Renderer.Render(reply)
	return nil
}

// KillProcess kills a background process in the sandbox
func (h *SandboxHandler) KillProcess(ctx *CLIContext, cmd *cobra.Command, args []string) error {
	sandboxName := args[0]
	processID := args[1]

	info, err := h.GetSandboxInfo(ctx, sandboxName)
	if err != nil {
		return err
	}

	client := NewSandboxClient(info.Domain, info.SandboxSecret)

	err = client.KillProcess(ctx.Context, processID)
	if err != nil {
		return &errors.CLIError{
			What:       "Error while killing process in sandbox",
			Why:        "failed to kill process",
			Additional: []string{fmt.Sprintf("Process ID: %s", processID)},
			Orig:       err,
			Solution:   "Check that the process ID is correct and the process is running",
		}
	}

	log.Infof("Process %s killed successfully", processID)
	return nil
}

// ProcessLogs streams logs from a background process
func (h *SandboxHandler) ProcessLogs(ctx *CLIContext, cmd *cobra.Command, args []string) error {
	sandboxName := args[0]
	processID := args[1]

	follow, err := cmd.Flags().GetBool("follow")
	if err != nil {
		return &errors.CLIError{
			What:     "Error parsing flags",
			Why:      "failed to parse --follow flag",
			Orig:     err,
			Solution: "Check the flag syntax",
		}
	}

	info, err := h.GetSandboxInfo(ctx, sandboxName)
	if err != nil {
		return err
	}

	client := NewSandboxClient(info.Domain, info.SandboxSecret)

	if follow {
		log.Info("Streaming logs (press Ctrl+C to stop)...")
	}

	err = client.StreamProcessLogs(ctx.Context, processID, follow, func(timestamp, stream, data string) {
		// Format output with stream indicator
		streamIndicator := "stdout"
		if stream == StreamStderr {
			streamIndicator = "stderr"
		}
		fmt.Printf("[%s] %s\n", streamIndicator, data)
	})

	if err != nil {
		return &errors.CLIError{
			What:       "Error while streaming process logs",
			Why:        "failed to stream logs",
			Additional: []string{fmt.Sprintf("Process ID: %s", processID)},
			Orig:       err,
			Solution:   "Check that the process ID is correct",
		}
	}

	return nil
}

// ListProcessesReply implements the renderer interface for process listing
type ListProcessesReply struct {
	processes []ProcessInfo
	full      bool
}

func NewListProcessesReply(processes []ProcessInfo, full bool) *ListProcessesReply {
	return &ListProcessesReply{
		processes: processes,
		full:      full,
	}
}

func (ListProcessesReply) Title() string {
	return "Processes"
}

func (r *ListProcessesReply) MarshalBinary() ([]byte, error) {
	return json.Marshal(r.processes)
}

func (r *ListProcessesReply) Headers() []string {
	return []string{"id", "pid", "status", "command"}
}

func (r *ListProcessesReply) Fields() []map[string]string {
	resp := make([]map[string]string, 0, len(r.processes))

	for _, p := range r.processes {
		command := p.Command
		if !r.full && len(command) > 50 {
			command = command[:47] + "..."
		}

		pidStr := "-"
		if p.PID > 0 {
			pidStr = fmt.Sprintf("%d", p.PID)
		}

		fields := map[string]string{
			"id":      renderer.FormatID(p.ID, r.full),
			"pid":     pidStr,
			"status":  p.Status,
			"command": command,
		}
		resp = append(resp, fields)
	}

	return resp
}
