package koyeb

import (
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/errors"
	"github.com/spf13/cobra"
)

// Constants for sandbox operations
const (
	DefaultTimeout   = 30
	MinPort          = 1
	MaxPort          = 65535
	SandboxSecretKey = "SANDBOX_SECRET"
	StreamStdout     = "stdout"
	StreamStderr     = "stderr"
)

func NewSandboxCmd() *cobra.Command {
	h := NewSandboxHandler()

	sandboxCmd := &cobra.Command{
		Use:     "sandbox ACTION",
		Aliases: []string{"sb"},
		Short:   "Sandbox - interactive execution environments",
		Long: `Sandbox commands for interacting with sandbox services.

Sandboxes are created using 'koyeb service create --type=sandbox'.
These commands provide additional functionality for running commands,
managing processes, filesystem operations, and port exposure.`,
	}

	// list command - list all sandboxes
	listSandboxCmd := &cobra.Command{
		Use:   "list",
		Short: "List sandboxes",
		RunE:  WithCLIContext(h.List),
	}
	listSandboxCmd.Flags().StringP("app", "a", "", "App")
	listSandboxCmd.Flags().StringP("name", "n", "", "Sandbox name")
	sandboxCmd.AddCommand(listSandboxCmd)

	// create command - create a new sandbox
	createSandboxCmd := &cobra.Command{
		Use:   "create NAME",
		Short: "Create a new sandbox",
		Args:  cobra.ExactArgs(1),
		Example: `
# Create a sandbox in an app
$> koyeb sandbox create myapp/mysandbox

# Create with a custom docker image
$> koyeb sandbox create myapp/mysandbox --docker myregistry/myimage

# Create with custom secret
$> koyeb sandbox create myapp/mysandbox --env SANDBOX_SECRET=mysecret

# Create and wait for deployment
$> koyeb sandbox create myapp/mysandbox --wait
`,
		RunE: WithCLIContext(h.Create),
	}
	// Add only sandbox-compatible flags (not the full service definition flags)
	addSandboxCreateFlags(createSandboxCmd)
	sandboxCmd.AddCommand(createSandboxCmd)

	// run command - execute a command synchronously
	runSandboxCmd := &cobra.Command{
		Use:   "run NAME COMMAND [ARGS...]",
		Short: "Execute a command in the sandbox",
		Args:  cobra.MinimumNArgs(2),
		Example: `
# Run a simple command
$> koyeb sandbox run myapp/mysandbox echo "Hello World"

# Run a command with arguments
$> koyeb sandbox run myapp/mysandbox ls -la /app

# Run a python script
$> koyeb sandbox run myapp/mysandbox python script.py

# Run with custom working directory
$> koyeb sandbox run myapp/mysandbox --cwd /app python main.py

# Run with streaming output
$> koyeb sandbox run myapp/mysandbox --stream long-running-command

# Run with custom timeout (in seconds)
$> koyeb sandbox run myapp/mysandbox --timeout 120 long-running-command
`,
		RunE: WithCLIContext(h.Run),
	}
	runSandboxCmd.Flags().String("cwd", "", "Working directory for the command")
	runSandboxCmd.Flags().StringSlice("env", nil, "Environment variables (KEY=VALUE)")
	runSandboxCmd.Flags().Int("timeout", DefaultTimeout, "Command timeout in seconds")
	runSandboxCmd.Flags().Bool("stream", false, "Stream output in real-time")
	sandboxCmd.AddCommand(runSandboxCmd)

	// start command - start a background process
	startProcessCmd := &cobra.Command{
		Use:     "start NAME COMMAND [ARGS...]",
		Short:   "Start a background process in the sandbox",
		Aliases: []string{"launch"},
		Args:    cobra.MinimumNArgs(2),
		Example: `
# Start a web server in background
$> koyeb sandbox start myapp/mysandbox python -m http.server 8080

# Start a process with custom working directory
$> koyeb sandbox start myapp/mysandbox --cwd /app npm start
`,
		RunE: WithCLIContext(h.StartProcess),
	}
	startProcessCmd.Flags().String("cwd", "", "Working directory for the process")
	startProcessCmd.Flags().StringSlice("env", nil, "Environment variables (KEY=VALUE)")
	sandboxCmd.AddCommand(startProcessCmd)

	// ps command - list processes
	psCmd := &cobra.Command{
		Use:     "ps NAME",
		Short:   "List background processes in the sandbox",
		Aliases: []string{"list-processes"},
		Args:    cobra.ExactArgs(1),
		RunE:    WithCLIContext(h.ListProcesses),
	}
	sandboxCmd.AddCommand(psCmd)

	// kill command - kill a process
	killCmd := &cobra.Command{
		Use:   "kill NAME PROCESS_ID",
		Short: "Kill a background process in the sandbox",
		Args:  cobra.ExactArgs(2),
		RunE:  WithCLIContext(h.KillProcess),
	}
	sandboxCmd.AddCommand(killCmd)

	// logs command - stream process logs
	processLogsCmd := &cobra.Command{
		Use:   "logs NAME PROCESS_ID",
		Short: "Stream logs from a background process",
		Args:  cobra.ExactArgs(2),
		RunE:  WithCLIContext(h.ProcessLogs),
	}
	processLogsCmd.Flags().BoolP("follow", "f", false, "Follow log output (like tail -f)")
	sandboxCmd.AddCommand(processLogsCmd)

	// expose-port command
	exposePortCmd := &cobra.Command{
		Use:   "expose-port NAME PORT",
		Short: "Expose a port from the sandbox via TCP proxy",
		Args:  cobra.ExactArgs(2),
		Example: `
# Expose port 8080
$> koyeb sandbox expose-port myapp/mysandbox 8080
`,
		RunE: WithCLIContext(h.ExposePort),
	}
	sandboxCmd.AddCommand(exposePortCmd)

	// unexpose-port command
	unexposePortCmd := &cobra.Command{
		Use:   "unexpose-port NAME",
		Short: "Unexpose the currently exposed port",
		Args:  cobra.ExactArgs(1),
		RunE:  WithCLIContext(h.UnexposePort),
	}
	sandboxCmd.AddCommand(unexposePortCmd)

	// health command
	healthCmd := &cobra.Command{
		Use:   "health NAME",
		Short: "Check sandbox health status",
		Args:  cobra.ExactArgs(1),
		RunE:  WithCLIContext(h.Health),
	}
	sandboxCmd.AddCommand(healthCmd)

	// filesystem subcommand
	fsCmd := &cobra.Command{
		Use:     "fs",
		Aliases: []string{"filesystem"},
		Short:   "Filesystem operations",
	}

	// fs read
	fsReadCmd := &cobra.Command{
		Use:   "read NAME PATH",
		Short: "Read a file from the sandbox",
		Args:  cobra.ExactArgs(2),
		RunE:  WithCLIContext(h.FsRead),
	}
	fsCmd.AddCommand(fsReadCmd)

	// fs write
	fsWriteCmd := &cobra.Command{
		Use:   "write NAME PATH [CONTENT]",
		Short: "Write content to a file in the sandbox",
		Long: `Write content to a file in the sandbox.
Content can be provided as an argument or via stdin with -f flag.`,
		Args: cobra.RangeArgs(2, 3),
		Example: `
# Write inline content
$> koyeb sandbox fs write myapp/mysandbox /tmp/hello.txt "Hello World"

# Write from local file
$> koyeb sandbox fs write myapp/mysandbox /tmp/script.py -f ./local-script.py
`,
		RunE: WithCLIContext(h.FsWrite),
	}
	fsWriteCmd.Flags().StringP("file", "f", "", "Read content from local file")
	fsCmd.AddCommand(fsWriteCmd)

	// fs ls
	fsLsCmd := &cobra.Command{
		Use:   "ls NAME [PATH]",
		Short: "List directory contents in the sandbox",
		Args:  cobra.RangeArgs(1, 2),
		RunE:  WithCLIContext(h.FsLs),
	}
	fsLsCmd.Flags().BoolP("long", "l", false, "Use long listing format with details")
	fsCmd.AddCommand(fsLsCmd)

	// fs mkdir
	fsMkdirCmd := &cobra.Command{
		Use:   "mkdir NAME PATH",
		Short: "Create a directory in the sandbox",
		Args:  cobra.ExactArgs(2),
		RunE:  WithCLIContext(h.FsMkdir),
	}
	fsCmd.AddCommand(fsMkdirCmd)

	// fs rm
	fsRmCmd := &cobra.Command{
		Use:   "rm NAME PATH",
		Short: "Remove a file or directory from the sandbox",
		Args:  cobra.ExactArgs(2),
		RunE:  WithCLIContext(h.FsRm),
	}
	fsRmCmd.Flags().BoolP("recursive", "r", false, "Remove directories recursively")
	fsCmd.AddCommand(fsRmCmd)

	// fs upload
	fsUploadCmd := &cobra.Command{
		Use:   "upload NAME LOCAL_PATH REMOTE_PATH",
		Short: "Upload a local file or directory to the sandbox (max 1G per file)",
		Args:  cobra.ExactArgs(3),
		RunE:  WithCLIContext(h.FsUpload),
	}
	fsUploadCmd.Flags().BoolP("recursive", "r", false, "Upload directories recursively")
	fsUploadCmd.Flags().BoolP("force", "f", false, "Overwrite existing remote directory")
	fsCmd.AddCommand(fsUploadCmd)

	// fs download
	fsDownloadCmd := &cobra.Command{
		Use:   "download NAME REMOTE_PATH LOCAL_PATH",
		Short: "Download a file from the sandbox",
		Args:  cobra.ExactArgs(3),
		RunE:  WithCLIContext(h.FsDownload),
	}
	fsCmd.AddCommand(fsDownloadCmd)

	sandboxCmd.AddCommand(fsCmd)

	return sandboxCmd
}

