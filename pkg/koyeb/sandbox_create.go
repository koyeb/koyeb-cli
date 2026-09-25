package koyeb

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"

	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// Create creates a new sandbox service with appropriate defaults
func (h *SandboxHandler) Create(ctx *CLIContext, cmd *cobra.Command, args []string) error {
	if err := setProjectHeader(ctx, cmd); err != nil {
		return err
	}
	return createSandbox(ctx, cmd, args, defaultSandboxCreateDeps())
}

// sandboxCreateDeps carries the API-touching steps of the create flow as
// seams so the wiring is testable without a live client.
type sandboxCreateDeps struct {
	getAppID        func(ctx *CLIContext, name string) (string, error)
	createApp       func(ctx *CLIContext, name string) (string, error)
	resolveSnapshot func(ctx *CLIContext, ref string) (string, koyeb.InstanceSnapshotType)
	createService   func(ctx *CLIContext, cmd *cobra.Command, args []string, req *koyeb.CreateService) (*koyeb.Service, error)
	waitForService  func(ctx *CLIContext, cmd *cobra.Command, serviceID string) error
	deleteApp       func(ctx *CLIContext, appID string)
	deleteService   func(ctx *CLIContext, serviceID string)
	renderService   func(ctx *CLIContext, cmd *cobra.Command, serviceID string)
}

func defaultSandboxCreateDeps() sandboxCreateDeps {
	return sandboxCreateDeps{
		getAppID: getAppIdByName,
		createApp: func(ctx *CLIContext, name string) (string, error) {
			createApp := koyeb.NewCreateAppWithDefaults()
			createApp.SetName(name)
			lifecycle := koyeb.NewAppLifeCycleWithDefaults()
			lifecycle.SetDeleteWhenEmpty(true)
			createApp.SetLifeCycle(*lifecycle)
			reply, err := NewAppHandler().CreateApp(ctx, createApp)
			if err != nil {
				return "", err
			}
			app := reply.GetApp()
			return app.GetId(), nil
		},
		resolveSnapshot: resolveSnapshotFlags,
		createService: func(ctx *CLIContext, cmd *cobra.Command, args []string, req *koyeb.CreateService) (*koyeb.Service, error) {
			return NewServiceHandler().createService(ctx, cmd, args, req)
		},
		waitForService: waitForSandboxDeployment,
		deleteApp:      deleteAppBestEffort,
		deleteService:  deleteServiceBestEffort,
		renderService:  renderServiceState,
	}
}

