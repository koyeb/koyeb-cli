package koyeb

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
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
	sandboxCmd.PersistentFlags().StringP("project", "p", "", "Workspace ID or name")
	sandboxCmd.PersistentFlags().String("workspace", "", "Workspace ID or name (alias for --project)")

	// list command - list all sandboxes
	listSandboxCmd := &cobra.Command{
		Use:   "list",
		Short: "List sandboxes",
		RunE:  WithCLIContext(h.List),
	}
	listSandboxCmd.Flags().StringP("app", "a", "", "App")
	listSandboxCmd.Flags().StringP("name", "n", "", "Sandbox name")
	sandboxCmd.AddCommand(listSandboxCmd)

	// get command - get a sandbox's status and connection info
	getSandboxCmd := &cobra.Command{
		Use:               "get NAME",
		Short:             "Get sandbox status and connection info",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completeServiceIdentifiers,
		RunE:              WithCLIContext(h.Get),
	}
	sandboxCmd.AddCommand(getSandboxCmd)

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

	// delete command - delete a sandbox
	deleteSandboxCmd := &cobra.Command{
		Use:               "delete NAME",
		Short:             "Delete a sandbox",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completeServiceIdentifiers,
		RunE:              WithCLIContext(h.Delete),
	}
	sandboxCmd.AddCommand(deleteSandboxCmd)

	// update command - update lifecycle and network policy of a running sandbox
	updateSandboxCmd := &cobra.Command{
		Use:   "update NAME",
		Short: "Update a running sandbox's lifecycle and network policy",
		Args:  cobra.ExactArgs(1),
		Long: `Update a running sandbox's lifecycle and egress network policy, mirroring
the SDKs' update_lifecycle and update_network_policy.

Lifecycle flags update only the fields you supply: the other settings keep
their current values, and a 0 duration disables the corresponding delay.

Network policy flags configure outbound (egress) traffic: --block-network
denies all outbound traffic, --outbound-allowlist denies by default and
allows only the listed destinations (bare IPs are normalized to /32 (IPv4)
or /128 (IPv6)), and --no-network-policy reverts to the platform default
(unrestricted outbound access).

Applying an update redeploys the sandbox: it is restarted and any
in-memory or non-persisted state is lost.`,
		Example: `
# Auto-delete the sandbox 24 hours after creation
$> koyeb sandbox update myapp/mysandbox --delete-after-delay 24h

# Auto-delete the sandbox 30 minutes after it sleeps due to inactivity
$> koyeb sandbox update myapp/mysandbox --delete-after-inactivity-delay 30m

# Disable the delete-after-creation delay
$> koyeb sandbox update myapp/mysandbox --delete-after-delay 0

# Block all outbound network traffic
$> koyeb sandbox update myapp/mysandbox --block-network

# Allow outbound traffic only to specific destinations
$> koyeb sandbox update myapp/mysandbox --outbound-allowlist 10.0.0.0/8

# Revert to the platform default network policy
$> koyeb sandbox update myapp/mysandbox --no-network-policy
`,
		RunE: WithCLIContext(h.Update),
	}
	addSandboxLifeCycleFlags(updateSandboxCmd.Flags())
	addNetworkPolicyFlags(updateSandboxCmd.Flags())
	sandboxCmd.AddCommand(updateSandboxCmd)

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

	// kill-all command - kill every running background process
	killAllCmd := &cobra.Command{
		Use:   "kill-all NAME",
		Short: "Kill all running background processes in the sandbox",
		Long: `Kill all running background processes in the sandbox.

Every running process is killed in one call; a process that fails to die does
not stop the remaining ones. The killed count is reported. Mirrors the Koyeb
SDKs' kill_all.`,
		Args: cobra.ExactArgs(1),
		Example: `
# Kill every running background process in one call
$> koyeb sandbox kill-all myapp/mysandbox
`,
		RunE: WithCLIContext(h.KillAllProcesses),
	}
	sandboxCmd.AddCommand(killAllCmd)

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

	// snapshot subcommand - instance snapshots of a running sandbox
	sandboxSnapshotCmd := &cobra.Command{
		Use:   "snapshot",
		Short: "Instance snapshot operations",
		Long: `Instance snapshot operations for sandboxes.

An instance snapshot captures the state of a running sandbox. Snapshots boot
new sandboxes via 'koyeb sandbox create --snapshot' and are what the Koyeb
SDKs' snapshot() and spawn() consume. These are distinct from the volume
snapshots managed by 'koyeb snapshots'.`,
	}

	createSandboxSnapshotCmd := &cobra.Command{
		Use:   "create NAME",
		Short: "Create an instance snapshot of a running sandbox",
		Args:  cobra.ExactArgs(1),
		Example: `
# Snapshot a running sandbox
$> koyeb sandbox snapshot create myapp/mysandbox

# Name the snapshot and wait until it is available
$> koyeb sandbox snapshot create myapp/mysandbox --name my-snap --wait

# Capture a full snapshot (filesystem plus running processes)
$> koyeb sandbox snapshot create myapp/mysandbox --type full
`,
		RunE: WithCLIContext(h.SnapshotCreate),
	}
	createSandboxSnapshotCmd.Flags().String("name", "", "Snapshot name (default: <sandbox>-snapshot-<unix timestamp>)")
	createSandboxSnapshotCmd.Flags().String("type", "filesystem", "Snapshot type: filesystem or full")
	createSandboxSnapshotCmd.Flags().Bool("wait", false, "Wait until the snapshot is available (opt-in; the Koyeb SDKs wait by default)")
	createSandboxSnapshotCmd.Flags().Duration("wait-timeout", 10*time.Minute, "Wait timeout duration")
	createSandboxSnapshotCmd.Flags().Float64("poll-interval", 5, "Seconds between availability polls when --wait is set")
	sandboxSnapshotCmd.AddCommand(createSandboxSnapshotCmd)
	sandboxCmd.AddCommand(sandboxSnapshotCmd)

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
Content can be provided as an argument or from a local file with -f flag.
Parent directories must already exist (create them first with fs mkdir).`,
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
		Long: `Upload a local file or directory to the sandbox.
