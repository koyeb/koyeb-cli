package koyeb

import (
	"errors"
	"os"

	homedir "github.com/mitchellh/go-homedir"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	// Used for flags.
	file         string
	cfgFile      string
	apiurl       string
	token        string
	outputFormat string
	debug        bool

	rootCmd = &cobra.Command{
		Use:   "koyeb",
		Short: "Koyeb cli",
	}
	initCmd = &cobra.Command{
		Use:   "init",
		Short: "Init configuration",
		Run:   Init,
	}

	getCmd = &cobra.Command{
		Use:     "get [resource]",
		Aliases: []string{"g", "list"},
		Short:   "Display one or many resources",
	}
	describeCmd = &cobra.Command{
		Use:     "describe [resource]",
		Aliases: []string{"d", "desc", "show"},
		Short:   "Display one resources",
	}
	updateCmd = &cobra.Command{
		Use:     "update [resource]",
		Aliases: []string{"u", "update", "edit"},
		Short:   "Update one resources",
	}
	createCmd = &cobra.Command{
		Use:     "create [resource]",
		Aliases: []string{"c", "new"},
		Short:   "Create a resource from a file",
	}
	deleteCmd = &cobra.Command{
		Use:     "delete [resource]",
		Aliases: []string{"del", "rm"},
		Short:   "Delete resources by name and id",
	}
)

func notImplemented(cmd *cobra.Command, args []string) error {
	return errors.New("Not implemented")
}

func Log(cmd *cobra.Command, args []string) {
	log.Infof("Cmd %v", cmd)
	log.Infof("Cmd has parent %v", cmd.HasParent())
	log.Infof("Cmd parent %v", cmd.Parent())
	log.Infof("Args %v", args)
}

func genericArgs(cmd *cobra.Command, args []string) error {
	if len(args) < 1 {
		return errors.New("requires a resource argument")
	}
	return nil
}

func Run() error {
	return rootCmd.Execute()
}

func er(msg interface{}) {
	log.Errorf("Error: %s", msg)
	os.Exit(1)
}

func init() {
	log.SetFormatter(&log.TextFormatter{})

	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "config file (default is $HOME/.koyeb.yaml)")
	rootCmd.PersistentFlags().StringVarP(&outputFormat, "output", "o", "", "output format (yaml,json,table)")
	rootCmd.PersistentFlags().BoolVarP(&debug, "debug", "d", false, "debug")
	rootCmd.PersistentFlags().String("url", "https://app.koyeb.com", "url of the api")
	rootCmd.PersistentFlags().String("token", "", "API token")
	viper.BindPFlag("url", rootCmd.PersistentFlags().Lookup("url"))
	viper.BindPFlag("token", rootCmd.PersistentFlags().Lookup("token"))
	viper.BindPFlag("debug", rootCmd.PersistentFlags().Lookup("debug"))

	rootCmd.AddCommand(initCmd)

	// Create
	rootCmd.AddCommand(createCmd)
	createCmd.AddCommand(createStackCommand)
	createStackCommand.Flags().StringVarP(&file, "file", "f", "", "Manifest file")
	createStackCommand.MarkFlagRequired("file")
	createCmd.AddCommand(createStoreCommand)
	createStoreCommand.Flags().StringVarP(&file, "file", "f", "", "Manifest file")
	createStoreCommand.MarkFlagRequired("file")
	createCmd.AddCommand(createSecretCommand)
	createSecretCommand.Flags().StringVarP(&file, "file", "f", "", "Manifest file")
	createSecretCommand.MarkFlagRequired("file")

	// Get
	rootCmd.AddCommand(getCmd)
	getCmd.AddCommand(getAllCommand)
	getCmd.AddCommand(getStackCommand)
	getCmd.AddCommand(getStoreCommand)
	getCmd.AddCommand(getSecretCommand)

	// Describe
	rootCmd.AddCommand(describeCmd)
	describeCmd.AddCommand(describeStackCommand)
	describeCmd.AddCommand(describeStoreCommand)
	describeCmd.AddCommand(describeSecretCommand)

	// Update
	rootCmd.AddCommand(updateCmd)
	updateStackCommand.Flags().StringVarP(&file, "file", "f", "", "Manifest file")
	updateCmd.AddCommand(updateStackCommand)
	updateStoreCommand.Flags().StringVarP(&file, "file", "f", "", "Manifest file")
	updateCmd.AddCommand(updateStoreCommand)
	updateSecretCommand.Flags().StringVarP(&file, "file", "f", "", "Manifest file")
	updateCmd.AddCommand(updateSecretCommand)

	// Delete
	rootCmd.AddCommand(deleteCmd)
	deleteCmd.AddCommand(deleteStackCommand)
	deleteCmd.AddCommand(deleteStoreCommand)
	deleteCmd.AddCommand(deleteSecretCommand)

}

func initConfig() {

	if debug {
		log.SetLevel(log.DebugLevel)
	}

	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			er(err)
		}

		viper.AddConfigPath(home)
		viper.SetConfigName(".koyeb")
	}

	viper.AutomaticEnv()

	// TODO check if no .koyeb.yaml or no --config file exists, if not, ask if we want to create a new one

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			log.Error("Configuration not found, use `koyeb init` to create a new one, or use `--config`.")
		} else {
			er(err)
		}
	} else {
		log.Debugf("Using config file: %s", viper.ConfigFileUsed())
	}

	apiurl = viper.GetString("url")
	token = viper.GetString("token")
	debug = viper.GetBool("debug")
}