// createSandbox runs the create flow against the seams in deps. Cleanup is
// phase-aware like the Python SDK: any failure before the service exists
// removes the app this command auto-created (_delete_created_app), and a
// wait failure deletes the service when --cleanup-on-failure is set.
func createSandbox(ctx *CLIContext, cmd *cobra.Command, args []string, deps sandboxCreateDeps) error {
	svcHandler := NewServiceHandler()

	appName, err := svcHandler.parseAppName(cmd, args[0])
	if err != nil {
		return err
	}

	appID, err := deps.getAppID(ctx, appName)
	if err != nil {
		return err
	}

	createdAppID := ""
	if appID == "" {
		log.Infof("Application `%s` does not exist, creating it", appName)
		createdAppID, err = deps.createApp(ctx, appName)
		if err != nil {
			return err
		}
	}
	appCleanup := createdAppID != ""
	defer func() {
		if appCleanup {
			deps.deleteApp(ctx, createdAppID)
		}
	}()

	createDefinition := koyeb.NewDeploymentDefinitionWithDefaults()
	if err := parseSandboxDefinitionFlags(ctx, cmd, createDefinition, svcHandler); err != nil {
		return err
	}
	createDefinition.SetType(koyeb.DEPLOYMENTDEFINITIONTYPE_SANDBOX)

	// Ensure SANDBOX_SECRET exists - explicit flag value, then --env, then generated
	applySandboxSecretFlag(cmd.Flags(), createDefinition)
	ensureSandboxSecret(createDefinition)

	exposedPortProtocol, err := cmd.Flags().GetString("exposed-port-protocol")
	if err != nil {
		return err
	}
	configureSandboxPortsAndRoutes(createDefinition, exposedPortProtocol)

	serviceName, err := svcHandler.parseServiceNameWithoutApp(cmd, args[0])
	if err != nil {
		return err
	}
	createDefinition.SetName(serviceName)

	// Resolve the snapshot only after flag validation so invalid flags do
	// not pay lookup round-trips; a FULL snapshot boots without a definition.
	snapshotID, snapshotType := deps.resolveSnapshot(ctx, GetStringFlags(cmd, "snapshot"))

	createService := koyeb.NewCreateServiceWithDefaults()
	if lifecycle := svcHandler.parseLifeCycle(cmd.Flags(), nil); lifecycle != nil {
		createService.SetLifeCycle(*lifecycle)
	}

	var currentNetworkPolicy *koyeb.NetworkPolicy
	if createDefinition.HasNetworkPolicy() {
		np := createDefinition.GetNetworkPolicy()
		currentNetworkPolicy = &np
	}
	networkPolicy, networkPolicyChanged, err := svcHandler.parseNetworkPolicy(cmd.Flags(), currentNetworkPolicy)
	if err != nil {
		return err
	}
	if networkPolicyChanged && networkPolicy != nil {
		createDefinition.SetNetworkPolicy(*networkPolicy)
	}

	createService.SetDefinition(*createDefinition)
	if snapshotID != "" {
		wireSnapshot(createService, snapshotID, snapshotType, serviceName)
	}

	service, err := deps.createService(ctx, cmd, args, createService)
	if err != nil {
		return err
	}
	appCleanup = false

	if wait := GetBoolFlags(cmd, "wait"); wait {
		if err := deps.waitForService(ctx, cmd, service.GetId()); err != nil {
			if GetBoolFlags(cmd, "cleanup-on-failure") {
				deps.deleteService(ctx, service.GetId())
			}
			return err
		}
	}

	deps.renderService(ctx, cmd, service.GetId())
	return nil
}

// waitForSandboxDeployment polls the created service until it is ready,
// using the SDKs' fail-closed classification: HEALTHY/DEGRADED are ready,
// STARTING/RESUMING are in progress, and anything else — including PAUSED
// and unknown values — will not become ready. Transient GetService errors
// are retried until the timeout, like the SDKs' wait loops.
func waitForSandboxDeployment(ctx *CLIContext, cmd *cobra.Command, serviceID string) error {
	waitTimeout, err := waitTimeoutFlag(cmd)
	if err != nil {
		return err
	}

	getStatus := func(c context.Context, id string) (koyeb.ServiceStatus, error) {
		res, _, err := ctx.Client.ServicesApi.GetService(c, id).Execute()
		if err != nil {
			return "", err
		}
		service := res.GetService()
		if !service.HasStatus() {
			return "", fmt.Errorf("service %s has no status", id)
		}
		return service.GetStatus(), nil
	}
	terminalErr := func(status koyeb.ServiceStatus) error {
		return &errors.CLIError{
			What:     "Sandbox deployment failed",
			Why:      fmt.Sprintf("Service '%s' reached terminal state '%s' and will not become ready.", serviceID, status),
			Solution: errors.CLIErrorSolution("Inspect the sandbox with `koyeb service logs " + serviceID + " -t build`"),
		}
	}
	timeoutErr := func() error {
		return &errors.CLIError{
			What:     "Timed out waiting for the sandbox deployment",
			Why:      fmt.Sprintf("service %s did not become ready within %s", serviceID, waitTimeout),
			Solution: errors.CLIErrorSolution("Check the service status with `koyeb service get " + serviceID + "`, or raise --wait-timeout"),
		}
	}

	return waitForServiceStatus(ctx.Context, serviceID, waitTimeout, waitPollInterval(cmd), getStatus, terminalErr, timeoutErr)
}

