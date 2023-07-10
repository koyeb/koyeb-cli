package koyeb

import (
	"context"
	"fmt"
	"io/fs"
	"os"

	"github.com/koyeb/koyeb-cli/pkg/koyeb/errors"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/renderer"
	homedir "github.com/mitchellh/go-homedir"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	Version    = "develop"
	BuildDate  = "-"
	Commit     = "-"
	GithubRepo = "koyeb/koyeb-cli"

	// Used for flags.
	cfgFile      string
	apiurl       string
	token        string
	outputFormat renderer.OutputFormat
	debug        bool

	rootCmd = &cobra.Command{
		Use:               "koyeb RESOURCE ACTION",
		Short:             "Koyeb CLI",
		DisableAutoGenTag: true,
		// By default, Cobra prints the error and the command usage when RunE
		// returns an error. This behavior is desirable in case of a user error
		// (unexpected flag provided, for example), but not in case of an
		// runtime errors (API error, for example).
		// To have more control over the error handling, we set SilenceUsage and
		// SilenceErrors.
		SilenceUsage:  true,
		SilenceErrors: true,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			if err := initConfig(); err != nil {
				return err
			}
			DetectUpdates()
			SetupCLIContext(cmd)
			return nil
		},
	}
	loginCmd = &cobra.Command{
		Use:   "login",
		Short: "Login to your Koyeb account",
		RunE:  Login,
	}
	versionCmd = &cobra.Command{
		Use:   "version",
		Short: "Get version",
		Run:   PrintVersion,
	}
)

func Log(cmd *cobra.Command, args []string) {
	log.Infof("Cmd %v", cmd)
	log.Infof("Cmd has parent %v", cmd.HasParent())
	log.Infof("Cmd parent %v", cmd.Parent())
	log.Infof("Args %v", args)
}

func GetRootCmd() *cobra.Command {
	return rootCmd
}

func Run() error {
	ctx := context.Background()
	err := rootCmd.ExecuteContext(ctx)
	if err != nil {
		log.Error(err)
	}
	return err
}

func er(msg interface{}) {
	log.Errorf("Error: %s", msg)
	os.Exit(1)
}

func PrintVersion(cmd *cobra.Command, args []string) {
	fmt.Printf("%s\n", Version)
	log.Debugf("Date: %s", BuildDate)
	log.Debugf("Commit: %s", Commit)
}

func init() {
	log.SetFormatter(&log.TextFormatter{})

	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "config file (default is $HOME/.koyeb.yaml)")
	rootCmd.PersistentFlags().VarP(&outputFormat, "output", "o", "output format (yaml,json,table)")
	rootCmd.PersistentFlags().BoolVarP(&debug, "debug", "d", false, "debug")
	rootCmd.PersistentFlags().BoolP("full", "", false, "show full id")
	rootCmd.PersistentFlags().String("url", "https://app.koyeb.com", "url of the api")
	rootCmd.PersistentFlags().String("token", "", "API token")

	// viper.BindPFlag returns an error only if the second argument is nil, which is never the case here, so we ignore the error
	viper.BindPFlag("url", rootCmd.PersistentFlags().Lookup("url"))     //nolint:errcheck
	viper.BindPFlag("token", rootCmd.PersistentFlags().Lookup("token")) //nolint:errcheck
	viper.BindPFlag("debug", rootCmd.PersistentFlags().Lookup("debug")) //nolint:errcheck

	rootCmd.AddCommand(loginCmd)
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(completionCmd)

	rootCmd.AddCommand(NewSecretCmd())
	rootCmd.AddCommand(NewAppCmd())
	rootCmd.AddCommand(NewDomainCmd())
	rootCmd.AddCommand(NewServiceCmd())
	rootCmd.AddCommand(NewInstanceCmd())
	rootCmd.AddCommand(NewDeploymentCmd())
}

func initConfig() error {
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
			return &errors.CLIError{
				What:       "Error while initializing the CLI",
				Why:        "we were unable to find your home directory",
				Additional: nil,
				Orig:       err,
				Solution:   "Please provide a config file with the --config flag, or set the $HOME environment variable",
			}
		}

		viper.AddConfigPath(home)
		viper.SetConfigName(".koyeb")
	}

	viper.AutomaticEnv()
	viper.SetEnvPrefix("koyeb")

	if "" != loginCmd.CalledAs() || "" != versionCmd.CalledAs() {
		return nil
	}

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			if viper.GetString("token") != "" {
				log.Debug("Configuration not found, using token from cmdline.")
			} else {
				return &errors.CLIError{
					What: "Error while initializing the CLI",
					Why:  "we were unable to find your configuration file",
					Additional: []string{
						"The configuration file is usually located in $HOME/.koyeb.yaml",
					},
					Orig:     err,
					Solution: "Use `koyeb login` to create a new configuration file, or provide an existing one with the --config flag. If you provided a configuration file, make sure the $HOME environment variable is set correctly. If you don't want to use a configuration file, you can set the --token flag to your API token.",
				}
			}
		} else if _, ok := err.(*fs.PathError); ok {
			return &errors.CLIError{
				What:       "Error while initializing the CLI",
				Why:        "we were unable to load your configuration file",
				Additional: []string{"You provided a configuration file, but we couldn't load it."},
				Orig:       err,
				Solution:   "Make sure the configuration file exists and is readable.",
			}
		} else if _, ok := err.(viper.UnsupportedConfigError); ok {
			return &errors.CLIError{
				What:       "Error while initializing the CLI",
				Why:        "the configuration file format is not supported",
				Additional: nil,
				Orig:       err,
				Solution:   "Change the name of the configuration file to add the .yaml extension. If you don't want to use a configuration file, you can set the --token flag to your API token.",
			}
		} else {
			return &errors.CLIError{
				What: "Error while initializing the CLI",
				Why:  "we were unable to load your configuration file",
				Additional: []string{
					"The configuration file exists and is readable, but we couldn't load it.",
				},
				Orig:     err,
				Solution: "Make sure the configuration file is a valid YAML file.",
			}
		}
	} else {
		log.Debugf("Using config file: %s", viper.ConfigFileUsed())
	}

	apiurl = viper.GetString("url")
	token = viper.GetString("token")
	debug = viper.GetBool("debug")
	return nil
}
