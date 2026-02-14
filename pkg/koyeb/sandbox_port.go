package koyeb

import (
	"fmt"

	"github.com/koyeb/koyeb-cli/pkg/koyeb/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// ExposePort binds a port to the TCP proxy
func (h *SandboxHandler) ExposePort(ctx *CLIContext, cmd *cobra.Command, args []string) error {
	sandboxName := args[0]
	portStr := args[1]

	// Validate port number
	port, err := ValidatePort(portStr)
	if err != nil {
		return err
	}

	info, err := h.GetSandboxInfo(ctx, sandboxName)
	if err != nil {
		return err
	}

	client := NewSandboxClient(info.Domain, info.SandboxSecret)

	result, err := client.BindPort(ctx.Context, fmt.Sprintf("%d", port))
	if err != nil {
		return &errors.CLIError{
			What:       "Error while exposing port",
			Why:        "failed to bind port to TCP proxy",
			Additional: []string{fmt.Sprintf("Port: %d", port)},
			Orig:       err,
			Solution:   "Check that the sandbox is running and supports TCP proxy",
		}
	}

	if !result.Success {
		if result.CurrentPort != "" {
			return &errors.CLIError{
				What:       "Error while exposing port",
				Why:        fmt.Sprintf("port %s is already bound", result.CurrentPort),
				Additional: []string{"Only one port can be exposed at a time"},
				Solution:   errors.CLIErrorSolution(fmt.Sprintf("Run 'koyeb sandbox unexpose-port %s' first, then try again", sandboxName)),
			}
		}
		return &errors.CLIError{
			What:       "Error while exposing port",
			Why:        result.Error,
			Additional: nil,
			Solution:   errors.CLIErrorSolution(fmt.Sprintf("Check that the port number is valid (%d-%d)", MinPort, MaxPort)),
		}
	}

	log.Infof("Port %d exposed successfully", port)
	fmt.Printf("Port %d is now exposed via TCP proxy\n", port)

	// Use the proxy port from the response, not a hardcoded value
	if result.ProxyPort != "" {
		fmt.Printf("External proxy port: %s\n", result.ProxyPort)
		fmt.Printf("Access it at: %s:%s\n", info.Domain, result.ProxyPort)
	} else {
		fmt.Printf("Access it via: %s\n", info.Domain)
	}

	return nil
}

// UnexposePort unbinds the current port from the TCP proxy
func (h *SandboxHandler) UnexposePort(ctx *CLIContext, cmd *cobra.Command, args []string) error {
	sandboxName := args[0]

	info, err := h.GetSandboxInfo(ctx, sandboxName)
	if err != nil {
		return err
	}

	client := NewSandboxClient(info.Domain, info.SandboxSecret)

	result, err := client.UnbindPort(ctx.Context)
	if err != nil {
		return &errors.CLIError{
			What:       "Error while unexposing port",
			Why:        "failed to unbind port from TCP proxy",
			Additional: nil,
			Orig:       err,
			Solution:   "Check that the sandbox is running",
		}
	}

	if !result.Success {
		if result.Error != "" {
			return &errors.CLIError{
				What:       "Error while unexposing port",
				Why:        result.Error,
				Additional: nil,
				Solution:   "Check that a port is currently exposed",
			}
		}
		// No error but no success - likely no port was bound
		fmt.Println("No port was currently exposed")
		return nil
	}

	log.Infof("Port unexposed successfully")
	fmt.Println("Port binding removed")

	return nil
}
