package koyeb

import (
	"bufio"
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/errors"
	"github.com/manifoldco/promptui"
	stderrors "github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

type SecretType string

const (
	SecretTypeSimple               SecretType = "simple"
	SecretTypeRegistryDockerHub    SecretType = "registry-dockerhub"
	SecretTypeRegistryPrivate      SecretType = "registry-private"
	SecretTypeRegistryDigitalOcean SecretType = "registry-digital-ocean"
	SecretTypeRegistryGitlab       SecretType = "registry-gitlab"
	SecretTypeRegistryGCP          SecretType = "registry-gcp"
	SecretTypeRegistryAzure        SecretType = "registry-azure"
)

// Implement the pflag.Value interface to parse the --type flag.
func (f *SecretType) String() string {
	return string(*f)
}

// Implement the pflag.Value interface to parse the --type flag.
func (f *SecretType) Type() string {
	return "type"
}

// SecretTypeAllValues returns all the possible values for a secret type as strings.
func SecretTypeAllValues() []string {
	return []string{
		string(SecretTypeSimple),
		string(SecretTypeRegistryDockerHub),
		string(SecretTypeRegistryPrivate),
		string(SecretTypeRegistryDigitalOcean),
		string(SecretTypeRegistryGitlab),
		string(SecretTypeRegistryGCP),
		string(SecretTypeRegistryAzure),
	}
}

// Implement the pflag.Value interface to parse the --type flag.
func (f *SecretType) Set(param string) error {
	values := SecretTypeAllValues()
	for _, value := range values {
		if param == value {
			*f = SecretType(param)
			return nil
		}
	}
	return stderrors.New(fmt.Sprintf("invalid secret type. Valid values are: %s", strings.Join(values, ", ")))
}

func NewSecretCmd() *cobra.Command {
	h := NewSecretHandler()

	secretCmd := &cobra.Command{
		Use:     "secrets ACTION",
		Aliases: []string{"sec", "secret"},
		Short:   "Secrets",
	}

	flagSecretType := SecretTypeSimple // default value
	createSecretCmd := &cobra.Command{
		Use:   "create NAME",
		Short: "Create secret",
		Args:  cobra.ExactArgs(1),
		RunE: WithCLIContext(func(ctx *CLIContext, cmd *cobra.Command, args []string) error {
			secret := koyeb.NewCreateSecretWithDefaults()
			secret.SetName(args[0])

			var err error
			var abort bool
			switch flagSecretType {
			// --type=simple
			case SecretTypeSimple:
				secret.SetType(koyeb.SECRETTYPE_SIMPLE)
				var value string
				value, abort, err = getSecretValue(cmd.Flags())
				secret.SetValue(value)
			// --type=registry-*
			case SecretTypeRegistryDockerHub:
				secret.SetType(koyeb.SECRETTYPE_REGISTRY)
				cfg := koyeb.NewDockerHubRegistryConfigurationWithDefaults()
				abort, err = parseRegistry(cmd.Flags(), cfg)
				secret.SetDockerHubRegistry(*cfg)
			case SecretTypeRegistryPrivate:
				secret.SetType(koyeb.SECRETTYPE_REGISTRY)
				cfg := koyeb.NewPrivateRegistryConfigurationWithDefaults()
				abort, err = parsePrivateRegistry(cmd.Flags(), cfg)
				secret.SetPrivateRegistry(*cfg)
			case SecretTypeRegistryDigitalOcean:
				secret.SetType(koyeb.SECRETTYPE_REGISTRY)
				cfg := koyeb.NewDigitalOceanRegistryConfigurationWithDefaults()
				abort, err = parseRegistry(cmd.Flags(), cfg)
				secret.SetDigitalOceanRegistry(*cfg)
			case SecretTypeRegistryGitlab:
				secret.SetType(koyeb.SECRETTYPE_REGISTRY)
				cfg := koyeb.NewGitLabRegistryConfigurationWithDefaults()
				abort, err = parseRegistry(cmd.Flags(), cfg)
				secret.SetGitlabRegistry(*cfg)
			case SecretTypeRegistryGCP:
				secret.SetType(koyeb.SECRETTYPE_REGISTRY)
				cfg := koyeb.NewGCPContainerRegistryConfigurationWithDefaults()
				err = parseGCPRegistry(cmd.Flags(), cfg)
				secret.SetGcpContainerRegistry(*cfg)
			case SecretTypeRegistryAzure:
				secret.SetType(koyeb.SECRETTYPE_REGISTRY)
				cfg := koyeb.NewAzureContainerRegistryConfigurationWithDefaults()
				abort, err = parseAzureRegistry(cmd.Flags(), cfg)
				secret.SetAzureContainerRegistry(*cfg)
			default:
				panic("Unkown secret type:" + flagSecretType)
			}
			if abort || err != nil {
				return err
			}
			return h.Create(ctx, cmd, args, secret)
		}),
	}
	createSecretCmd.Flags().Var(&flagSecretType, "type", fmt.Sprintf("Secret type (%s)", strings.Join(SecretTypeAllValues(), ", ")))
	addSecretFlags(createSecretCmd.Flags(), &flagSecretType)
	secretCmd.AddCommand(createSecretCmd)

	getSecretCmd := &cobra.Command{
		Use:   "get NAME",
		Short: "Get secret",
		Args:  cobra.ExactArgs(1),
		RunE:  WithCLIContext(h.Get),
	}
	secretCmd.AddCommand(getSecretCmd)

	listSecretCmd := &cobra.Command{
		Use:   "list",
		Short: "List secrets",
		RunE:  WithCLIContext(h.List),
	}
	secretCmd.AddCommand(listSecretCmd)

	describeSecretCmd := &cobra.Command{
		Use:   "describe NAME",
		Short: "Describe secret",
		Args:  cobra.ExactArgs(1),
		RunE:  WithCLIContext(h.Describe),
	}
	secretCmd.AddCommand(describeSecretCmd)

	updateSecretCmd := &cobra.Command{
		Use:   "update NAME",
		Short: "Update secret",
		Args:  cobra.ExactArgs(1),
		RunE: WithCLIContext(func(ctx *CLIContext, cmd *cobra.Command, args []string) error {
			secretID, err := ResolveSecretArgs(ctx, args[0])
			if err != nil {
				return err
			}

			res, resp, err := ctx.Client.SecretsApi.GetSecret(ctx.Context, secretID).Execute()
			if err != nil {
				return errors.NewCLIErrorFromAPIError(
					fmt.Sprintf("Error while creating the secret `%s`", args[0]),
					err,
					resp,
				)
			}

			abort := false
			switch res.Secret.GetType() {
			case koyeb.SECRETTYPE_SIMPLE:
				var value string
				value, abort, err = getSecretValue(cmd.Flags())
				res.Secret.SetValue(value)
			case koyeb.SECRETTYPE_REGISTRY:
				if registry, ok := res.Secret.GetDockerHubRegistryOk(); ok {
					abort, err = parseRegistry(cmd.Flags(), registry)
				} else if registry, ok := res.Secret.GetPrivateRegistryOk(); ok {
					abort, err = parseRegistry(cmd.Flags(), registry)
				} else if registry, ok := res.Secret.GetDigitalOceanRegistryOk(); ok {
					abort, err = parseRegistry(cmd.Flags(), registry)
				} else if registry, ok := res.Secret.GetGitlabRegistryOk(); ok {
					abort, err = parseRegistry(cmd.Flags(), registry)
				} else if registry, ok := res.Secret.GetGcpContainerRegistryOk(); ok {
					err = parseGCPRegistry(cmd.Flags(), registry)
				} else if registry, ok := res.Secret.GetAzureContainerRegistryOk(); ok {
					abort, err = parseAzureRegistry(cmd.Flags(), registry)
				}
			default:
				panic("Unkown secret type: " + res.Secret.GetType())
			}

			if abort || err != nil {
				return err
			}

			return h.Update(ctx, cmd, args, res.Secret)
		}),
	}
	addSecretFlags(updateSecretCmd.Flags(), &flagSecretType)
	secretCmd.AddCommand(updateSecretCmd)

	deleteSecretCmd := &cobra.Command{
		Use:   "delete NAME",
		Short: "Delete secret",
		Args:  cobra.ExactArgs(1),
		RunE:  WithCLIContext(h.Delete),
	}
	secretCmd.AddCommand(deleteSecretCmd)

	revealSecretCmd := &cobra.Command{
		Use:     "reveal NAME",
		Aliases: []string{"show"},
		Short:   "Show secret value",
		Args:    cobra.ExactArgs(1),
		RunE:    WithCLIContext(h.Reveal),
	}
	secretCmd.AddCommand(revealSecretCmd)

	return secretCmd
}

func addSecretFlags(flags *pflag.FlagSet, secretType *SecretType) {
	flags.StringP("value", "v", "", "Secret Value")
	flags.Bool("value-from-stdin", false, "Secret Value from stdin")
	flags.String("registry-username", "", "Registry username. Only valid with --type=registry-*")
	flags.String("registry-url", "", "Registry URL. Only valid with --type=registry-private and --type=registry-gcp, otherwise ignored")
	flags.String("registry-keyfile", "", "Registry URL. Only valid with --type=registry-gcp, otherwise ignored")
	flags.String("registry-name", "", "Registry name. Only valid with --type=registry-azure, otherwise ignored")
}

// getSecretValue parses --value of stdin if --value-from-stdin is set. The
// bool returned is true if the user canceled the prompt (ie: pressed Ctrl+C).
func getSecretValue(flags *pflag.FlagSet) (string, bool, error) {
	if flags.Lookup("value-from-stdin").Changed && flags.Lookup("value").Changed {
		return "", false, &errors.CLIError{
			What:       "Invalid arguments to create a secret",
			Why:        "you can't provide both --value and --value-from-stdin at the same time",
			Additional: nil,
			Orig:       nil,
			Solution:   "Remove one of the flags",
		}
	}
	if flags.Lookup("value").Changed {
		password, _ := flags.GetString("value")
		return password, false, nil
	} else if flags.Lookup("value-from-stdin").Changed {
		var input []string

		scanner := bufio.NewScanner(os.Stdin)
		for {
			scanner.Scan()
			text := scanner.Text()
			if len(text) != 0 {
				input = append(input, text)
			} else {
				break
			}
		}
		return strings.Join(input, "\n"), false, nil
	}
	prompt := promptui.Prompt{
		Label: "Enter your secret",
		Mask:  '*',
	}

	result, err := prompt.Run()
	// When user cancels the prompt, we return nil to cancel the command
	if err != nil {
		return "", true, nil
	}
	return result, false, nil
}

// Interface implemented by all the registries expecting a username and a password (ie: all except GCP).
type RegistryWithUsernameAndPassword interface {
	SetUsername(string)
	SetPassword(string)
}

// parseUsernameAndPasswordFlags parses the flags --registry-username and prompt the user for the password
func parseUsernameAndPasswordFlags(flags *pflag.FlagSet, registry RegistryWithUsernameAndPassword) (bool, error) {
	if !flags.Lookup("registry-username").Changed {
		return false, &errors.CLIError{
			What:       "Invalid arguments",
			Why:        "the argument --registry-username is required",
			Additional: nil,
			Orig:       nil,
			Solution:   "Provide the flag --registry-username and try again",
		}
	}
	username, _ := flags.GetString("registry-username")
	registry.SetUsername(username)

	password, abort, err := getSecretValue(flags)
	if abort || err != nil {
		return abort, err
	}
	if password == "" {
		return false, &errors.CLIError{
			What: "Invalid arguments",
			Why:  "the argument --value or --value-from-stdin is required",
			Additional: []string{
				"You must define a password for the registry secret",
				"Use --value=<password> to provide the password directly,",
				"or --value-from-stdin to read the password from stdin",
			},
			Orig:     nil,
			Solution: "Provide the required flag and try again",
		}
	}
	registry.SetPassword(password)
	return false, nil
}

// Interface implemented by all the registries expecting a URL.
type RegistryWithURL interface {
	SetUrl(string)
}

// parseURLFlag parses the flag --registry-url.
func parseURLFlag(flags *pflag.FlagSet, registry RegistryWithURL) error {
	if !flags.Lookup("registry-url").Changed {
		return &errors.CLIError{
			What:       "Invalid arguments",
			Why:        "the argument --registry-url is required",
			Additional: nil,
			Orig:       nil,
			Solution:   "Provide the flag --registry-url and try again",
		}
	}

	url, _ := flags.GetString("registry-url")
	registry.SetUrl(url)
	return nil
}

// parseRegistry is the generic function to parse the flags specific to --type=registry-* (except GCP, Azure and Private).
func parseRegistry(flags *pflag.FlagSet, registry RegistryWithUsernameAndPassword) (bool, error) {
	return parseUsernameAndPasswordFlags(flags, registry)
}

// parsePrivateRegistry parses the flags specific to --type=registry-private.
func parsePrivateRegistry(flags *pflag.FlagSet, registry *koyeb.PrivateRegistryConfiguration) (bool, error) {
	if err := parseURLFlag(flags, registry); err != nil {
		return false, err
	}
	return parseUsernameAndPasswordFlags(flags, registry)
}

// parseGCPRegistry parses the flags specific to --type=registry-gcp.
func parseGCPRegistry(flags *pflag.FlagSet, registry *koyeb.GCPContainerRegistryConfiguration) error {
	if err := parseURLFlag(flags, registry); err != nil {
		return err
	}

	if !flags.Lookup("registry-keyfile").Changed {
		return &errors.CLIError{
			What:       "Invalid arguments",
			Why:        "the argument --registry-keyfile is required",
			Additional: nil,
			Orig:       nil,
			Solution:   "Provide the flag --registry-keyfile and try again",
		}
	}

	path, _ := flags.GetString("registry-keyfile")
	keyfile, err := os.Open(path)
	if err != nil {
		return &errors.CLIError{
			What:       "Problem with the keyfile",
			Why:        "unable to open the keyfile provided in --registry-keyfile",
			Additional: nil,
			Orig:       err,
			Solution:   "Make sure the file exists and is readable and try again",
		}
	}
	defer keyfile.Close()

	data, err := io.ReadAll(keyfile)
	if err != nil {
		return &errors.CLIError{
			What:       "Problem with the keyfile",
			Why:        "unable to read the keyfile provided in --registry-keyfile",
			Additional: nil,
			Orig:       err,
			Solution:   "Make sure the file exists and is readable and try again",
		}
	}
	registry.SetKeyfileContent(base64.StdEncoding.EncodeToString(data))
	return nil
}

// parseAzureRegistry parses the flags specific to --type=registry-azure.
func parseAzureRegistry(flags *pflag.FlagSet, registry *koyeb.AzureContainerRegistryConfiguration) (bool, error) {
	if !flags.Lookup("registry-name").Changed {
		return false, &errors.CLIError{
			What:       "Invalid arguments",
			Why:        "the argument --registry-name is required",
			Additional: nil,
			Orig:       nil,
			Solution:   "Provide the flag --registry-name and try again",
		}
	}
	return parseUsernameAndPasswordFlags(flags, registry)
}

func NewSecretHandler() *SecretHandler {
	return &SecretHandler{}
}

type SecretHandler struct {
}

func ResolveSecretArgs(ctx *CLIContext, val string) (string, error) {
	secretMapper := ctx.Mapper.Secret()
	id, err := secretMapper.ResolveID(val)
	if err != nil {
		return "", err
	}
	return id, nil
}
