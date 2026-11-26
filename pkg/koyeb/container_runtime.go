package koyeb

import (
	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
	"github.com/spf13/pflag"
)

// Container runtime flag, shared by `service create/update`, `deploy`,
// `sandbox create` and the pool surfaces (the runtime lives on
// DeploymentDefinition — any change creates a new deployment). The flag is
// never --runtime: "runtime" already means runtime logs in the CLI.
func addContainerRuntimeFlags(flags *pflag.FlagSet) {
	flags.String("container-runtime", "",
		`Container runtime of the workload: "docker", "kata" or "gvisor". `+
			"Empty (the default) uses the region default runtime")
}

// applyContainerRuntimeFlag sets the definition's container runtime when
// --container-runtime was passed, and leaves the definition alone
// otherwise — so `service update` and the pool full-replace round-trip the
// stored runtime. The CLI does not validate the allowed runtimes: the value
// is sent verbatim and API rejections surface unchanged.
func applyContainerRuntimeFlag(flags *pflag.FlagSet, definition *koyeb.DeploymentDefinition) {
	if !flags.Lookup("container-runtime").Changed {
		return
	}
	runtime, _ := flags.GetString("container-runtime")
	// An explicit empty value drops back to the region default runtime.
	if runtime == "" {
		definition.ContainerRuntime = nil
		return
	}
	definition.SetContainerRuntime(runtime)
}