// deleteAppBestEffort removes an app auto-created by this command after a
// failed creation; cleanup failures are logged, never raised, so the
// original error reaches the user.
func deleteAppBestEffort(ctx *CLIContext, appID string) {
	_, _, err := ctx.Client.AppsApi.DeleteApp(ctx.Context, appID).Execute()
	if err != nil {
		log.Warnf("Failed to delete app `%s` after sandbox creation failure", appID)
	}
}

// deleteServiceBestEffort removes a sandbox whose wait failed; cleanup
// failures are logged, never raised, so the original error reaches the user.
func deleteServiceBestEffort(ctx *CLIContext, serviceID string) {
	_, _, err := ctx.Client.ServicesApi.DeleteService(ctx.Context, serviceID).Execute()
	if err != nil {
		log.Warnf("Failed to delete service `%s` after sandbox wait failure", serviceID)
	}
}

// resolveSnapshotFlags resolves a snapshot reference to an instance
// snapshot ID and type, like the Python SDK: ID first, then name, then raw
// string. An empty reference returns empty values.
func resolveSnapshotFlags(ctx *CLIContext, ref string) (string, koyeb.InstanceSnapshotType) {
	if ref == "" {
		return "", ""
	}

	get := func(c context.Context, id string) (*koyeb.InstanceSnapshot, error) {
		reply, _, err := ctx.Client.InstanceSnapshotsApi.GetInstanceSnapshot(c, id).Execute()
		if err != nil {
			return nil, err
		}
		snapshot := reply.GetInstanceSnapshot()
		return &snapshot, nil
	}
	list := func(c context.Context) ([]koyeb.InstanceSnapshot, error) {
		reply, _, err := ctx.Client.InstanceSnapshotsApi.ListInstanceSnapshots(c).Name(ref).Execute()
		if err != nil {
			return nil, err
		}
		return reply.GetInstanceSnapshots(), nil
	}

	return resolveSnapshotRef(ctx.Context, ref, get, list)
}

// resolveSnapshotRef resolves a snapshot name-or-ID the way the Python SDK
// does: ID lookup first (fast path for UUIDs), then name lookup, then the
// raw string with the FILESYSTEM type. Lookup failures fall through instead
// of erroring — an unknown ID surfaces server-side at service creation.
func resolveSnapshotRef(ctx context.Context, ref string,
	get func(context.Context, string) (*koyeb.InstanceSnapshot, error),
	list func(context.Context) ([]koyeb.InstanceSnapshot, error),
) (string, koyeb.InstanceSnapshotType) {
	snapshot, err := get(ctx, ref)
	if err == nil && snapshot != nil && snapshot.GetId() != "" {
		return snapshot.GetId(), snapshot.GetType()
	}
	snapshots, err := list(ctx)
	if err == nil {
		for _, s := range snapshots {
			if s.GetName() == ref {
				return s.GetId(), s.GetType()
			}
		}
	}
	return ref, koyeb.INSTANCESNAPSHOTTYPE_FILESYSTEM
}

// wireSnapshot wires boot-from-snapshot on the create request. FULL
// snapshots boot without a definition (the API infers it from the
// snapshot); other snapshot types keep the built definition.
func wireSnapshot(createService *koyeb.CreateService, snapshotID string, snapshotType koyeb.InstanceSnapshotType, serviceName string) {
	createService.SetInstanceSnapshotId(snapshotID)
	if snapshotType == koyeb.INSTANCESNAPSHOTTYPE_FULL {
		createService.Definition = nil
		createService.SetName(serviceName)
	}
}

