---
title: "Koyeb CLI Reference"
shortTitle: Reference
description: "Discover all the commands available via the Koyeb CLI and how to use them to interact with the Koyeb serverless platform directly from the terminal."
---

The Koyeb CLI allows you to interact with Koyeb directly from the terminal. This documentation references all commands and options available in the CLI.

If you have not installed the Koyeb CLI yet, please read the [installation guide](/docs/cli/installation).
## koyeb

Koyeb CLI

### Options

```
  -a, --app string      App
  -c, --config string   config file (default is $HOME/.koyeb.yaml)
  -d, --debug           debug
  -h, --help            help for koyeb
  -o, --output string   output format (yaml,json,table)
      --token string    API token
      --url string      url of the api (default "https://app.koyeb.com")
```



* [koyeb apps](#koyeb-apps)	 - Apps
fault
* [koyeb login](#koyeb-login)	 - Login to your Koyeb account
* [koyeb secrets](#koyeb-secrets)	 - Secrets
* [koyeb services](#koyeb-services)	 - Services
* [koyeb version](#koyeb-version)	 - Get version

## koyeb apps

Apps

### Options

```
  -h, --help   help for apps
```

### Options inherited from parent commands

```
  -a, --app string      App
  -c, --config string   config file (default is $HOME/.koyeb.yaml)
  -d, --debug           debug
  -o, --output string   output format (yaml,json,table)
      --token string    API token
      --url string      url of the api (default "https://app.koyeb.com")
```



* [koyeb](#koyeb)	 - Koyeb CLI
* [koyeb apps create](#koyeb-apps-create)	 - Create apps
* [koyeb apps current](#koyeb-apps-current)	 - Get current app
* [koyeb apps delete](#koyeb-apps-delete)	 - Delete apps
* [koyeb apps describe](#koyeb-apps-describe)	 - Describe apps
* [koyeb apps get](#koyeb-apps-get)	 - Get app
* [koyeb apps init](#koyeb-apps-init)	 - Create app and service
* [koyeb apps list](#koyeb-apps-list)	 - List apps
* [koyeb apps switch](#koyeb-apps-switch)	 - Set current app
* [koyeb apps update](#koyeb-apps-update)	 - Update apps

## koyeb apps create

Create apps

```
koyeb apps create [name] [flags]
```

### Options

```
  -h, --help   help for create
```

### Options inherited from parent commands

```
  -a, --app string      App
  -c, --config string   config file (default is $HOME/.koyeb.yaml)
  -d, --debug           debug
  -o, --output string   output format (yaml,json,table)
      --token string    API token
      --url string      url of the api (default "https://app.koyeb.com")
```



* [koyeb apps](#koyeb-apps)	 - Apps

## koyeb apps current

Get current app

```
koyeb apps current [flags]
```

### Options

```
  -h, --help   help for current
```

### Options inherited from parent commands

```
  -a, --app string      App
  -c, --config string   config file (default is $HOME/.koyeb.yaml)
  -d, --debug           debug
  -o, --output string   output format (yaml,json,table)
      --token string    API token
      --url string      url of the api (default "https://app.koyeb.com")
```



* [koyeb apps](#koyeb-apps)	 - Apps

## koyeb apps delete

Delete apps

```
koyeb apps delete [name] [flags]
```

### Options

```
  -f, --force   Force delete app and services
  -h, --help    help for delete
```

### Options inherited from parent commands

```
  -a, --app string      App
  -c, --config string   config file (default is $HOME/.koyeb.yaml)
  -d, --debug           debug
  -o, --output string   output format (yaml,json,table)
      --token string    API token
      --url string      url of the api (default "https://app.koyeb.com")
```



* [koyeb apps](#koyeb-apps)	 - Apps

## koyeb apps describe

Describe apps

```
koyeb apps describe [name] [flags]
```

### Options

```
  -h, --help   help for describe
```

### Options inherited from parent commands

```
  -a, --app string      App
  -c, --config string   config file (default is $HOME/.koyeb.yaml)
  -d, --debug           debug
  -o, --output string   output format (yaml,json,table)
      --token string    API token
      --url string      url of the api (default "https://app.koyeb.com")
```



* [koyeb apps](#koyeb-apps)	 - Apps

## koyeb apps get

Get app

```
koyeb apps get [name] [flags]
```

### Options

```
  -h, --help   help for get
```

### Options inherited from parent commands

```
  -a, --app string      App
  -c, --config string   config file (default is $HOME/.koyeb.yaml)
  -d, --debug           debug
  -o, --output string   output format (yaml,json,table)
      --token string    API token
      --url string      url of the api (default "https://app.koyeb.com")
```



* [koyeb apps](#koyeb-apps)	 - Apps

## koyeb apps init

Create app and service

```
koyeb apps init [name] [flags]
```

### Options

```
      --docker string                           Docker image (default "koyeb/demo")
      --docker-args strings                     Docker args
      --docker-command string                   Docker command
      --docker-private-registry-secret string   Docker private registry secret
      --env strings                             Env
  -h, --help                                    help for init
      --instance-type string                    Instance type (default "nano")
      --max-scale int                           Max scale (default 1)
      --min-scale int                           Min scale (default 1)
      --ports strings                           Ports (default [80:http])
      --regions strings                         Regions (default [par])
      --routes strings                          Ports (default [/:80])
```

### Options inherited from parent commands

```
  -a, --app string      App
  -c, --config string   config file (default is $HOME/.koyeb.yaml)
  -d, --debug           debug
  -o, --output string   output format (yaml,json,table)
      --token string    API token
      --url string      url of the api (default "https://app.koyeb.com")
```



* [koyeb apps](#koyeb-apps)	 - Apps

## koyeb apps list

List apps

```
koyeb apps list [flags]
```

### Options

```
  -h, --help   help for list
```

### Options inherited from parent commands

```
  -a, --app string      App
  -c, --config string   config file (default is $HOME/.koyeb.yaml)
  -d, --debug           debug
  -o, --output string   output format (yaml,json,table)
      --token string    API token
      --url string      url of the api (default "https://app.koyeb.com")
```



* [koyeb apps](#koyeb-apps)	 - Apps

## koyeb apps switch

Set current app

```
koyeb apps switch [flags]
```

### Options

```
  -h, --help   help for switch
```

### Options inherited from parent commands

```
  -a, --app string      App
  -c, --config string   config file (default is $HOME/.koyeb.yaml)
  -d, --debug           debug
  -o, --output string   output format (yaml,json,table)
      --token string    API token
      --url string      url of the api (default "https://app.koyeb.com")
```



* [koyeb apps](#koyeb-apps)	 - Apps

## koyeb apps update

Update apps

```
koyeb apps update [name] [flags]
```

### Options

```
  -h, --help          help for update
  -n, --name string   Name of the app
```

### Options inherited from parent commands

```
  -a, --app string      App
  -c, --config string   config file (default is $HOME/.koyeb.yaml)
  -d, --debug           debug
  -o, --output string   output format (yaml,json,table)
      --token string    API token
      --url string      url of the api (default "https://app.koyeb.com")
```



* [koyeb apps](#koyeb-apps)	 - Apps

## koyeb login

Login to your Koyeb account

```
koyeb login [flags]
```

### Options

```
  -h, --help   help for login
```

### Options inherited from parent commands

```
  -a, --app string      App
  -c, --config string   config file (default is $HOME/.koyeb.yaml)
  -d, --debug           debug
  -o, --output string   output format (yaml,json,table)
      --token string    API token
      --url string      url of the api (default "https://app.koyeb.com")
```



* [koyeb](#koyeb)	 - Koyeb CLI

## koyeb secrets

Secrets

### Options

```
  -h, --help   help for secrets
```

### Options inherited from parent commands

```
  -a, --app string      App
  -c, --config string   config file (default is $HOME/.koyeb.yaml)
  -d, --debug           debug
  -o, --output string   output format (yaml,json,table)
      --token string    API token
      --url string      url of the api (default "https://app.koyeb.com")
```



* [koyeb](#koyeb)	 - Koyeb CLI
* [koyeb secrets create](#koyeb-secrets-create)	 - Create secrets
* [koyeb secrets delete](#koyeb-secrets-delete)	 - Delete secrets
* [koyeb secrets describe](#koyeb-secrets-describe)	 - Describe secrets
* [koyeb secrets get](#koyeb-secrets-get)	 - Get secret
* [koyeb secrets list](#koyeb-secrets-list)	 - List secrets
* [koyeb secrets update](#koyeb-secrets-update)	 - Update secrets

## koyeb secrets create

Create secrets

```
koyeb secrets create [name] [flags]
```

### Options

```
  -h, --help               help for create
  -v, --value string       Secret Value
      --value-from-stdin   Secret Value from stdin
```

### Options inherited from parent commands

```
  -a, --app string      App
  -c, --config string   config file (default is $HOME/.koyeb.yaml)
  -d, --debug           debug
  -o, --output string   output format (yaml,json,table)
      --token string    API token
      --url string      url of the api (default "https://app.koyeb.com")
```



* [koyeb secrets](#koyeb-secrets)	 - Secrets

## koyeb secrets delete

Delete secrets

```
koyeb secrets delete [name] [flags]
```

### Options

```
  -h, --help   help for delete
```

### Options inherited from parent commands

```
  -a, --app string      App
  -c, --config string   config file (default is $HOME/.koyeb.yaml)
  -d, --debug           debug
  -o, --output string   output format (yaml,json,table)
      --token string    API token
      --url string      url of the api (default "https://app.koyeb.com")
```



* [koyeb secrets](#koyeb-secrets)	 - Secrets

## koyeb secrets describe

Describe secrets

```
koyeb secrets describe [name] [flags]
```

### Options

```
  -h, --help   help for describe
```

### Options inherited from parent commands

```
  -a, --app string      App
  -c, --config string   config file (default is $HOME/.koyeb.yaml)
  -d, --debug           debug
  -o, --output string   output format (yaml,json,table)
      --token string    API token
      --url string      url of the api (default "https://app.koyeb.com")
```



* [koyeb secrets](#koyeb-secrets)	 - Secrets

## koyeb secrets get

Get secret

```
koyeb secrets get [name] [flags]
```

### Options

```
  -h, --help   help for get
```

### Options inherited from parent commands

```
  -a, --app string      App
  -c, --config string   config file (default is $HOME/.koyeb.yaml)
  -d, --debug           debug
  -o, --output string   output format (yaml,json,table)
      --token string    API token
      --url string      url of the api (default "https://app.koyeb.com")
```



* [koyeb secrets](#koyeb-secrets)	 - Secrets

## koyeb secrets list

List secrets

```
koyeb secrets list [flags]
```

### Options

```
  -h, --help   help for list
```

### Options inherited from parent commands

```
  -a, --app string      App
  -c, --config string   config file (default is $HOME/.koyeb.yaml)
  -d, --debug           debug
  -o, --output string   output format (yaml,json,table)
      --token string    API token
      --url string      url of the api (default "https://app.koyeb.com")
```



* [koyeb secrets](#koyeb-secrets)	 - Secrets

## koyeb secrets update

Update secrets

```
koyeb secrets update [name] [flags]
```

### Options

```
  -h, --help               help for update
  -v, --value string       Secret Value
      --value-from-stdin   Secret Value from stdin
```

### Options inherited from parent commands

```
  -a, --app string      App
  -c, --config string   config file (default is $HOME/.koyeb.yaml)
  -d, --debug           debug
  -o, --output string   output format (yaml,json,table)
      --token string    API token
      --url string      url of the api (default "https://app.koyeb.com")
```



* [koyeb secrets](#koyeb-secrets)	 - Secrets

## koyeb services

Services

### Options

```
  -h, --help   help for services
```

### Options inherited from parent commands

```
  -a, --app string      App
  -c, --config string   config file (default is $HOME/.koyeb.yaml)
  -d, --debug           debug
  -o, --output string   output format (yaml,json,table)
      --token string    API token
      --url string      url of the api (default "https://app.koyeb.com")
```



* [koyeb](#koyeb)	 - Koyeb CLI
* [koyeb services create](#koyeb-services-create)	 - Create services
* [koyeb services delete](#koyeb-services-delete)	 - Delete services
* [koyeb services describe](#koyeb-services-describe)	 - Describe services
* [koyeb services get](#koyeb-services-get)	 - Get service
* [koyeb services list](#koyeb-services-list)	 - List services
* [koyeb services logs](#koyeb-services-logs)	 - Get the service logs
* [koyeb services redeploy](#koyeb-services-redeploy)	 - Redeploy services
* [koyeb services update](#koyeb-services-update)	 - Update services

## koyeb services create

Create services

```
koyeb services create [name] [flags]
```

### Options

```
      --docker string                           Docker image (default "koyeb/demo")
      --docker-args strings                     Docker args
      --docker-command string                   Docker command
      --docker-private-registry-secret string   Docker private registry secret
      --env strings                             Env
  -h, --help                                    help for create
      --instance-type string                    Instance type (default "nano")
      --max-scale int                           Max scale (default 1)
      --min-scale int                           Min scale (default 1)
      --ports strings                           Ports (default [80:http])
      --regions strings                         Regions (default [par])
      --routes strings                          Ports (default [/:80])
```

### Options inherited from parent commands

```
  -a, --app string      App
  -c, --config string   config file (default is $HOME/.koyeb.yaml)
  -d, --debug           debug
  -o, --output string   output format (yaml,json,table)
      --token string    API token
      --url string      url of the api (default "https://app.koyeb.com")
```



* [koyeb services](#koyeb-services)	 - Services

## koyeb services delete

Delete services

```
koyeb services delete [name] [flags]
```

### Options

```
  -h, --help   help for delete
```

### Options inherited from parent commands

```
  -a, --app string      App
  -c, --config string   config file (default is $HOME/.koyeb.yaml)
  -d, --debug           debug
  -o, --output string   output format (yaml,json,table)
      --token string    API token
      --url string      url of the api (default "https://app.koyeb.com")
```



* [koyeb services](#koyeb-services)	 - Services

## koyeb services describe

Describe services

```
koyeb services describe [name] [flags]
```

### Options

```
  -h, --help   help for describe
```

### Options inherited from parent commands

```
  -a, --app string      App
  -c, --config string   config file (default is $HOME/.koyeb.yaml)
  -d, --debug           debug
  -o, --output string   output format (yaml,json,table)
      --token string    API token
      --url string      url of the api (default "https://app.koyeb.com")
```



* [koyeb services](#koyeb-services)	 - Services

## koyeb services get

Get service

```
koyeb services get [name] [flags]
```

### Options

```
  -h, --help   help for get
```

### Options inherited from parent commands

```
  -a, --app string      App
  -c, --config string   config file (default is $HOME/.koyeb.yaml)
  -d, --debug           debug
  -o, --output string   output format (yaml,json,table)
      --token string    API token
      --url string      url of the api (default "https://app.koyeb.com")
```



* [koyeb services](#koyeb-services)	 - Services

## koyeb services list

List services

```
koyeb services list [flags]
```

### Options

```
  -h, --help   help for list
```

### Options inherited from parent commands

```
  -a, --app string      App
  -c, --config string   config file (default is $HOME/.koyeb.yaml)
  -d, --debug           debug
  -o, --output string   output format (yaml,json,table)
      --token string    API token
      --url string      url of the api (default "https://app.koyeb.com")
```



* [koyeb services](#koyeb-services)	 - Services

## koyeb services logs

Get the service logs

```
koyeb services logs [name] [flags]
```

### Options

```
  -h, --help              help for logs
      --instance string   Instance
      --stderr            Get stderr stream
```

### Options inherited from parent commands

```
  -a, --app string      App
  -c, --config string   config file (default is $HOME/.koyeb.yaml)
  -d, --debug           debug
  -o, --output string   output format (yaml,json,table)
      --token string    API token
      --url string      url of the api (default "https://app.koyeb.com")
```



* [koyeb services](#koyeb-services)	 - Services

## koyeb services redeploy

Redeploy services

```
koyeb services redeploy [name] [flags]
```

### Options

```
  -h, --help   help for redeploy
```

### Options inherited from parent commands

```
  -a, --app string      App
  -c, --config string   config file (default is $HOME/.koyeb.yaml)
  -d, --debug           debug
  -o, --output string   output format (yaml,json,table)
      --token string    API token
      --url string      url of the api (default "https://app.koyeb.com")
```



* [koyeb services](#koyeb-services)	 - Services

## koyeb services update

Update services

```
koyeb services update [name] [flags]
```

### Options

```
      --docker string                           Docker image (default "koyeb/demo")
      --docker-args strings                     Docker args
      --docker-command string                   Docker command
      --docker-private-registry-secret string   Docker private registry secret
      --env strings                             Env
  -h, --help                                    help for update
      --instance-type string                    Instance type (default "nano")
      --max-scale int                           Max scale (default 1)
      --min-scale int                           Min scale (default 1)
      --ports strings                           Ports (default [80:http])
      --regions strings                         Regions (default [par])
      --routes strings                          Ports (default [/:80])
```

### Options inherited from parent commands

```
  -a, --app string      App
  -c, --config string   config file (default is $HOME/.koyeb.yaml)
  -d, --debug           debug
  -o, --output string   output format (yaml,json,table)
      --token string    API token
      --url string      url of the api (default "https://app.koyeb.com")
```



* [koyeb services](#koyeb-services)	 - Services

## koyeb version

Get version

```
koyeb version [flags]
```

### Options

```
  -h, --help   help for version
```

### Options inherited from parent commands

```
  -a, --app string      App
  -c, --config string   config file (default is $HOME/.koyeb.yaml)
  -d, --debug           debug
  -o, --output string   output format (yaml,json,table)
      --token string    API token
      --url string      url of the api (default "https://app.koyeb.com")
```



* [koyeb](#koyeb)	 - Koyeb CLI

