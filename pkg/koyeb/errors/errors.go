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
	ASCII      bool             // Whether to use only ASCII characters in the error message
}

func (e *CLIError) Error() string {
	var buf bytes.Buffer
	var tmplError string

	if e.ASCII {
		tmplError = `!!! {{.What}}: {{.Why}}
{{if .Additional}}
> Additional details
{{range .Additional}}{{.}}
{{end}}
{{- end}}
> How to solve the issue?
{{.Solution}}
{{- if notNil .Orig}}

> The original error was:
{{.Orig.Error}}
{{- end}}
`
	} else {
		tmplError = `❌ {{.What}}: {{.Why}}
{{if .Additional}}
🔎 Additional details
{{range .Additional}}{{.}}
{{end}}
{{- end}}
🏥 How to solve the issue?
{{.Solution}}
{{- if notNil .Orig}}

🕦 The original error was:
{{.Orig.Error}}
{{- end}}
`
	}

	tpl := template.Must(template.New("error").Funcs(
		template.FuncMap{
			// The `notNil` function allows distinguishing between nil and empty errors.
			//
			// It serves the purpose of selectively hiding the "original error" section when the original error is nil,
			// while still displaying it if the original error is an empty string.
			//
			// Consider the following scenario:
			// type customErr string
			// func (e customErr) Error() string { return "the error was:" + string(e) }
			// CLIError{..., Orig: customErr("")}
			//
			// In this case, if we rely solely on `{{if .Orig}}` without utilizing `notNil`, the section would be hidden
			// because the standard library would interpret the error as an empty string.
			//
			// This distinction is particularly relevant in cases like `viper.UnsupportedConfigError`.
			"notNil": func(e error) bool { return e != nil },
		}).Parse(tmplError))

	err := tpl.Execute(&buf, *e)
	// This should never happen, as the template is hardcoded in the source code.
	if err != nil {
		panic(err)
	}
	return buf.String()
}

const (
	SolutionTryAgainOrUpdateOrIssue CLIErrorSolution = "Please try again, and if the problem persists, try upgrading to the latest version of the CLI. If the problem still persists, please open an issue at https://github.com/koyeb/koyeb-cli/issues/new and include the output of the command you ran with the --debug flag enabled."
	SolutionUpdateOrIssue           CLIErrorSolution = "Please try upgrading to the latest version of the CLI. If the problem still persists, please open an issue at https://github.com/koyeb/koyeb-cli/issues/new and include the output of the command you ran with the --debug flag enabled."
	SolutionFixRequest              CLIErrorSolution = "Fix the request, and try again"
	SolutionFixConfig               CLIErrorSolution = "Fix your configuration and try again"
)