func NewSandboxHandler() *SandboxHandler {
	return &SandboxHandler{}
}

type SandboxHandler struct{}

// SandboxInfo contains the information needed to connect to a sandbox
type SandboxInfo struct {
	ServiceID     string
	AppID         string
	Domain        string
	SandboxSecret string
	ProxyPort     string // The external proxy port (from health check)
}

// ValidatePort checks if a port number is valid
func ValidatePort(portStr string) (int, error) {
	port, err := strconv.Atoi(portStr)
	if err != nil {
		return 0, &errors.CLIError{
			What:     "Invalid port",
			Why:      fmt.Sprintf("'%s' is not a valid number", portStr),
			Solution: "Provide a port number between 1 and 65535",
		}
	}

	if port < MinPort || port > MaxPort {
		return 0, &errors.CLIError{
			What:     "Invalid port",
			Why:      fmt.Sprintf("port %d is out of valid range", port),
			Solution: errors.CLIErrorSolution(fmt.Sprintf("Provide a port number between %d and %d", MinPort, MaxPort)),
		}
	}

	return port, nil
}

// ParseEnvVars parses environment variables from KEY=VALUE format
// Returns the parsed map and any warnings for malformed entries
func ParseEnvVars(envSlice []string) (map[string]string, []string) {
	env := make(map[string]string)
	var warnings []string

	for _, e := range envSlice {
		if e == "" {
			continue
		}

		parts := strings.SplitN(e, "=", 2)
		if len(parts) != 2 {
			warnings = append(warnings, fmt.Sprintf("ignoring malformed env var '%s' (expected KEY=VALUE format)", e))
			continue
		}

		key := strings.TrimSpace(parts[0])
		if key == "" {
			warnings = append(warnings, fmt.Sprintf("ignoring env var with empty key: '%s'", e))
			continue
		}

		env[key] = parts[1]
	}

	return env, warnings
}

