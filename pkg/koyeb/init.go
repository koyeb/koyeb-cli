package koyeb

import (
	"errors"
	"fmt"
	"os"

	"github.com/manifoldco/promptui"
	homedir "github.com/mitchellh/go-homedir"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func Init(cmd *cobra.Command, args []string) {

	home, err := homedir.Dir()
	if err != nil {
		er(err)
	}
	configPath := home + "/.koyeb.yaml"

	prompt := promptui.Prompt{
		Label:     fmt.Sprintf("Do you want to create a new configuration file in (%s)", configPath),
		IsConfirm: true,
	}
	_, err = prompt.Run()
	if err != nil {
		os.Exit(1)
	}

	validate := func(input string) error {
		if len(input) != 64 {
			return errors.New("Invalid api credential")
		}
		return nil
	}

	prompt = promptui.Prompt{
		Label:    "Enter your api credential",
		Validate: validate,
		Mask:     '*',
	}

	result, err := prompt.Run()
	if err != nil {
		er(err)
	}

	viper.Set("token", result)

	viper.SetConfigType("yaml")
	err = viper.SafeWriteConfig()
	if err != nil {
		er(err)
	}

	log.Infof("Creating new configuration in %s", configPath)
}
