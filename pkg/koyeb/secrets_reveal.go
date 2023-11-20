package koyeb

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/koyeb/koyeb-cli/pkg/koyeb/errors"
	"github.com/spf13/cobra"
)

func (h *SecretHandler) Reveal(ctx *CLIContext, cmd *cobra.Command, args []string) error {
	secret, err := ResolveSecretArgs(ctx, args[0])
	if err != nil {
		return err
	}

	// RevealSecret require to pass an empty body
	body := make(map[string]interface{})
	_, resp, err := ctx.Client.SecretsApi.RevealSecret(ctx.Context, secret).Body(body).Execute()

	// The field Value of RevealSecretReply is generated from a google.protobuf.Value type which is represented as a
	// map[string]interface{}.
	// The function RevealSecret(...).Execute() returns an error, because it is unable to unmarshal the response body.
	// Here, we only return the error for the case where the response status code is not 200 and compute the secret value
	// from the response body.
	if resp.StatusCode != 200 && err != nil {
		return errors.NewCLIErrorFromAPIError(
			fmt.Sprintf("Error while revealing the secret `%s`", args[0]),
			err,
			resp,
		)
	}

	buffer, err := io.ReadAll(resp.Body)
	if err != nil {
		return &errors.CLIError{
			What: "Error while reading the response body",
			Why:  "the response body could not be read",
			Additional: []string{
				"The Koyeb API to retrieve a secret value returned a response body that could not be read.",
			},
			Orig:     nil,
			Solution: "Try to update the CLI to the latest version. If the problem persists, please create an issue on https://github.com/koyeb/koyeb-cli/issues/new",
		}
	}

	output := map[string]interface{}{}
	if err := json.Unmarshal(buffer, &output); err != nil {
		return &errors.CLIError{
			What: "Error while unmarshalling the response body",
			Why:  "the response body could not be unmarshalled",
			Additional: []string{
				"The Koyeb API to retrieve a secret value returned a response body that could not be unmarshalled.",
			},
			Orig:     nil,
			Solution: "Try to update the CLI to the latest version. If the problem persists, please create an issue on https://github.com/koyeb/koyeb-cli/issues/new",
		}
	}

	if value, ok := output["value"]; ok {
		switch v := value.(type) {
		case map[string]interface{}:
			for key, value := range v {
				fmt.Printf("%s: %v\n", key, value)
			}
			return nil
		case string:
			fmt.Printf("%s\n", v)
			return nil
		default:
			return &errors.CLIError{
				What: "Error while reading the secret value",
				Why:  "the secret value has an unexpected format",
				Additional: []string{
					"The Koyeb API to retrieve a secret value returned a secret type that the CLI could not understand.",
				},
				Orig:     nil,
				Solution: "Try to update the CLI to the latest version. If the problem persists, please create an issue on https://github.com/koyeb/koyeb-cli/issues/new",
			}
		}
	}
	return &errors.CLIError{
		What: "Error while reading the secret value",
		Why:  "the secret value has an unexpected format",
		Additional: []string{
			"The Koyeb API to retrieve a secret value returned a response body that the CLI could not understand.",
		},
		Orig:     nil,
		Solution: "Try to update the CLI to the latest version. If the problem persists, please create an issue on https://github.com/koyeb/koyeb-cli/issues/new",
	}
}