// ValidateTimeout checks if timeout is valid and returns a default if not set
func ValidateTimeout(timeout int) int {
	if timeout < 1 {
		return DefaultTimeout
	}
	return timeout
}

// GetSandboxInfo resolves sandbox name to connection info
func (h *SandboxHandler) GetSandboxInfo(ctx *CLIContext, name string) (*SandboxInfo, error) {
	return h.fetchSandboxInfo(ctx, name)
}

// fetchSandboxInfo retrieves sandbox info from the API
func (h *SandboxHandler) fetchSandboxInfo(ctx *CLIContext, name string) (*SandboxInfo, error) {
	// Resolve service ID
	serviceMapper := ctx.Mapper.Service()
	serviceID, err := serviceMapper.ResolveID(name)
	if err != nil {
		return nil, &errors.CLIError{
			What:       "Error while resolving sandbox",
			Why:        fmt.Sprintf("could not find sandbox '%s'", name),
			Additional: nil,
			Orig:       err,
			Solution:   "Make sure the sandbox exists. Use 'koyeb service list' to see available services.",
		}
	}

	// Get service details
	serviceRes, resp, err := ctx.Client.ServicesApi.GetService(ctx.Context, serviceID).Execute()
	if err != nil {
		return nil, errors.NewCLIErrorFromAPIError(
			fmt.Sprintf("Error while retrieving sandbox '%s'", name),
			err,
			resp,
		)
	}

	service := serviceRes.GetService()

	// Verify it's a sandbox type
	if service.GetType() != koyeb.SERVICETYPE_SANDBOX {
		return nil, &errors.CLIError{
			What:       "Error while accessing sandbox",
			Why:        fmt.Sprintf("service '%s' is not a sandbox (type: %s)", name, service.GetType()),
			Additional: nil,
			Solution:   "Use a service created with --type=sandbox",
		}
	}

	appID := service.GetAppId()

	// Get app to find domain
	appRes, resp, err := ctx.Client.AppsApi.GetApp(ctx.Context, appID).Execute()
	if err != nil {
		return nil, errors.NewCLIErrorFromAPIError(
			fmt.Sprintf("Error while retrieving application for sandbox '%s'", name),
			err,
			resp,
		)
	}

	app := appRes.GetApp()
	domains := app.GetDomains()
	if len(domains) == 0 {
		return nil, &errors.CLIError{
			What:       "Error while accessing sandbox",
			Why:        "sandbox has no domain assigned",
			Additional: []string{"The sandbox may still be deploying"},
			Solution:   "Wait for the sandbox to be fully deployed and try again",
		}
	}

	// Find the best domain - prefer .koyeb.app domains for reliability
	domain := selectBestDomain(domains)

	// Get deployment to find SANDBOX_SECRET
	deploymentID := service.GetActiveDeploymentId()
	if deploymentID == "" {
		deploymentID = service.GetLatestDeploymentId()
	}

	if deploymentID == "" {
		return nil, &errors.CLIError{
			What:       "Error while accessing sandbox",
			Why:        "sandbox has no deployment",
			Additional: nil,
			Solution:   "Wait for the sandbox to be deployed",
		}
	}

	deploymentRes, resp, err := ctx.Client.DeploymentsApi.GetDeployment(ctx.Context, deploymentID).Execute()
	if err != nil {
		return nil, errors.NewCLIErrorFromAPIError(
			fmt.Sprintf("Error while retrieving deployment for sandbox '%s'", name),
			err,
			resp,
		)
	}

	deployment := deploymentRes.GetDeployment()
	definition := deployment.GetDefinition()
	envVars := definition.GetEnv()

	var sandboxSecret string
	for _, env := range envVars {
		if env.GetKey() == SandboxSecretKey {
			sandboxSecret = env.GetValue()
			break
		}
	}

	if sandboxSecret == "" {
		return nil, &errors.CLIError{
			What:       "Error while accessing sandbox",
			Why:        fmt.Sprintf("%s not found in sandbox environment", SandboxSecretKey),
			Additional: []string{"The sandbox may not have been configured correctly"},
			Solution:   errors.CLIErrorSolution(fmt.Sprintf("Ensure the sandbox was created with %s environment variable", SandboxSecretKey)),
		}
	}

	// The API returns env var values base64-encoded
	decoded, err := base64.StdEncoding.DecodeString(sandboxSecret)
	if err != nil {
		// If it's not valid base64, use the value as-is
		decoded = []byte(sandboxSecret)
	}
	sandboxSecret = string(decoded)

	return &SandboxInfo{
		ServiceID:     serviceID,
		AppID:         appID,
		Domain:        domain,
		SandboxSecret: sandboxSecret,
	}, nil
}

