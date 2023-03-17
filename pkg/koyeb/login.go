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

func Login(cmd *cobra.Command, args []string) error {

	configPath := ""
	if cfgFile != "" {
		configPath = cfgFile
	} else {
		home, err := homedir.Dir()
		if err != nil {
			er(err)
		}
		configPath = home + "/.koyeb.yaml"
	}
	viper.SetConfigFile(configPath)

	writeFileMessage := fmt.Sprintf("Do you want to create a new configuration file in (%s)", configPath)
	if _, err := os.Stat(configPath); !errors.Is(err, os.ErrNotExist) {
		writeFileMessage = fmt.Sprintf("Do you want to update configuration file (%s)", configPath)
	}

	prompt := promptui.Prompt{
		Label:     writeFileMessage,
		IsConfirm: true,
	}
	_, err := prompt.Run()
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
	err = viper.WriteConfig()
	if err != nil {
		er(err)
	}

	log.Infof("Creating new configuration in %s", configPath)
	return nil
}
