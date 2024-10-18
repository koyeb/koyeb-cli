package errors

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"

	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
)

// NewCLIErrorFromAPIError takes an error returned by an API call, for example
// `ctx.Client.SecretsApi.ListSecrets(ctx).Execute()`, and returns a CLIError
// which contains more context about the error.
func NewCLIErrorFromAPIError(what string, err error, resp *http.Response) *CLIError {
	ret := &CLIError{
		What: what,
	}

	if resp != nil {
		switch resp.StatusCode {
		case 429:
			ret.Why = "the Koyeb API returned an error HTTP/429: Too Many Requests because you have exceeded the rate limit"
			ret.Solution = "Please try again in a few seconds."
			return ret
		case 501:
			ret.Why = "the Koyeb API returned an error HTTP/501: Not Implemented because the feature you are trying to use is not yet implemented"
			ret.Solution = "Reach out to Koyeb support for more information."
			if message, ok := getMessage(resp); ok {
				ret.Additional = append(ret.Additional, message)
			}
			return ret
		}
	}

	ret.Orig = err

	var genericErr *koyeb.GenericOpenAPIError
	var unmarshalErr *json.UnmarshalTypeError
	var urlError *url.Error

	if errors.As(err, &genericErr) {
		switch genericErrModel := genericErr.Model().(type) {
		case koyeb.ErrorWithFields:
			ret.Why = fmt.Sprintf("the Koyeb API returned an error %d: %s", *genericErrModel.Status, genericErrModel.GetMessage())
			ret.Solution = SolutionFixRequest
			for _, f := range genericErrModel.GetFields() {
				ret.Additional = append(ret.Additional, fmt.Sprintf("Field %s: %s", f.GetField(), f.GetDescription()))
			}
		case koyeb.Error:
			if genericErrModel.GetStatus() == 401 {
				ret.Why = "your authentication token is invalid or has expired"
				ret.Solution = "Please login again using `koyeb login`, or provide a valid token using the `--token` flag."
				ret.Orig = nil // the original error contains "401 Unauthorized" which is not very useful. Remove it.
			} else {
				ret.Why = fmt.Sprintf("the Koyeb API returned an error %d: %s", *genericErrModel.Status, genericErrModel.GetMessage())
				ret.Solution = SolutionFixRequest
			}
		default:
			if resp != nil {
				ret.Why = fmt.Sprintf("the Koyeb API returned an unexpected error HTTP/%d that the CLI was unable to process, likely due to a bug in the CLI", resp.StatusCode)
				ret.Solution = SolutionUpdateOrIssue
			} else {
				ret.Why = "the Koyeb API returned an unexpected error, not bound to an HTTP response, that the CLI was unable to process, likely due to a bug in the CLI"
				ret.Solution = SolutionUpdateOrIssue
			}
		}
		return ret
	} else if errors.As(err, &unmarshalErr) {
		ret.Why = "the Koyeb API returned an error that the CLI was unable to parse, likely due to a bug in the CLI."
		ret.Solution = SolutionTryAgainOrUpdateOrIssue
	} else if errors.As(err, &urlError) {
		ret.Why = "the CLI was unable to query the Koyeb API because of an issue on your machine or in your configuration"
		ret.Solution = SolutionFixConfig
	} else {
		ret.Why = "the Koyeb API returned an error that the CLI was unable to process, likely due to a bug in the CLI or a problem in your configuration."
		ret.Solution = SolutionTryAgainOrUpdateOrIssue
	}
	return ret
}

func getMessage(r *http.Response) (string, bool) {
	body := map[string]any{}
	if err := json.NewDecoder(r.Body).Decode(&body); err == nil {
		if message, ok := body["message"]; ok {
			if messageStr, ok := message.(string); ok {
				return messageStr, true
			}
		}
	}

	return "", false
}
