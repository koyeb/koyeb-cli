package koyeb

import (
	"errors"
	"fmt"
	"os"

	koyeb_errors "github.com/koyeb/koyeb-cli/pkg/koyeb/errors"
	"github.com/manifoldco/promptui"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/term"
)

func Login(cmd *cobra.Command, args []string) error {
	configPath := ""
	if cfgFile != "" {
		configPath = cfgFile
	} else {
		home, err := getHomeDir()
		if err != nil {
			return err
		}
		configPath = home + "/.koyeb.yaml"
	}
	viper.SetConfigFile(configPath)

	if !term.IsTerminal(int(os.Stdin.Fd())) {
		return &koyeb_errors.CLIError{
			What:       "Unable to start interactive mode",
			Why:        "the command `koyeb login` requires an interactive terminal.",
			Additional: []string{"Make sure you are not piping the input of the command"},
			Orig:       nil,
			Solution:   "Instead of calling `koyeb login`, create a configuration file manually in ~/.koyeb.yaml",
		}
	}

	if _, err := os.Stat(configPath); !errors.Is(err, os.ErrNotExist) {
		prompt := promptui.Prompt{
			Label:     fmt.Sprintf("Do you want to overwrite your current configuration file (%s)", configPath),
			IsConfirm: true,
		}
		_, err := prompt.Run()
		// If user cancels (ctrl+d, ctrl+c, enter)
		if err != nil {
			return nil
		}
	}

	validate := func(input string) error {
		if len(input) != 64 {
			return errors.New("Invalid api credential")
		}
		return nil
	}

	prompt := promptui.Prompt{
		Label:    "Enter your api access token, you can create a new token here ( https://app.koyeb.com/account/api )",
		Validate: validate,
		Mask:     '*',
	}

	result, err := prompt.Run()
	if err != nil {
		er(err)
	}

	viper.Set("token", result)

	viper.SetConfigType("yaml")
	viper.SetConfigPermissions(os.FileMode(0o600))
	err = viper.WriteConfig()
	if err != nil {
		er(err)
	}

	log.Infof("Creating new configuration in %s", configPath)
	return nil
}