// parseSandboxDefinitionFlags parses the sandbox-compatible flags using ServiceHandler methods
func parseSandboxDefinitionFlags(ctx *CLIContext, cmd *cobra.Command, def *koyeb.DeploymentDefinition, svcHandler *ServiceHandler) error {
	flags := cmd.Flags()

	// Parse docker source using ServiceHandler method
	dockerSource := koyeb.NewDockerSourceWithDefaults()
	parsedDocker, err := svcHandler.parseDockerSource(ctx, flags, dockerSource)
	if err != nil {
		return err
	}
	// Default to koyeb/sandbox if --docker was not explicitly set
	if !flags.Changed("docker") {
		parsedDocker.SetImage("koyeb/sandbox")
	}
	def.SetDocker(*parsedDocker)

	// Parse instance type using ServiceHandler method
	def.SetInstanceTypes(svcHandler.parseInstanceType(flags, nil))

	// Tri-state mesh, mirroring the SDKs: unset keeps the definition default
	// (AUTO), --enable-mesh maps to ENABLED, --enable-mesh=false to DISABLED.
	if flags.Changed("enable-mesh") {
		if enableMesh, _ := flags.GetBool("enable-mesh"); enableMesh {
			def.SetMesh(koyeb.DEPLOYMENTMESH_ENABLED)
		} else {
			def.SetMesh(koyeb.DEPLOYMENTMESH_DISABLED)
		}
	}

	// Validate the exposed port protocol before building the definition.
	protocol, err := flags.GetString("exposed-port-protocol")
	if err != nil {
		return err
	}
	if protocol != "http" && protocol != "http2" {
		return &errors.CLIError{
			What:     "Invalid exposed port protocol",
			Why:      fmt.Sprintf("Invalid protocol '%s'. Must be one of ('http', 'http2')", protocol),
			Orig:     nil,
			Solution: "Use --exposed-port-protocol http or --exposed-port-protocol http2",
		}
	}

	// Create-time TCP proxy on port 3031, mirroring the SDKs' enable_tcp_proxy.
	if enableTCPProxy, _ := flags.GetBool("enable-tcp-proxy"); enableTCPProxy {
		port := int64(3031)
		proxyProtocol := koyeb.PROXYPORTPROTOCOL_TCP
		def.SetProxyPorts([]koyeb.DeploymentProxyPort{{Port: &port, Protocol: &proxyProtocol}})
	}

	// Parse regions using ServiceHandler method
	regions, err := svcHandler.parseRegions(flags, nil)
	if err != nil {
		return err
	}
	def.SetRegions(regions)

	// Parse environment variables using ServiceHandler method
	envVars, err := svcHandler.parseEnv(flags, nil)
	if err != nil {
		return err
	}
	def.SetEnv(envVars)

	// Parse config files using ServiceHandler method
	parsedFiles, err := svcHandler.parseConfigFiles(ctx, flags, nil)
	if err != nil {
		return err
	}
	def.SetConfigFiles(parsedFiles)

	// Parse scaling for sandbox: only min-scale is supported.
	// max-scale is always 1 (no autoscaling for sandboxes).
	// We handle this directly instead of calling svcHandler.parseScalings()
	// because that function accesses flags (max-scale, scale, autoscaling-*)
	// via flags.Lookup().Changed which would panic since those flags are not
	// registered on the sandbox create command.
	minScale, _ := flags.GetInt64("min-scale")
	scaling := koyeb.NewDeploymentScalingWithDefaults()
	scaling.SetMin(minScale)
	scaling.SetMax(1)

	// Parse sleep delay targets (require min-scale 0).
	// Handled inline for the same reason as above: setScalingsTargets()
	// unconditionally looks up autoscaling flags that don't exist here.
	if flags.Lookup("light-sleep-delay").Changed || flags.Lookup("deep-sleep-delay").Changed {
		if minScale > 0 {
			return &errors.CLIError{
				What: "Error while configuring the sandbox",
				Why:  "--light-sleep-delay and --deep-sleep-delay can only be used when min-scale is 0",
				Additional: []string{
					"Sleep delays are only applicable to services that can scale to zero.",
					"Set --min-scale 0 to enable scale-to-zero before configuring sleep delays.",
				},
				Orig:     nil,
				Solution: "Add --min-scale 0 to your command and try again",
			}
		}

		lightSleepDuration, _ := flags.GetDuration("light-sleep-delay")
		deepSleepDuration, _ := flags.GetDuration("deep-sleep-delay")

		sid := koyeb.NewDeploymentScalingTargetSleepIdleDelay()
		hasValue := false
		if flags.Lookup("light-sleep-delay").Changed && lightSleepDuration > 0 {
			sid.SetLightSleepValue(int64(lightSleepDuration.Seconds()))
			hasValue = true
		}
		if flags.Lookup("deep-sleep-delay").Changed && deepSleepDuration > 0 {
			sid.SetDeepSleepValue(int64(deepSleepDuration.Seconds()))
			hasValue = true
		}
		if hasValue {
			target := koyeb.NewDeploymentScalingTarget()
			target.SetSleepIdleDelay(*sid)
			scaling.Targets = []koyeb.DeploymentScalingTarget{*target}
		}
	}

	def.SetScalings([]koyeb.DeploymentScaling{*scaling})

	return nil
}