// selectBestDomain selects the most reliable domain from the list
func selectBestDomain(domains []koyeb.Domain) string {
	// Prefer .koyeb.app domains as they're managed by the platform
	for _, d := range domains {
		name := d.GetName()
		if strings.HasSuffix(name, ".koyeb.app") {
			return name
		}
	}

	// Fall back to first domain
	return domains[0].GetName()
}

// GetClientWithHealthCheck creates a client and verifies sandbox is healthy
func (h *SandboxHandler) GetClientWithHealthCheck(ctx *CLIContext, sandboxName string) (*SandboxClient, *SandboxInfo, error) {
	info, err := h.GetSandboxInfo(ctx, sandboxName)
	if err != nil {
		return nil, nil, err
	}

	client := NewSandboxClient(info.Domain, info.SandboxSecret)

	// Perform health check
	health, err := client.Health(ctx.Context)
	if err != nil {
		return nil, nil, &errors.CLIError{
			What:       "Error while connecting to sandbox",
			Why:        "failed to reach sandbox",
			Additional: []string{fmt.Sprintf("Domain: %s", info.Domain)},
			Orig:       err,
			Solution:   "Check that the sandbox is running and accessible",
		}
	}

	if !health.Healthy {
		return nil, nil, &errors.CLIError{
			What:       "Sandbox is not healthy",
			Why:        health.Status,
			Additional: nil,
			Solution:   "Wait for the sandbox to become healthy or check sandbox logs",
		}
	}

	// Store proxy port if available
	if health.ProxyPort != "" {
		info.ProxyPort = health.ProxyPort
	}

	return client, info, nil
}

