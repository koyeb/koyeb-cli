package koyeb

import (
	"os"

	"github.com/spf13/cobra"
)

var completionCmd = newCompletionCmd()

// newCompletionCmd builds the completion command as a constructor, like for
// every other command, so that the package-level completionCmd singleton can
// be rebuilt with the state of a fresh process: cobra remembers that a
// command was called for the lifetime of the process, and skipConfigLoading
// relies on that state.
func newCompletionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "completion [bash|zsh|fish|powershell]",
		Short: "Generate completion script",
		Long: `To load completions:

Bash:

$ source <(koyeb completion bash)

# To load completions for each session, execute once:
Linux:
  $ koyeb completion bash > /etc/bash_completion.d/koyeb
MacOS:
  $ koyeb completion bash > /usr/local/etc/bash_completion.d/koyeb

Zsh:

# If shell completion is not already enabled in your environment you will need
# to enable it.  You can execute the following once:

$ echo "autoload -U compinit; compinit" >> ~/.zshrc

# To load completions for each session, execute once:
$ koyeb completion zsh > "${fpath[1]}/_koyeb"

# You will need to start a new shell for this setup to take effect.

Fish:

$ koyeb completion fish | source

# To load completions for each session, execute once:
$ koyeb completion fish > ~/.config/fish/completions/koyeb.fish
`,
		DisableFlagsInUseLine: true,
		ValidArgs:             []string{"bash", "zsh", "fish", "powershell"},
		Args:                  cobra.ExactValidArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			switch args[0] {
			case "bash":
				return cmd.Root().GenBashCompletionV2(os.Stdout, true)
			case "zsh":
				return cmd.Root().GenZshCompletion(os.Stdout)
			case "fish":
				return cmd.Root().GenFishCompletion(os.Stdout, true)
			case "powershell":
				return cmd.Root().GenPowerShellCompletionWithDesc(os.Stdout)
			}
			return nil
		},
	}
}
