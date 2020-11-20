![Build](https://github.com/koyeb/koyeb-cli/workflows/Release/badge.svg)

# Koyeb CLI

The Koyeb CLI (Command Line Interface) is a powerful tool to manage your Koyeb serverless infrastructure directly from your terminal.

# Installation

## Release from GitHub

The CLI can be installed from pre-compiled binaries for macOS (darwin), Linux and Windows.

### MacOS

```shell
# Fetch latest release URL for macOS from GitHub
LATEST_RELEASE_URL=$(curl -s https://api.github.com/repos/koyeb/koyeb-cli/releases/latest | grep 'http.*koyeb-cli-darwin-x86_64' | awk '{print $2}' | sed 's|[\"\,]*||g')
# Install latest version as /usr/local/bin/koyeb
curl -sL $LATEST_RELEASE_URL -o /usr/local/bin/koyeb
chmod +x /usr/local/bin/koyeb
```

### Linux

```shell
# Fetch latest release URL for Linux from GitHub
LATEST_RELEASE_URL=$(curl -s https://api.github.com/repos/koyeb/koyeb-cli/releases/latest | grep 'http.*koyeb-cli-linux-x86_64' | awk '{print $2}' | sed 's|[\"\,]*||g')
# Install latest version as /usr/local/bin/koyeb
curl -sL $LATEST_RELEASE_URL -o /usr/local/bin/koyeb
chmod +x /usr/local/bin/koyeb
```

### Windows

Simply download the latest release: https://github.com/koyeb/koyeb-cli/releases.

## Living at the Edge

To install the latest `koyeb` binary with go, simply run:

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

### General usage

```shell
➜ koyeb --help
Koyeb cli

Usage:
  koyeb [command]

Available Commands:
  completion  Generate completion script
  create      Create a resource from a file
  delete      Delete resources by name and id
  describe    Display one resources
  get         Display one or many resources
  help        Help about any command
  init        Init configuration
  invoke      Invoke a function
  logs        Get the log of one resources
  update      Update one resources
  version     Get version

Flags:
  -c, --config string   config file (default is $HOME/.koyeb.yaml)
  -d, --debug           debug
  -h, --help            help for koyeb
  -o, --output string   output format (yaml,json,table)
      --token string    API token
      --url string      url of the api (default "https://app.koyeb.com")

Use "koyeb [command] --help" for more information about a command.
```



# Enabling shell auto-completion

`koyeb` has auto-completion support for `bash` and `zsh`. 

## Bash

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

## Zsh

If shell completion is not already enabled in your environment you will need to enable it.  You can execute the following once:

```shell
echo "autoload -U compinit; compinit" >> ~/.zshrc
```

To automatically load completions for all your shell session, execute once:

```shell
koyeb completion zsh > "${fpath[1]}/_koyeb"
```

You will need to start a new shell for this setup to take effect.



# Contribute

Checkout [CONTRIBUTING.md](CONTRIBUTING.md)

