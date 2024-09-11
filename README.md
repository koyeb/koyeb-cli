![Build](https://github.com/koyeb/koyeb-cli/workflows/Release/badge.svg)

# Koyeb CLI

The Koyeb CLI (Command Line Interface) is a powerful tool to manage your Koyeb serverless infrastructure directly from your terminal.

## Installation

### Download from GitHub

The CLI can be installed from pre-compiled binaries for macOS (darwin), Linux and Windows. You can download the appropriate version from the [Releases](https://github.com/koyeb/koyeb-cli/releases) page.

### MacOS

You can install the latest version of the Koyeb CLI on macOS using [Homebrew](http://brew.sh/):

```shell
brew install koyeb/tap/koyeb
```

You can upgrade an existing installation of the Koyeb CLI running:

```
brew upgrade koyeb
```

### Living at the Edge

To install the latest `koyeb` binary with go, simply run:

```shell
go get github.com/koyeb/koyeb-cli/cmd/koyeb
go install github.com/koyeb/koyeb-cli/cmd/koyeb
```

If you need a go environment, follow the [official Go installation documentation](https://golang.org/doc/install).


## Getting started

### Initial configuration

Generate an API token and run `koyeb login` to create a new configuration file.

```shell
➜ koyeb login
? Do you want to create a new configuration file in (/Users/kbot/.koyeb.yaml)? [y/N]
✗ Enter your personal access token. You can create a new token here (https://app.koyeb.com/user/settings/api/): *************
INFO[0006] Creating new configuration in /Users/kbot/.koyeb.yaml
```

### General usage

```shell
➜ koyeb --help
Koyeb CLI

Usage:
  koyeb [command]

Available Commands:
  apps                 Apps
  archives             Archives
  completion           Generate completion script
  databases            Databases
  deploy               Deploy a directory to Koyeb
  deployments          Deployments
  domains              Domains
  help                 Help about any command
  instances            Instances
  login                Login to your Koyeb account
  metrics              Metrics
  organizations        Organization
  regional-deployments Regional deployments
  secrets              Secrets
  services             Services
  version              Get version
  volumes              Manage persistent volumes

Flags:
  -c, --config string         config file (default is $HOME/.koyeb.yaml)
  -d, --debug                 enable the debug output
      --debug-full            do not hide sensitive information (tokens) in the debug output
      --force-ascii           only output ascii characters (no unicode emojis)
      --full                  do not truncate output
  -h, --help                  help for koyeb
      --organization string   organization ID
  -o, --output output         output format (yaml,json,table)
      --token string          API token
      --url string            url of the api (default "https://app.koyeb.com")

Use "koyeb [command] --help" for more information about a command.
```


### Enabling shell auto-completion

`koyeb` has auto-completion support for `bash`, `zsh` and `fish`.

#### Bash

You can easily do `source <(koyeb completion bash)` to add completion to your current Bash session.

To load completions for all sessions, simply add the auto-completion script to your `bash_completion.d` folder.

On Linux:

```shell
koyeb completion bash > /etc/bash_completion.d/koyeb
```

On MacOs:

```shell
koyeb completion bash > /usr/local/etc/bash_completion.d/koyeb
```

You will need to start a new shell for this setup to take effect.

#### Zsh

If shell completion is not already enabled in your environment you will need to enable it.  You can execute the following once:

```shell
echo "autoload -U compinit; compinit" >> ~/.zshrc
```

To automatically load completions for all your shell session, execute once:

```shell
koyeb completion zsh > "${fpath[1]}/_koyeb"
```

You will need to start a new shell for this setup to take effect.

#### Fish

You can easily run `koyeb completion fish | source` to add completions to your current Fish session.

To automatically load completions for all your shell session, execute once:

```shell
koyeb completion fish > ~/.config/fish/completions/koyeb.fish
```

## Contribute

Checkout [CONTRIBUTING.md](CONTRIBUTING.md)

