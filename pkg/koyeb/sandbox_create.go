package koyeb

import (
	"crypto/rand"
	"encoding/base64"

	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// Create creates a new sandbox service with appropriate defaults
func (h *SandboxHandler) Create(ctx *CLIContext, cmd *cobra.Command, args []string) error {
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

	if appId == "" {
		log.Infof("Application `%s` does not exist, creating it", appName)
		createApp := koyeb.NewCreateAppWithDefaults()
		createApp.SetName(appName)
		lifecycle := koyeb.NewAppLifeCycleWithDefaults()
		lifecycle.SetDeleteWhenEmpty(true)
		createApp.SetLifeCycle(*lifecycle)
		appHandler := NewAppHandler()
		if _, err := appHandler.CreateApp(ctx, createApp); err != nil {
			return err
		}
	}

	// Parse sandbox-compatible flags using ServiceHandler methods
	if err := parseSandboxDefinitionFlags(ctx, cmd, createDefinition, svcHandler); err != nil {
		return err
	}

	// Force type to SANDBOX
	createDefinition.SetType(koyeb.DEPLOYMENTDEFINITIONTYPE_SANDBOX)

	// Ensure SANDBOX_SECRET exists - generate if not provided
	ensureSandboxSecret(createDefinition)

	// Configure sandbox-specific ports and routes (always use defaults for sandbox)
	configureSandboxPortsAndRoutes(createDefinition, false, false)

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

	createService.SetDefinition(*createDefinition)

	// Delegate to ServiceHandler.Create for API call
	return svcHandler.Create(ctx, cmd, args, createService)
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
	secret := base64.URLEncoding.EncodeToString(secretBytes)

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
// Port 3031: Application endpoint at /
func configureSandboxPortsAndRoutes(def *koyeb.DeploymentDefinition, portsExplicitlySet, routesExplicitlySet bool) {
	// Set sandbox default ports unless user explicitly set --ports flag
	if !portsExplicitlySet {
		port3030 := koyeb.NewDeploymentPortWithDefaults()
		port3030.SetPort(3030)
		port3030.SetProtocol("http")

		port3031 := koyeb.NewDeploymentPortWithDefaults()
		port3031.SetPort(3031)
		port3031.SetProtocol("http")

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
