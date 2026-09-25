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
	createService := koyeb.NewCreateServiceWithDefaults()
	createDefinition := koyeb.NewDeploymentDefinitionWithDefaults()

	// Use ServiceHandler for parsing common flags - reuse existing methods to avoid duplication
	svcHandler := NewServiceHandler()

	// Auto-create app if it doesn't exist
	appName, err := svcHandler.parseAppName(cmd, args[0])
	if err != nil {
		return err
	}

	appId, err := getAppIdByName(ctx, appName)
	if err != nil {
		return err
	}

	createdAppID := ""
	if appId == "" {
		log.Infof("Application `%s` does not exist, creating it", appName)
		createApp := koyeb.NewCreateAppWithDefaults()
		createApp.SetName(appName)
		lifecycle := koyeb.NewAppLifeCycleWithDefaults()
		lifecycle.SetDeleteWhenEmpty(true)
		createApp.SetLifeCycle(*lifecycle)
		appHandler := NewAppHandler()
		reply, err := appHandler.CreateApp(ctx, createApp)
		if err != nil {
			return err
		}
		app := reply.GetApp()
		createdAppID = app.GetId()
	}

	// Resolve the snapshot reference before building the definition: a FULL
	// snapshot boots without one.
	snapshotID, snapshotType := resolveSnapshotFlags(ctx, cmd)

	// Parse sandbox-compatible flags using ServiceHandler methods
	if err := parseSandboxDefinitionFlags(ctx, cmd, createDefinition, svcHandler); err != nil {
		return err
	}

	// Force type to SANDBOX
	createDefinition.SetType(koyeb.DEPLOYMENTDEFINITIONTYPE_SANDBOX)

	// Ensure SANDBOX_SECRET exists - explicit flag value, then --env, then generated
	applySandboxSecretFlag(cmd.Flags(), createDefinition)
	ensureSandboxSecret(createDefinition)

	// Configure sandbox-specific ports and routes (always use defaults for sandbox)
	exposedPortProtocol, err := cmd.Flags().GetString("exposed-port-protocol")
	if err != nil {
		return err
	}
	configureSandboxPortsAndRoutes(createDefinition, exposedPortProtocol, false, false)

	// Set service name
	serviceName, err := svcHandler.parseServiceNameWithoutApp(cmd, args[0])
	if err != nil {
		return err
	}
	createDefinition.SetName(serviceName)

	// Parse lifecycle flags using ServiceHandler method
	lifecycle := svcHandler.parseLifeCycle(cmd.Flags(), nil)
	if lifecycle != nil {
		createService.SetLifeCycle(*lifecycle)
	}

	// Parse network policy flags using ServiceHandler method
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

	// The sandbox flow owns create and wait so cleanup can react to which
	// phase failed (Python: app cleanup on create failure, service cleanup
	// gated by cleanup_on_failure on wait failure).
	service, err := svcHandler.createService(ctx, cmd, args, createService)
	if err != nil {
		if createdAppID != "" {
			deleteAppBestEffort(ctx, createdAppID)
		}
		return err
	}
	defer renderServiceState(ctx, cmd, service.GetId())

	if wait := GetBoolFlags(cmd, "wait"); wait {
		if err := waitForServiceDeployment(ctx, cmd, service.GetId()); err != nil {
			if GetBoolFlags(cmd, "cleanup-on-failure") {
				deleteServiceBestEffort(ctx, service.GetId())
			}
			return err
		}
	}

	return nil
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

// resolveSnapshotFlags resolves --snapshot to an instance snapshot ID and
// type. An empty flag returns empty values; otherwise the reference is
// resolved like the Python SDK: ID first, then name, then raw string.
func resolveSnapshotFlags(ctx *CLIContext, cmd *cobra.Command) (string, koyeb.InstanceSnapshotType) {
	ref := GetStringFlags(cmd, "snapshot")
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

// ensureSandboxSecret adds SANDBOX_SECRET env var if not already present
func ensureSandboxSecret(def *koyeb.DeploymentDefinition) {
	envVars := def.GetEnv()

	// Check if SANDBOX_SECRET already exists
	for _, env := range envVars {
		if env.GetKey() == SandboxSecretKey {
			return // Already set by user
		}
	}

	// Generate secure random secret (32 bytes, URL-safe base64)
	secretBytes := make([]byte, 32)
	rand.Read(secretBytes)
	secret := base64.RawURLEncoding.EncodeToString(secretBytes)

	// Create new env var
	newEnv := koyeb.NewDeploymentEnvWithDefaults()
	newEnv.SetKey(SandboxSecretKey)
	newEnv.SetValue(secret)

	// Copy scopes from existing env vars if present
	if len(envVars) > 0 && len(envVars[0].GetScopes()) > 0 {
		newEnv.SetScopes(envVars[0].GetScopes())
	}

	def.SetEnv(append(envVars, *newEnv))
}

// configureSandboxPortsAndRoutes sets up default sandbox ports and routes
// Port 3030: Management interface at /koyeb-sandbox/
// Port 3031: Application endpoint at / (protocol: exposedPortProtocol)
func configureSandboxPortsAndRoutes(def *koyeb.DeploymentDefinition, exposedPortProtocol string, portsExplicitlySet, routesExplicitlySet bool) {
	// Set sandbox default ports unless user explicitly set --ports flag
	if !portsExplicitlySet {
		port3030 := koyeb.NewDeploymentPortWithDefaults()
		port3030.SetPort(3030)
		port3030.SetProtocol("http")

		port3031 := koyeb.NewDeploymentPortWithDefaults()
		port3031.SetPort(3031)
		port3031.SetProtocol(exposedPortProtocol)

		def.SetPorts([]koyeb.DeploymentPort{*port3030, *port3031})
	}

	// Set sandbox default routes unless user explicitly set --routes flag
	if !routesExplicitlySet {
		routeManagement := koyeb.NewDeploymentRouteWithDefaults()
		routeManagement.SetPort(3030)
		routeManagement.SetPath("/koyeb-sandbox/")

		routeApp := koyeb.NewDeploymentRouteWithDefaults()
		routeApp.SetPort(3031)
		routeApp.SetPath("/")

		def.SetRoutes([]koyeb.DeploymentRoute{*routeManagement, *routeApp})
	}
}
