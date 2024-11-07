// Parse the variables --env variables from `koyeb service update` and `koyeb service create`.
package flags_list

import (
	"fmt"
	"strings"

	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/errors"
)

type FlagEnv struct {
	BaseFlag
	key      string
	isSecret bool
	value    string
}

func NewEnvListFromFlags(values []string) ([]Flag[koyeb.DeploymentEnv], error) {
	ret := make([]Flag[koyeb.DeploymentEnv], 0, len(values))

	for _, value := range values {
		env := &FlagEnv{BaseFlag: BaseFlag{cliValue: value}}

		if strings.HasPrefix(value, "!") {
			env.markedForDeletion = true
			split := strings.Split(value, "=")
			if len(split) != 1 {
				return nil, &errors.CLIError{
					What: "Error while configuring the service",
					Why:  fmt.Sprintf("unable to parse the environment variable \"%s\"", value),
					Additional: []string{
						"To delete an environment variable, prefix it with !",
						"You must omit the value, e.g. !KEY and not !KEY=value",
						"Do not forget to escape the ! character if you are using a shell, e.g. \\!KEY or '!KEY'",
					},
					Orig:     nil,
					Solution: "Fix the environment variable and try again",
				}
			}
			env.key = split[0][1:] // Skip the ! character
		} else {
			split := strings.SplitN(value, "=", 2)
			// If there is no =, or the key is empty, or the value refers to a secret without a name
			if len(split) != 2 || split[0] == "" || split[1] == "@" {
				return nil, &errors.CLIError{
					What: "Error while configuring the service",
					Why:  fmt.Sprintf("unable to parse the environment variable \"%s\"", value),
					Additional: []string{
						"Environment variables must be specified as KEY=VALUE",
						"To use a secret as a value, specify KEY=@SECRET_NAME",
						"To specify an empty value, specify KEY=",
						"To remove an environment variable, prefix it with !",
						"Do not forget to escape the ! character if you are using a shell, e.g. \\!KEY or '!KEY'",
					},
					Orig:     nil,
					Solution: "Fix the environment variable and try again",
				}
			}
			env.key = split[0]
			if strings.HasPrefix(split[1], "@") {
				env.isSecret = true
				env.value = split[1][1:]
			} else {
				env.value = split[1]
			}
		}
		ret = append(ret, env)
	}
	return ret, nil
}

func (f *FlagEnv) IsEqualTo(env koyeb.DeploymentEnv) bool {
	return f.key == *env.Key
}

func (f *FlagEnv) UpdateItem(env *koyeb.DeploymentEnv) {
	env.Key = koyeb.PtrString(f.key)
	if f.isSecret {
		env.Secret = koyeb.PtrString(f.value)
		env.Value = nil
	} else {
		env.Secret = nil
		env.Value = koyeb.PtrString(f.value)
	}
}

func (f *FlagEnv) CreateNewItem() *koyeb.DeploymentEnv {
	item := koyeb.NewDeploymentEnvWithDefaults()
	f.UpdateItem(item)
	return item
}
