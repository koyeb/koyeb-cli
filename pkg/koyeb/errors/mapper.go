package errors

import (
	"fmt"
)

// NewCLIErrorForMapperResolve returns a new CLIError when a mapper is unable to resolve an identifier to an object ID.
func NewCLIErrorForMapperResolve(objectType string, objectId string, supportedFormats []string) *CLIError {
	additional := []string{fmt.Sprintf("The supported formats to resolve a %s are:", objectType)}
	for _, e := range supportedFormats {
		additional = append(additional, fmt.Sprintf("* %s", e))
	}

	ret := &CLIError{
		What:       fmt.Sprintf("Unable to find the %s `%s`", objectType, objectId),
		Why:        "no object could be found from the provided identifier",
		Additional: additional,
		Orig:       nil,
		Solution:   CLIErrorSolution(fmt.Sprintf("Provide a valid %s identifier", objectType)),
	}
	return ret
}
