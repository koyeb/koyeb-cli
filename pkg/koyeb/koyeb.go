package koyeb

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"runtime"

	koyeb_errors "github.com/koyeb/koyeb-cli/pkg/koyeb/errors"
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
	apiurl       = "https://app.koyeb.com"
	cfgFile      string
	token        string
	outputFormat renderer.OutputFormat
	forceASCII   bool
	debugFull    bool
	debug        bool
	organization string

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

func isHelpCalled(rootCmd *cobra.Command) bool {
	for _, subcmd := range rootCmd.Commands() {
		if subcmd.Name() == "help" {
			return subcmd.CalledAs() != ""
		}
	}
	return false
}

func skipConfigLoading(rootCmd *cobra.Command) bool {
	return loginCmd.CalledAs() != "" || versionCmd.CalledAs() != "" ||
		completionCmd.CalledAs() != "" || isHelpCalled(rootCmd)
}

func GetRootCommand() *cobra.Command {
	rootCmd := &cobra.Command{
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
			if err := initConfig(cmd.Root()); err != nil {
				return err
			}
			DetectUpdates()
			return SetupCLIContext(cmd, organization)
		},
	}

	log.SetFormatter(&log.TextFormatter{})

	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "config file (default is $HOME/.koyeb.yaml)")
	rootCmd.PersistentFlags().VarP(&outputFormat, "output", "o", "output format (yaml,json,table)")
	rootCmd.PersistentFlags().BoolVarP(&debug, "debug", "d", false, "enable the debug output")
	rootCmd.PersistentFlags().BoolVar(&debugFull, "debug-full", false, "do not hide sensitive information (tokens) in the debug output")
	rootCmd.PersistentFlags().BoolVar(&forceASCII, "force-ascii", false, "only output ascii characters (no unicode emojis)")
	rootCmd.PersistentFlags().BoolP("full", "", false, "do not truncate output")
	rootCmd.PersistentFlags().String("url", "https://app.koyeb.com", "url of the api")
	rootCmd.PersistentFlags().String("token", "", "API token")
	rootCmd.PersistentFlags().StringVar(&organization, "organization", "", "organization ID")

	// viper.BindPFlag returns an error only if the second argument is nil, which is never the case here, so we ignore the error
	viper.BindPFlag("url", rootCmd.PersistentFlags().Lookup("url"))                   //nolint:errcheck
	viper.BindPFlag("token", rootCmd.PersistentFlags().Lookup("token"))               //nolint:errcheck
	viper.BindPFlag("debug", rootCmd.PersistentFlags().Lookup("debug"))               //nolint:errcheck
	viper.BindPFlag("organization", rootCmd.PersistentFlags().Lookup("organization")) //nolint:errcheck

	rootCmd.AddCommand(loginCmd)
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(completionCmd)

	rootCmd.AddCommand(NewOrganizationCmd())
	rootCmd.AddCommand(NewSecretCmd())
	rootCmd.AddCommand(NewAppCmd())
	rootCmd.AddCommand(NewDomainCmd())
	rootCmd.AddCommand(NewServiceCmd())
	rootCmd.AddCommand(NewInstanceCmd())
	rootCmd.AddCommand(NewDeploymentCmd())
	rootCmd.AddCommand(NewRegionalDeploymentCmd())
	rootCmd.AddCommand(NewDatabaseCmd())
	rootCmd.AddCommand(NewMetricsCmd())
	rootCmd.AddCommand(NewArchiveCmd())
	rootCmd.AddCommand(NewDeployCmd())
	rootCmd.AddCommand(NewVolumeCmd())
	rootCmd.AddCommand(NewSnapshotCmd())
	rootCmd.AddCommand(NewComposeCmd())
	rootCmd.AddCommand(NewSandboxCmd())
	return rootCmd
}

