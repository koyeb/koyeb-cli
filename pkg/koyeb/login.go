package koyeb

import (
	"errors"
	"fmt"
	"os"

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
		log.Fatalf("Unable to read from stdin, please launch the command in interactive mode")
	}

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
	viper.SetConfigPermissions(os.FileMode(0o600))
	err = viper.WriteConfig()
	if err != nil {
		er(err)
	}

	log.Infof("Creating new configuration in %s", configPath)
	return nil
}
