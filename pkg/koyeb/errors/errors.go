package errors

import (
	"bytes"
	"text/template"
)

type CLIErrorSolution string

// CLIError represents a user-friendly error that can be displayed to the user.
// It follows the pattern described in this blog post:
// https://wix-ux.com/when-life-gives-you-lemons-write-better-error-messages-46c5223e1a2f
//
// The idea is to provide a maximum of information to the user when an error
// happens: what exactly caused the error? Why did it happen? How to solve it?
type CLIError struct {
	What       string           // What was the user doing when the error happened. For example: "creating an app"
	Why        string           // Why did the error happen. For example: "API returned an error"
	Additional []string         // Additional information to display to the user. For example: "the field 'name' is required"
	Orig       error            // Original error
	Solution   CLIErrorSolution // How to solve the error. For example: "update the CLI"
}

const tmplError = `⚠️  {{.What}}: {{.Why}} ⚠️️
{{if .Additional}}
🔎 Additional details
{{range .Additional}}{{.}}
{{end}}{{end}}
👨‍⚕️ How to solve the issue?
{{.Solution}}{{if .Orig}}

🕦 The original error was:
{{.Orig.Error}}{{end}}
`

var (
	tpl = template.Must(template.New("error").Parse(tmplError))
)

func (e *CLIError) Error() string {
	var buf bytes.Buffer

	err := tpl.Execute(&buf, *e)
	// This should never happen, as the template is hardcoded in the source code.
	if err != nil {
		panic(err)
	}
	return buf.String()
}

const (
	solutionTryAgainOrUpdateOrIssue CLIErrorSolution = "Please try again, and if the problem persists, try upgrading to the latest version of the CLI. If the problem still persists, please open an issue at https://github.com/koyeb/koyeb-cli/issues/new and include the output of the command you ran with the --debug flag enabled."
	solutionUpdateOrIssue           CLIErrorSolution = "Please try upgrading to the latest version of the CLI. If the problem still persists, please open an issue at https://github.com/koyeb/koyeb-cli/issues/new and include the output of the command you ran with the --debug flag enabled."
	solutionFixRequest              CLIErrorSolution = "Fix the request, and try again"
	solutionFixConfig               CLIErrorSolution = "Fix your configuration and try again"
)