// applySandboxSecretFlag sets SANDBOX_SECRET from --sandbox-secret when
// provided, overriding any --env value (SDK parity: the explicit secret
// wins). The generated fallback lives in ensureSandboxSecret.
func applySandboxSecretFlag(flags *pflag.FlagSet, def *koyeb.DeploymentDefinition) {
	secret, err := flags.GetString("sandbox-secret")
	if err != nil || secret == "" {
		return
	}

	setSandboxSecretValue(def, secret)
}

// setSandboxSecretValue replaces any existing SANDBOX_SECRET entry with
// value, preserving the env scopes of the remaining variables.
func setSandboxSecretValue(def *koyeb.DeploymentDefinition, value string) {
	envVars := def.GetEnv()
	filtered := make([]koyeb.DeploymentEnv, 0, len(envVars)+1)
	for _, env := range envVars {
		if env.GetKey() != SandboxSecretKey {
			filtered = append(filtered, env)
		}
	}

	newEnv := koyeb.NewDeploymentEnvWithDefaults()
	newEnv.SetKey(SandboxSecretKey)
	newEnv.SetValue(value)
	if len(filtered) > 0 && len(filtered[0].GetScopes()) > 0 {
		newEnv.SetScopes(filtered[0].GetScopes())
	}

	def.SetEnv(append(filtered, *newEnv))
}

// ensureSandboxSecret adds SANDBOX_SECRET if not already present, generating
// a URL-safe random secret (32 bytes, unpadded base64 — token_urlsafe parity).
func ensureSandboxSecret(def *koyeb.DeploymentDefinition) {
	for _, env := range def.GetEnv() {
		if env.GetKey() == SandboxSecretKey {
			return // Already set by user
		}
	}

	secretBytes := make([]byte, 32)
	rand.Read(secretBytes)
	setSandboxSecretValue(def, base64.RawURLEncoding.EncodeToString(secretBytes))
}

// configureSandboxPortsAndRoutes sets up the sandbox default ports and
// routes: port 3030 (management interface at /koyeb-sandbox/, always http)
// and port 3031 (application endpoint at /, protocol: exposedPortProtocol).
func configureSandboxPortsAndRoutes(def *koyeb.DeploymentDefinition, exposedPortProtocol string) {
	port3030 := koyeb.NewDeploymentPortWithDefaults()
	port3030.SetPort(3030)
	port3030.SetProtocol("http")

	port3031 := koyeb.NewDeploymentPortWithDefaults()
	port3031.SetPort(3031)
	port3031.SetProtocol(exposedPortProtocol)

	def.SetPorts([]koyeb.DeploymentPort{*port3030, *port3031})

	routeManagement := koyeb.NewDeploymentRouteWithDefaults()
	routeManagement.SetPort(3030)
	routeManagement.SetPath("/koyeb-sandbox/")

	routeApp := koyeb.NewDeploymentRouteWithDefaults()
	routeApp.SetPort(3031)
	routeApp.SetPath("/")

	def.SetRoutes([]koyeb.DeploymentRoute{*routeManagement, *routeApp})
}
