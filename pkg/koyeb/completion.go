package koyeb

import (
	"github.com/spf13/cobra"
	"os"
)

var completionCmd = &cobra.Command{
	Use:   "completion [bash|zsh|powershell]",
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
`,
	DisableFlagsInUseLine: true,
	ValidArgs:             []string{"bash", "zsh", "powershell"},
	Args:                  cobra.ExactValidArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		switch args[0] {
		case "bash":
			cmd.Root().GenBashCompletion(os.Stdout)
		case "zsh":
			cmd.Root().GenZshCompletion(os.Stdout)
		case "powershell":
			cmd.Root().GenPowerShellCompletion(os.Stdout)
		}
	},
}
