package koyeb

import (
	"errors"
	"fmt"
	"io/fs"
	"os"

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
	file         string
	cfgFile      string
	apiurl       string
	token        string
	outputFormat string
	debug        bool

	rootCmd = &cobra.Command{
		Use:               "koyeb RESOURCE ACTION",
		Short:             "Koyeb CLI",
		DisableAutoGenTag: true,
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

func GetRootCmd() *cobra.Command {
	return rootCmd
}

func Run() error {
	DetectUpdates()
	return rootCmd.Execute()
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

	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "config file (default is $HOME/.koyeb.yaml)")
	rootCmd.PersistentFlags().StringVarP(&outputFormat, "output", "o", "", "output format (yaml,json,table)")
	rootCmd.PersistentFlags().BoolVarP(&debug, "debug", "d", false, "debug")
	rootCmd.PersistentFlags().BoolP("full", "", false, "show full id")
	rootCmd.PersistentFlags().String("url", "https://app.koyeb.com", "url of the api")
	rootCmd.PersistentFlags().String("token", "", "API token")
	viper.BindPFlag("url", rootCmd.PersistentFlags().Lookup("url"))
	viper.BindPFlag("token", rootCmd.PersistentFlags().Lookup("token"))
	viper.BindPFlag("debug", rootCmd.PersistentFlags().Lookup("debug"))

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
	viper.SetEnvPrefix("koyeb")

	if "" != loginCmd.CalledAs() {
		return
	}

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			if viper.GetString("token") != "" {
				log.Debug("Configuration not found, using token from cmdline.")
			} else {
				log.Fatal("Configuration not found, use `koyeb login` to create a new one, or use `--config`.")
			}
		} else if _, ok := err.(*fs.PathError); ok {
			log.Fatal("Configuration not found, use `koyeb login` to create a new one.")
		} else {
			log.Fatal(err)
		}
	} else {
		log.Debugf("Using config file: %s", viper.ConfigFileUsed())
	}

	apiurl = viper.GetString("url")
	token = viper.GetString("token")
	debug = viper.GetBool("debug")
}