func Run() error {
	defer func() {
		if r := recover(); r != nil {
			stacktrace := make([]byte, 4096)
			size := runtime.Stack(stacktrace, false)
			fmt.Fprintf(os.Stderr, "%s", &koyeb_errors.CLIError{
				What: "An unexpected error occured",
				Why:  "it's maybe our fault, or maybe not, we can't tell",
				Additional: []string{
					"The CLI should have handled this error gracefully, but it didn't.",
					"It might be a problem with the CLI itself, with the API, or with your configuration. Unfortuatey, we can't tell.",
					"In any case, we would love to hear about it so we can handle this error gracefully in a future version of the CLI.",
					fmt.Sprintf("\nStacktrace:\n-----------\n%s-----------", stacktrace[:size]),
				},
				Orig:     fmt.Errorf("%s", r),
				Solution: "Please open an issue at https://github.com/koyeb/koyeb-cli/issues/new and provide the command you ran, the error message, and the output of `koyeb version`",
			})
		}
	}()

	rootCmd := GetRootCommand()

	ctx := context.Background()
	err := rootCmd.ExecuteContext(ctx)
	if err != nil {
		var cliErr *koyeb_errors.CLIError

		// All the commands should return a CLIError. The only case where we get
		// a different error is when the command is not found, or when the user
		// provides an invalid flag.
		if !errors.As(err, &cliErr) {
			err = &koyeb_errors.CLIError{
				What:       "Invalid command",
				Why:        "the parameters you provided are invalid",
				Additional: nil,
				Orig:       err,
				Solution:   "Run `koyeb --help` to see the list of available commands, or `koyeb <command> --help` to see the help of a specific command",
				ASCII:      forceASCII,
			}
		} else {
			cliErr.ASCII = forceASCII
		}

		fmt.Fprintf(os.Stderr, "%s", err)
	}
	return err
}

func PrintVersion(cmd *cobra.Command, args []string) {
	fmt.Printf("%s\n", Version)
	log.Debugf("Date: %s", BuildDate)
	log.Debugf("Commit: %s", Commit)
}

// getHomeDir returns the home directory of the user, and a meaningful error in case of failure.
func getHomeDir() (string, error) {
	home, err := homedir.Dir()
	if err != nil {
		return "", &koyeb_errors.CLIError{
			What:       "Error finding your home directory",
			Why:        "we were unable to find your home directory",
			Additional: nil,
			Orig:       err,
			Solution:   "Please provide a config file with the --config flag, or set the $HOME environment variable",
		}
	}
	return home, nil
}

func initConfig(rootCmd *cobra.Command) error {
	if debug {
		log.SetLevel(log.DebugLevel)
	}

	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := getHomeDir()
		if err != nil {
			return err
		}
		viper.AddConfigPath(home)
		viper.SetConfigName(".koyeb")
	}

	viper.AutomaticEnv()
	viper.SetEnvPrefix("koyeb")

	if skipConfigLoading(rootCmd) {
		return nil
	}

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			if viper.GetString("token") != "" {
				log.Debug("Configuration not found, using token from cmdline.")
			} else {
				return &koyeb_errors.CLIError{
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
			return &koyeb_errors.CLIError{
				What:       "Error while initializing the CLI",
				Why:        "we were unable to load your configuration file",
				Additional: []string{"You provided a configuration file, but we couldn't load it."},
				Orig:       err,
				Solution:   "Make sure the configuration file exists and is readable.",
			}
		} else if _, ok := err.(viper.UnsupportedConfigError); ok {
			return &koyeb_errors.CLIError{
				What:       "Error while initializing the CLI",
				Why:        "the configuration file format is not supported",
				Additional: nil,
				Orig:       err,
				Solution:   "Change the name of the configuration file to add the .yaml extension. If you don't want to use a configuration file, you can set the --token flag to your API token.",
			}
		} else {
			return &koyeb_errors.CLIError{
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
	organization = viper.GetString("organization")
	return nil
}
