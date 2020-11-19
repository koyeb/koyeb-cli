# The Koyeb CLI
The Koyeb CLI (Command Line Interface) is a powerful tool to manage your Koyeb Stacks and Stores directly from your terminal.

# Installation

## Install from release

### Mac OS

```
# Install latest version from github
curl -sL $(curl -s https://api.github.com/repos/koyeb/koyeb-cli/releases/latest | grep 'http.*koyeb-cli-darwin-x86_64' | awk '{print $2}' | sed 's|[\"\,]*||g') -o /usr/local/bin/koyeb

chmod +x /usr/local/bin/koyeb

koyeb init

```

### Linux

```
# Install latest version from github
curl -sL $(curl -s https://api.github.com/repos/koyeb/koyeb-cli/releases/latest | grep 'http.*koyeb-cli-linux-x86_64' | awk '{print $2}' | sed 's|[\"\,]*||g') -o /usr/local/bin/koyeb

chmod +x /usr/local/bin/koyeb

koyeb init

```

## Living at the Edge

To install the `koyeb` binary, simply run:

```shell
go get github.com/koyeb/koyeb-cli/cmd/koyeb
go install github.com/koyeb/koyeb-cli/cmd/koyeb
```

If you need a go environment, follow the [official Go installation documentation](https://golang.org/doc/install).

## Getting started

### Initial configuration

Generate an API token and run `koyeb init` to create a new configuration file.

```shell
➜ koyeb init
? Do you want to create a new configuration file in (/Users/kbot/.koyeb.yaml)? [y/N]
Enter your api credential: ****************************************************************█
INFO[0006] Creating new configuration in /Users/kbot/.koyeb.yaml
```

### Setup completion

```shell
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
```

### General usage

```shell
➜ koyeb --help
Koyeb cli

Usage:
  koyeb [command]

Available Commands:
  create      Create a resource from a file
  delete      Delete resources by name and id
  describe    Display one resources
  get         Display one or many resources
  help        Help about any command
  init        Init configuration
  logs        Get the log of one resources
  run         Launch a new run for a resource
  update      Update one resources

Flags:
  -c, --config string   config file (default is $HOME/.koyeb.yaml)
  -d, --debug           debug
  -h, --help            help for koyeb
  -o, --output string   output format (yaml,json,table)
      --token string    API token
      --url string      url of the api (default "https://app.koyeb.com")

Use "koyeb [command] --help" for more information about a command.
```

## Contribute

If you want to contribute this might be useful

### Regenerate the configuration with the latest API spec

Use: `./scripts/update_api.sh`. This will update the yaml by pulling those in the [api doc](https://developer.koyeb.com/) and regenerate a client from scratch.
We check in the generated files, please make a separate commit with these.