Parent directories on the sandbox must already exist (create them first with fs mkdir).`,
		Args: cobra.ExactArgs(3),
		RunE: WithCLIContext(h.FsUpload),
	}
	fsUploadCmd.Flags().BoolP("recursive", "r", false, "Upload directories recursively")
	fsUploadCmd.Flags().BoolP("force", "f", false, "Overwrite existing remote directory")
	fsCmd.AddCommand(fsUploadCmd)

	// fs rename
	fsRenameCmd := &cobra.Command{
		Use:   "rename NAME OLD_PATH NEW_PATH",
		Short: "Rename a file or directory in the sandbox",
		Args:  cobra.ExactArgs(3),
		RunE:  WithCLIContext(h.FsRename),
	}
	fsCmd.AddCommand(fsRenameCmd)

	// fs move
	fsMoveCmd := &cobra.Command{
		Use:   "move NAME SOURCE_PATH DESTINATION_PATH",
		Short: "Move a file to a different directory in the sandbox",
		Args:  cobra.ExactArgs(3),
		RunE:  WithCLIContext(h.FsMove),
	}
	fsCmd.AddCommand(fsMoveCmd)

	// fs exists / is-file / is-dir
	fsExistsCmd := &cobra.Command{
		Use:   "exists NAME PATH",
		Short: "Check if a path exists in the sandbox",
		Args:  cobra.ExactArgs(2),
		RunE:  WithCLIContext(h.FsExists),
	}
	fsCmd.AddCommand(fsExistsCmd)

	fsIsFileCmd := &cobra.Command{
		Use:   "is-file NAME PATH",
		Short: "Check if a path is a regular file in the sandbox",
		Args:  cobra.ExactArgs(2),
		RunE:  WithCLIContext(h.FsIsFile),
	}
	fsCmd.AddCommand(fsIsFileCmd)

	fsIsDirCmd := &cobra.Command{
		Use:   "is-dir NAME PATH",
		Short: "Check if a path is a directory in the sandbox",
		Args:  cobra.ExactArgs(2),
		RunE:  WithCLIContext(h.FsIsDir),
	}
	fsCmd.AddCommand(fsIsDirCmd)

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
	// BaseURL, when non-empty, is the sandbox API base URL derived from
	// DeploymentMetadata.Sandbox.PublicUrl (with the "/koyeb-sandbox" path
	// appended). When set, it takes precedence over Domain for client
	// construction.
	BaseURL string
	// RoutingKey, when non-empty, is the value of
	// DeploymentMetadata.Sandbox.RoutingKey. It is sent as the X-Routing-Key
	// header on every sandbox HTTP request, in addition to the Authorization
	// Bearer token.
	RoutingKey string
}

// ConnectionURL returns the URL used to reach the sandbox executor: the
// API-provided base URL when DeploymentMetadata.Sandbox exposes one, and
// the legacy https://{Domain}/koyeb-sandbox URL otherwise. It is the same
// URL NewClient builds its client from.
func (i *SandboxInfo) ConnectionURL() string {
	if i.BaseURL != "" {
		return i.BaseURL
	}
	return fmt.Sprintf("https://%s/koyeb-sandbox", i.Domain)
}

// NewClient creates a SandboxClient using the connection information in i.
// When BaseURL is set (provided by the API via DeploymentMetadata.Sandbox),
// it is used as-is and the RoutingKey is sent as the X-Routing-Key header.
// Otherwise the legacy mechanism (https://{Domain}/koyeb-sandbox) is used.
// In all cases the SandboxSecret is sent as a Bearer token in the
// Authorization header.
func (i *SandboxInfo) NewClient(opts ...SandboxClientOption) *SandboxClient {
	if i.BaseURL != "" {
		return newSandboxClientFromBaseURL(i.BaseURL, i.SandboxSecret, i.RoutingKey, opts...)
	}
	return NewSandboxClient(i.Domain, i.SandboxSecret, opts...)
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

// sandboxOp is one executor operation against a resolved sandbox connection.
type sandboxOp func(client SandboxClientInterface, info *SandboxInfo) error

// withSandboxClient resolves the sandbox connection, builds an executor
// client, and runs op against it.
func withSandboxClient(ctx *CLIContext, sandboxName string, op sandboxOp, opts ...SandboxClientOption) error {
	info, err := fetchSandboxInfo(ctx, sandboxName)
	if err != nil {
		return err
	}
	return op(info.NewClient(opts...), info)
}

// resolveSandboxService resolves a sandbox reference (name, short ID, or
// full UUID) to its service through the id mapper, failing unless the
// service is a SANDBOX.
func resolveSandboxService(ctx *CLIContext, name string) (*koyeb.Service, error) {
	serviceID, err := ctx.Mapper.Service().ResolveID(name)
	if err != nil {
		return nil, &errors.CLIError{
			What:       "Error while resolving sandbox",
			Why:        fmt.Sprintf("could not find sandbox '%s'", name),
			Additional: nil,
			Orig:       err,
			Solution:   "Make sure the sandbox exists. Use 'koyeb service list' to see available services.",
		}
	}

	serviceRes, resp, err := ctx.API.GetService(ctx.Context, serviceID)
	if err != nil {
		return nil, errors.NewCLIErrorFromAPIError(
			fmt.Sprintf("Error while retrieving sandbox '%s'", name),
			err,
			resp,
		)
	}

	service := serviceRes.GetService()
	if service.GetType() != koyeb.SERVICETYPE_SANDBOX {
		return nil, &errors.CLIError{
			What:       "Error while accessing sandbox",
			Why:        fmt.Sprintf("service '%s' is not a sandbox (type: %s)", name, service.GetType()),
			Additional: nil,
			Solution:   "Use a service created with --type=sandbox",
		}
	}
	return &service, nil
}

// fetchSandboxInfo retrieves sandbox info from the API
func fetchSandboxInfo(ctx *CLIContext, name string) (*SandboxInfo, error) {
	service, err := resolveSandboxService(ctx, name)
	if err != nil {
		return nil, err
	}
	return sandboxConnectionInfo(ctx, name, service)
}

// sandboxConnectionInfo computes the executor connection information of a
// resolved sandbox: the domain, the SANDBOX_SECRET and, when the API exposes
// DeploymentMetadata.Sandbox, the public base URL and routing key.
func sandboxConnectionInfo(ctx *CLIContext, name string, service *koyeb.Service) (*SandboxInfo, error) {
	appID := service.GetAppId()

	// Get app to find domain
	appRes, resp, err := ctx.API.GetApp(ctx.Context, appID)
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

	deploymentRes, resp, err := ctx.API.GetDeployment(ctx.Context, deploymentID)
	if err != nil {
		return nil, errors.NewCLIErrorFromAPIError(
			fmt.Sprintf("Error while retrieving deployment for sandbox '%s'", name),
			err,
			resp,
		)
	}

	deployment := deploymentRes.GetDeployment()

	// SANDBOX_SECRET is always required: it is sent as a Bearer token in the
	// Authorization header on every sandbox HTTP request, regardless of
	// whether DeploymentMetadata.Sandbox is provided by the API.
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

	// When the API exposes DeploymentMetadata.Sandbox, prefer its values:
	// PublicUrl provides the host to reach (the "/koyeb-sandbox" path is
	// still appended), and RoutingKey is sent in the X-Routing-Key header in
	// addition to the Bearer token above.
	var (
		baseURL    string
		routingKey string
	)
	if deployment.HasMetadata() {
		metadata := deployment.GetMetadata()
		if metadata.HasSandbox() {
			sandbox := metadata.GetSandbox()
			if publicURL := sandbox.GetPublicUrl(); publicURL != "" {
				baseURL = strings.TrimRight(publicURL, "/") + "/koyeb-sandbox"
			}
			routingKey = sandbox.GetRoutingKey()
		}
	}

	return &SandboxInfo{
		ServiceID:     service.GetId(),
		AppID:         appID,
		Domain:        domain,
		SandboxSecret: sandboxSecret,
		BaseURL:       baseURL,
		RoutingKey:    routingKey,
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

// Health checks sandbox health status
func (h *SandboxHandler) Health(ctx *CLIContext, cmd *cobra.Command, args []string) error {
	sandboxName := args[0]

	return withSandboxClient(ctx, sandboxName, func(client SandboxClientInterface, _ *SandboxInfo) error {
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
	})
}

// addSandboxLifeCycleFlags registers the lifecycle flags shared by
// `sandbox create` and `sandbox update`.
func addSandboxLifeCycleFlags(flags *pflag.FlagSet) {
	flags.Duration("delete-after-delay", 0, "Auto-delete after duration (e.g., '24h')")
	flags.Duration("delete-after-inactivity-delay", 0, "Auto-delete after inactivity (e.g., '1h')")
}

func addSandboxCreateFlags(cmd *cobra.Command) {
	flags := cmd.Flags()

	// Sandbox-specific flags
	flags.StringP("app", "a", "", "Sandbox application")
	flags.Bool("wait", false, "Wait until the sandbox deployment is ready (opt-in; the Koyeb SDKs wait by default)")
	flags.Duration("wait-timeout", 5*time.Minute, "Wait timeout duration")
	flags.Float64("poll-interval", 0.5, "Seconds between readiness polls when --wait is set")
	flags.Bool("cleanup-on-failure", true, "Delete the sandbox when --wait fails or times out")

	// Parity flags mirroring the Koyeb SDKs' create options
	flags.Bool("enable-mesh", false, "Enable mesh for this sandbox (tri-state: unset=AUTO, true=ENABLED, false=DISABLED)")
	flags.String("exposed-port-protocol", "http", "Protocol for the exposed application port 3031 (http or http2)")
	flags.Bool("enable-tcp-proxy", false, "Expose port 3031 via TCP proxy at create time")
	flags.String("sandbox-secret", "", "Explicit sandbox secret (defaults to a generated one)")
	flags.String("snapshot", "", "Instance snapshot ID or name to boot the sandbox from")

	// Docker source flags (required for sandbox)
	addDockerSourceFlags(flags, sandboxPoolDockerSourceFlagUsage)

	// Instance type and regions: the shared bundle registers both.
	addInstanceTypeRegionsFlags(flags, sandboxPoolInstanceTypeRegionsFlagUsage)

	// Environment and configuration: the shared bundle registers both.
	addEnvConfigFilesFlags(flags, sandboxPoolEnvConfigFilesFlagUsage)

	// Lifecycle flags
	addSandboxLifeCycleFlags(flags)

	// Scaling and sleep delay flags: the shared bundle (sandboxes always
	// run at max-scale=1, no autoscaling).
	addScalingSleepDelayFlags(flags, sandboxPoolScalingSleepDelayFlagUsage)

	// Network policy flags
	addNetworkPolicyFlags(flags)

	// Other compatible flags
	flags.Bool("privileged", false, "Run in privileged mode")
}