// Health checks sandbox health status
func (h *SandboxHandler) Health(ctx *CLIContext, cmd *cobra.Command, args []string) error {
	sandboxName := args[0]

	info, err := h.GetSandboxInfo(ctx, sandboxName)
	if err != nil {
		return err
	}

	client := NewSandboxClient(info.Domain, info.SandboxSecret)

	health, err := client.Health(ctx.Context)
	if err != nil {
		return &errors.CLIError{
			What:       "Error checking sandbox health",
			Why:        "failed to connect to sandbox",
			Additional: nil,
			Orig:       err,
			Solution:   "Check that the sandbox is deployed and accessible",
		}
	}

	fmt.Printf("Sandbox: %s\n", sandboxName)
	fmt.Printf("Status: %s\n", health.Status)
	fmt.Printf("Healthy: %v\n", health.Healthy)
	if health.Version != "" {
		fmt.Printf("Version: %s\n", health.Version)
	}
	if health.Uptime > 0 {
		fmt.Printf("Uptime: %s\n", time.Duration(health.Uptime)*time.Second)
	}
	if health.ProxyPort != "" {
		fmt.Printf("Proxy Port: %s\n", health.ProxyPort)
	}

	return nil
}

// addSandboxCreateFlags adds only the flags that are compatible with sandbox services
// This excludes flags like --type, --git*, --archive*, --ports, --routes, --checks
// which are either not applicable or handled automatically for sandboxes
func addSandboxCreateFlags(cmd *cobra.Command) {
	flags := cmd.Flags()

	// Sandbox-specific flags
	flags.StringP("app", "a", "", "Sandbox application")
	flags.Bool("wait", false, "Wait until sandbox deployment is done")
	flags.Duration("wait-timeout", 5*time.Minute, "Wait timeout duration")

	// Docker source flags (required for sandbox)
	flags.String("docker", "", "Docker image (default: koyeb/sandbox)")
	flags.String("docker-private-registry-secret", "", "Docker private registry secret")
	flags.StringSlice("docker-entrypoint", []string{}, "Docker entrypoint")
	flags.String("docker-command", "", "Docker command")
	flags.StringSlice("docker-args", []string{}, "Docker command arguments")

	// Instance flags
	flags.String("instance-type", "nano", "Instance type")

	// Region flags
	flags.StringSlice("regions", []string{}, "Deployment regions")

	// Environment and configuration
	flags.StringSlice("env", []string{}, "Environment variables (KEY=VALUE)")
	flags.StringSlice("config-file", nil, "Config files (LOCAL:REMOTE:PERMS)")

	// Lifecycle flags
	flags.Duration("delete-after-delay", 0, "Auto-delete after duration (e.g., '24h')")
	flags.Duration("delete-after-inactivity-delay", 0, "Auto-delete after inactivity (e.g., '1h')")

	// Scaling flags (sandboxes always run with max-scale=1, no autoscaling)
	flags.Int64("min-scale", 1, "Min scale")

	// Sleep delay flags (require min-scale 0)
	flags.Duration("light-sleep-delay", 0,
		"Delay after which an idle service is put to light sleep. "+
			"Use duration format (e.g., '1m', '5m', '1h'). Set to 0 to disable.")
	flags.Duration("deep-sleep-delay", 0,
		"Delay after which an idle service is put to deep sleep. "+
			"Use duration format (e.g., '5m', '30m', '1h'). Set to 0 to disable.")

	// Other compatible flags
	flags.Bool("privileged", false, "Run in privileged mode")
}
