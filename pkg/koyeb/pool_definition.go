package koyeb

import (
	"fmt"
	"strings"

	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/errors"
	"github.com/spf13/pflag"
)

// addPoolFlags registers the curated flag set shared by pool create and
// pool update. The sandbox-only knobs (exposed-port-protocol,
// enable-tcp-proxy) are accepted on SANDBOX pools only —
// validatePoolSandboxKnobs rejects them on every other type. Pool secrets
// and mesh stay off the pool surface per the cross-client contract: the
// platform mints the executor secret and mesh stays AUTO.
func addPoolFlags(flags *pflag.FlagSet) {
	flags.Int64("size", 1, "Number of instances kept ready in the pool")

	flags.String("type", "sandbox", `Pool type: "web", "worker" or "sandbox" (default)`)
	flags.StringSlice(
		"ports",
		nil,
		"Member ports for WEB and WORKER pools using the format PORT[:PROTOCOL], for example --port 8080:http\n"+
			"PROTOCOL defaults to \"http\". Supported protocols are \"http\", \"http2\" and \"tcp\"\n"+
			"Explicit ports are rejected on SANDBOX pools: the sandbox wiring owns ports 3030/3031\n"+
			"To remove a port on update, prefix its number with '!', for example --port '!80'\n",
	)
	flags.StringSlice(
		"routes",
		nil,
		"Member routes for WEB and WORKER pools using the format PATH[:PORT], for example --route /foo:8080\n"+
			"PORT defaults to 8000\n"+
			"Explicit routes are rejected on SANDBOX pools: the sandbox wiring owns ports 3030/3031\n"+
			"To remove a route on update, prefix its path with '!', for example --route '!/foo'\n",
	)

	flags.String("docker", "", "Docker image (default: koyeb/sandbox)")
	flags.String("docker-private-registry-secret", "", "Docker private registry secret")
	flags.StringSlice("docker-entrypoint", []string{}, "Docker entrypoint")
	flags.String("docker-command", "", "Docker command")
	flags.StringSlice("docker-args", []string{}, "Docker command arguments")
	flags.Bool("privileged", false, "Whether the member containers run in privileged mode")

	flags.String("exposed-port-protocol", "http",
		"Protocol for the exposed application port 3031 (http or http2), SANDBOX pools only")
	flags.Bool("enable-tcp-proxy", false,
		"Expose port 3031 via TCP proxy, SANDBOX pools only")

	flags.String("instance-type", "micro", "Instance type")
	flags.StringSlice("regions", []string{}, "Deployment regions")

	flags.StringSlice("env", []string{}, "Environment variables (KEY=VALUE)")
	flags.StringSlice("config-file", nil, "Config files (LOCAL:REMOTE:PERMS)")

	flags.Int64("min-scale", 1, "Min scale")

	// Member egress policy, shared with the service and sandbox surfaces.
	addNetworkPolicyFlags(flags)

	flags.Duration("light-sleep-delay", 0,
		"Delay after which an idle service is put to light sleep. "+
			"Use duration format (e.g., '1m', '5m', '1h'). Set to 0 to disable.")
	flags.Duration("deep-sleep-delay", 0,
		"Delay after which an idle service is put to deep sleep. "+
			"Use duration format (e.g., '5m', '30m', '1h'). Set to 0 to disable.")

	// Match the service commands: --port and --route alias --ports and --routes.
	flags.SetNormalizeFunc(poolFlagAliases)
}

func poolFlagAliases(_ *pflag.FlagSet, name string) pflag.NormalizedName {
	aliases := map[string]string{
		"port":  "ports",
		"route": "routes",
	}
	if alias, exists := aliases[name]; exists {
		name = alias
	}
	return pflag.NormalizedName(name)
}

// parsePoolType parses the --type flag. Pools host WEB, WORKER and SANDBOX
// definitions; DATABASE is a product-rule exclusion rejected fail-fast
// with the rule spelled out, and anything else is invalid.
func parsePoolType(flags *pflag.FlagSet) (koyeb.DeploymentDefinitionType, error) {
	value, _ := flags.GetString("type")
	switch strings.ToUpper(value) {
	case "WEB":
		return koyeb.DEPLOYMENTDEFINITIONTYPE_WEB, nil
	case "WORKER":
		return koyeb.DEPLOYMENTDEFINITIONTYPE_WORKER, nil
	case "SANDBOX":
		return koyeb.DEPLOYMENTDEFINITIONTYPE_SANDBOX, nil
	case "DATABASE":
		return "", &errors.CLIError{
			What: "Error while configuring the pool",
			Why:  "the pool type \"database\" is not supported",
			Additional: []string{
				"Service pools host WEB, WORKER and SANDBOX definitions only; DATABASE definitions are not poolable.",
				"Use `koyeb database create` to provision a database instead.",
			},
			Orig:     nil,
			Solution: "Fix the --type flag and try again",
		}
	default:
		return "", &errors.CLIError{
			What: "Error while configuring the pool",
			Why:  "the --type flag is not valid",
			Additional: []string{
				`The --type flag must be one of "web", "worker" or "sandbox"`,
			},
			Orig:     nil,
			Solution: "Fix the --type flag and try again",
		}
	}
}

// validatePoolSandboxKnobs enforces the inverse wiring rule: the
// sandbox-only knobs (exposed port protocol, TCP proxy) never apply to
// non-SANDBOX pools — their members carry exactly the declared
// --port/--route wiring. Mirrors the python reference's fail-fast check.
func validatePoolSandboxKnobs(poolType koyeb.DeploymentDefinitionType, flags *pflag.FlagSet) error {
	if poolType == koyeb.DEPLOYMENTDEFINITIONTYPE_SANDBOX {
		return nil
	}
	protocolSet := flags.Lookup("exposed-port-protocol").Changed
	tcpProxy, _ := flags.GetBool("enable-tcp-proxy")
	if !protocolSet && !tcpProxy {
		return nil
	}
	return &errors.CLIError{
		What: "Error while configuring the pool",
		Why: "--exposed-port-protocol and --enable-tcp-proxy are sandbox-only options " +
			"and are not allowed on WEB/WORKER pools",
		Additional: []string{
			"The knobs configure the executor wiring that only SANDBOX pool members carry.",
			"Non-SANDBOX members are wired with the --port and --route flags instead.",
		},
		Orig:     nil,
		Solution: "Remove the sandbox-only flags or set --type sandbox, and try again",
	}
}

// validatePoolWiringFlags enforces the cross-client wiring rule: SANDBOX
// pools never take explicit ports or routes — the sandbox wiring owns
// ports 3030/3031, and user-declared wiring would break executor
// connectivity. Never silently ignore the flags.
func validatePoolWiringFlags(poolType koyeb.DeploymentDefinitionType, flags *pflag.FlagSet) error {
	if poolType != koyeb.DEPLOYMENTDEFINITIONTYPE_SANDBOX {
		return nil
	}
	for _, flag := range []string{"ports", "routes"} {
		if !flags.Lookup(flag).Changed {
			continue
		}
		return &errors.CLIError{
			What: "Error while configuring the pool",
			Why: fmt.Sprintf(
				"explicit %s are not allowed on SANDBOX pools: the sandbox wiring owns ports 3030/3031",
				flag),
			Additional: []string{
				"The pool members' executor connectivity depends on the sandbox wiring; user-declared wiring would break it.",
				`Create a WEB or WORKER pool with --type to declare member ports and routes.`,
			},
			Orig:     nil,
			Solution: "Remove the --port/--route flags or set --type web|worker, and try again",
		}
	}
	return nil
}
