package koyeb

import (
	"github.com/koyeb/koyeb-cli/pkg/koyeb/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func (h *OrganizationHandler) Switch(ctx *CLIContext, cmd *cobra.Command, args []string) error {
	organization, err := ResolveOrganizationArgs(ctx, args[0])
	if err != nil {
		return err
	}
	viper.Set("organization", organization)

	err = viper.WriteConfig()
	if err != nil {
		return &errors.CLIError{
			What: "Unable to switch the current organization",
			Why:  "we were unable to write the configuration file",
			Additional: []string{
				"The command `koyeb organization switch` needs to update your configuration file, usually located in $HOME/.koyeb.yaml",
				"If you do not have write access to this file, you can use the --config flag to specify a different location.",
				"Alternatively, you can manually edit the configuration file and set the organization field to the organization ID you want to use.",
				"You can also provide the organization UUID with the --organization flag.",
			},
			Orig:     err,
			Solution: "Fix the issue preventing the CLI to write the configuration file, or manually edit the configuration file",
		}
	}
	return nil
}
