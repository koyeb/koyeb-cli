---
title: "Koyeb CLI Reference"
shortTitle: Reference
description: "Discover all the commands available via the Koyeb CLI and how to use them to interact with the Koyeb serverless platform directly from the terminal."
---

# Koyeb CLI Reference"

The Koyeb CLI allows you to interact with Koyeb directly from the terminal. This documentation references all commands and options available in the CLI.

If you have not installed the Koyeb CLI yet, please read the [installation guide](/docs/cli/installation).
## koyeb

Koyeb CLI

### Options

```
  -c, --config string   config file (default is $HOME/.koyeb.yaml)
  -d, --debug           debug
      --full            show full id
  -h, --help            help for koyeb
  -o, --output output   output format (yaml,json,table)
      --token string    API token
      --url string      url of the api (default "https://app.koyeb.com")
```



* [koyeb apps](#koyeb-apps)	 - Apps

* [koyeb deployments](#koyeb-deployments)	 - Deployments
* [koyeb domains](#koyeb-domains)	 - Domains
* [koyeb instances](#koyeb-instances)	 - Instances
* [koyeb login](#koyeb-login)	 - Login to your Koyeb account
* [koyeb secrets](#koyeb-secrets)	 - Secrets
* [koyeb services](#koyeb-services)	 - Services
* [koyeb version](#koyeb-version)	 - Get version

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
  -c, --config string   config file (default is $HOME/.koyeb.yaml)
  -d, --debug           debug
      --full            show full id
  -o, --output output   output format (yaml,json,table)
      --token string    API token
      --url string      url of the api (default "https://app.koyeb.com")
```



* [koyeb](#koyeb)	 - Koyeb CLI

## koyeb apps

Apps

### Options

```
  -h, --help   help for apps
```

### Options inherited from parent commands

```
  -c, --config string   config file (default is $HOME/.koyeb.yaml)
  -d, --debug           debug
      --full            show full id
  -o, --output output   output format (yaml,json,table)
      --token string    API token
      --url string      url of the api (default "https://app.koyeb.com")
```



* [koyeb](#koyeb)	 - Koyeb CLI
* [koyeb apps create](#koyeb-apps-create)	 - Create app
* [koyeb apps delete](#koyeb-apps-delete)	 - Delete app
* [koyeb apps describe](#koyeb-apps-describe)	 - Describe app
* [koyeb apps get](#koyeb-apps-get)	 - Get app
* [koyeb apps init](#koyeb-apps-init)	 - Create app and service
* [koyeb apps list](#koyeb-apps-list)	 - List apps
* [koyeb apps pause](#koyeb-apps-pause)	 - Pause app
* [koyeb apps resume](#koyeb-apps-resume)	 - Resume app
* [koyeb apps update](#koyeb-apps-update)	 - Update app

## koyeb apps create

Create app

```
koyeb apps create NAME [flags]
```

### Options

```
  -h, --help   help for create
```

### Options inherited from parent commands

```
  -c, --config string   config file (default is $HOME/.koyeb.yaml)
  -d, --debug           debug
      --full            show full id
  -o, --output output   output format (yaml,json,table)
      --token string    API token
      --url string      url of the api (default "https://app.koyeb.com")
```



* [koyeb apps](#koyeb-apps)	 - Apps

## koyeb apps delete

Delete app

```
koyeb apps delete NAME [flags]
```

### Options

```
  -h, --help   help for delete
```

### Options inherited from parent commands

```
  -c, --config string   config file (default is $HOME/.koyeb.yaml)
  -d, --debug           debug
      --full            show full id
  -o, --output output   output format (yaml,json,table)
      --token string    API token
      --url string      url of the api (default "https://app.koyeb.com")
```



* [koyeb apps](#koyeb-apps)	 - Apps

## koyeb apps describe

Describe app

```
koyeb apps describe NAME [flags]
```

### Options

```
  -h, --help   help for describe
```

### Options inherited from parent commands

```
  -c, --config string   config file (default is $HOME/.koyeb.yaml)
  -d, --debug           debug
      --full            show full id
  -o, --output output   output format (yaml,json,table)
      --token string    API token
      --url string      url of the api (default "https://app.koyeb.com")
```



* [koyeb apps](#koyeb-apps)	 - Apps

## koyeb apps get

Get app

```
koyeb apps get NAME [flags]
```

### Options

```
  -h, --help   help for get
```

### Options inherited from parent commands

```
  -c, --config string   config file (default is $HOME/.koyeb.yaml)
  -d, --debug           debug
      --full            show full id
  -o, --output output   output format (yaml,json,table)
      --token string    API token
      --url string      url of the api (default "https://app.koyeb.com")
```



* [koyeb apps](#koyeb-apps)	 - Apps

## koyeb apps init

Create app and service

```
koyeb apps init NAME [flags]
```

### Options

```
      --checks strings                          HTTP healthcheck (<port>:http:<path>) and TCP healthcheck (<port>:tcp) - Available for "WEB" service only
      --docker string                           Docker image
      --docker-args strings                     Docker args
      --docker-command string                   Docker command
      --docker-entrypoint strings               Docker entrypoint
      --docker-private-registry-secret string   Docker private registry secret
      --env strings                             Env
      --git string                              Git repository
      --git-branch string                       Git branch
      --git-build-command string                Buid command (legacy, prefer git-buildpack-build-command)
      --git-builder string                      Builder to use, either "buildpack" (default) or "docker" (default "buildpack")
      --git-buildpack-build-command string      Buid command
      --git-buildpack-run-command string        Run command
      --git-docker-args strings                 Arguments for the Docker CMD
      --git-docker-command string               Docker CMD
      --git-docker-dockerfile string            Dockerfile path
      --git-docker-entrypoint strings           Docker entrypoint
      --git-docker-target string                Docker target
      --git-no-deploy-on-push                   Disable new deployments creation when code changes are pushed on the configured branch
      --git-run-command string                  Run command (legacy, prefer git-buildpack-run-command)
      --git-workdir string                      Path to the sub-directory containing the code to build and deploy
  -h, --help                                    help for init
      --instance-type string                    Instance type (default "nano")
      --max-scale int                           Max scale (default 1)
      --min-scale int                           Min scale (default 1)
      --ports strings                           Ports - Available for "WEB" service only (default [80:http])
      --regions strings                         Regions (default [fra])
      --routes strings                          Routes - Available for "WEB" service only (default [/:80])
      --type string                             Service type, either "WEB" or "WORKER" (default "WEB")
```

### Options inherited from parent commands

```
  -c, --config string   config file (default is $HOME/.koyeb.yaml)
  -d, --debug           debug
      --full            show full id
  -o, --output output   output format (yaml,json,table)
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
  -c, --config string   config file (default is $HOME/.koyeb.yaml)
  -d, --debug           debug
      --full            show full id
  -o, --output output   output format (yaml,json,table)
      --token string    API token
      --url string      url of the api (default "https://app.koyeb.com")
```



* [koyeb apps](#koyeb-apps)	 - Apps

## koyeb apps pause

Pause app

```
koyeb apps pause NAME [flags]
```

### Options

```
  -h, --help   help for pause
```

### Options inherited from parent commands

```
  -c, --config string   config file (default is $HOME/.koyeb.yaml)
  -d, --debug           debug
      --full            show full id
  -o, --output output   output format (yaml,json,table)
      --token string    API token
      --url string      url of the api (default "https://app.koyeb.com")
```



* [koyeb apps](#koyeb-apps)	 - Apps

## koyeb apps resume

Resume app

```
koyeb apps resume NAME [flags]
```

### Options

```
  -h, --help   help for resume
```

### Options inherited from parent commands

```
  -c, --config string   config file (default is $HOME/.koyeb.yaml)
  -d, --debug           debug
      --full            show full id
  -o, --output output   output format (yaml,json,table)
      --token string    API token
      --url string      url of the api (default "https://app.koyeb.com")
```



* [koyeb apps](#koyeb-apps)	 - Apps

## koyeb apps update

Update app

```
koyeb apps update NAME [flags]
```

### Options

```
  -h, --help          help for update
  -n, --name string   Name of the app
```

### Options inherited from parent commands

```
  -c, --config string   config file (default is $HOME/.koyeb.yaml)
  -d, --debug           debug
      --full            show full id
  -o, --output output   output format (yaml,json,table)
      --token string    API token
      --url string      url of the api (default "https://app.koyeb.com")
```



* [koyeb apps](#koyeb-apps)	 - Apps

## koyeb domains

Domains

### Options

```
  -h, --help   help for domains
```

### Options inherited from parent commands

```
  -c, --config string   config file (default is $HOME/.koyeb.yaml)
  -d, --debug           debug
      --full            show full id
  -o, --output output   output format (yaml,json,table)
      --token string    API token
      --url string      url of the api (default "https://app.koyeb.com")
```



* [koyeb](#koyeb)	 - Koyeb CLI
* [koyeb domains attach](#koyeb-domains-attach)	 - Attach a custom domain to an existing app
* [koyeb domains create](#koyeb-domains-create)	 - Create domain
* [koyeb domains delete](#koyeb-domains-delete)	 - Delete domain
* [koyeb domains describe](#koyeb-domains-describe)	 - Describe domain
* [koyeb domains detach](#koyeb-domains-detach)	 - Detach a custom domain from the app it is currently attached to
* [koyeb domains get](#koyeb-domains-get)	 - Get domain
* [koyeb domains list](#koyeb-domains-list)	 - List domains
* [koyeb domains refresh](#koyeb-domains-refresh)	 - Refresh a custom domain verification status

## koyeb domains attach

Attach a custom domain to an existing app

```
koyeb domains attach NAME APP [flags]
```

### Options

```
  -h, --help   help for attach
```

### Options inherited from parent commands

```
  -c, --config string   config file (default is $HOME/.koyeb.yaml)
  -d, --debug           debug
      --full            show full id
  -o, --output output   output format (yaml,json,table)
      --token string    API token
      --url string      url of the api (default "https://app.koyeb.com")
```



* [koyeb domains](#koyeb-domains)	 - Domains

## koyeb domains create

Create domain

```
koyeb domains create NAME [flags]
```

### Options

```
      --attach-to string   Upon creation, assign to given app
  -h, --help               help for create
```

### Options inherited from parent commands

```
  -c, --config string   config file (default is $HOME/.koyeb.yaml)
  -d, --debug           debug
      --full            show full id
  -o, --output output   output format (yaml,json,table)
      --token string    API token
      --url string      url of the api (default "https://app.koyeb.com")
```



* [koyeb domains](#koyeb-domains)	 - Domains

## koyeb domains delete

Delete domain

```
koyeb domains delete [flags]
```

### Options

```
  -h, --help   help for delete
```

### Options inherited from parent commands

```
  -c, --config string   config file (default is $HOME/.koyeb.yaml)
  -d, --debug           debug
      --full            show full id
  -o, --output output   output format (yaml,json,table)
      --token string    API token
      --url string      url of the api (default "https://app.koyeb.com")
```



* [koyeb domains](#koyeb-domains)	 - Domains

## koyeb domains describe

Describe domain

```
koyeb domains describe [flags]
```

### Options

```
  -h, --help   help for describe
```

### Options inherited from parent commands

```
  -c, --config string   config file (default is $HOME/.koyeb.yaml)
  -d, --debug           debug
      --full            show full id
  -o, --output output   output format (yaml,json,table)
      --token string    API token
      --url string      url of the api (default "https://app.koyeb.com")
```



* [koyeb domains](#koyeb-domains)	 - Domains

## koyeb domains detach

Detach a custom domain from the app it is currently attached to

```
koyeb domains detach NAME [flags]
```

### Options

```
  -h, --help   help for detach
```

### Options inherited from parent commands

```
  -c, --config string   config file (default is $HOME/.koyeb.yaml)
  -d, --debug           debug
      --full            show full id
  -o, --output output   output format (yaml,json,table)
      --token string    API token
      --url string      url of the api (default "https://app.koyeb.com")
```



* [koyeb domains](#koyeb-domains)	 - Domains

## koyeb domains get

Get domain

```
koyeb domains get NAME [flags]
```

### Options

```
  -h, --help   help for get
```

### Options inherited from parent commands

```
  -c, --config string   config file (default is $HOME/.koyeb.yaml)
  -d, --debug           debug
      --full            show full id
  -o, --output output   output format (yaml,json,table)
      --token string    API token
      --url string      url of the api (default "https://app.koyeb.com")
```



* [koyeb domains](#koyeb-domains)	 - Domains

## koyeb domains list

List domains

```
koyeb domains list [flags]
```

### Options

```
  -h, --help   help for list
```

### Options inherited from parent commands

```
  -c, --config string   config file (default is $HOME/.koyeb.yaml)
  -d, --debug           debug
      --full            show full id
  -o, --output output   output format (yaml,json,table)
      --token string    API token
      --url string      url of the api (default "https://app.koyeb.com")
```



* [koyeb domains](#koyeb-domains)	 - Domains

## koyeb domains refresh

Refresh a custom domain verification status

```
koyeb domains refresh NAME [flags]
```

### Options

```
  -h, --help   help for refresh
```

### Options inherited from parent commands

```
  -c, --config string   config file (default is $HOME/.koyeb.yaml)
  -d, --debug           debug
      --full            show full id
  -o, --output output   output format (yaml,json,table)
      --token string    API token
      --url string      url of the api (default "https://app.koyeb.com")
```



* [koyeb domains](#koyeb-domains)	 - Domains

## koyeb secrets

Secrets

### Options

```
  -h, --help   help for secrets
```

### Options inherited from parent commands

```
  -c, --config string   config file (default is $HOME/.koyeb.yaml)
  -d, --debug           debug
      --full            show full id
  -o, --output output   output format (yaml,json,table)
      --token string    API token
      --url string      url of the api (default "https://app.koyeb.com")
```



* [koyeb](#koyeb)	 - Koyeb CLI
* [koyeb secrets create](#koyeb-secrets-create)	 - Create secret
* [koyeb secrets delete](#koyeb-secrets-delete)	 - Delete secret
* [koyeb secrets describe](#koyeb-secrets-describe)	 - Describe secret
* [koyeb secrets get](#koyeb-secrets-get)	 - Get secret
* [koyeb secrets list](#koyeb-secrets-list)	 - List secrets
* [koyeb secrets update](#koyeb-secrets-update)	 - Update secret

## koyeb secrets create

Create secret

```
koyeb secrets create NAME [flags]
```

### Options

```
  -h, --help               help for create
  -v, --value string       Secret Value
      --value-from-stdin   Secret Value from stdin
```

### Options inherited from parent commands

```
  -c, --config string   config file (default is $HOME/.koyeb.yaml)
  -d, --debug           debug
      --full            show full id
  -o, --output output   output format (yaml,json,table)
      --token string    API token
      --url string      url of the api (default "https://app.koyeb.com")
```



* [koyeb secrets](#koyeb-secrets)	 - Secrets

## koyeb secrets delete

Delete secret

```
koyeb secrets delete NAME [flags]
```

### Options

```
  -h, --help   help for delete
```

### Options inherited from parent commands

```
  -c, --config string   config file (default is $HOME/.koyeb.yaml)
  -d, --debug           debug
      --full            show full id
  -o, --output output   output format (yaml,json,table)
      --token string    API token
      --url string      url of the api (default "https://app.koyeb.com")
```



* [koyeb secrets](#koyeb-secrets)	 - Secrets

## koyeb secrets describe

Describe secret

```
koyeb secrets describe NAME [flags]
```

### Options

```
  -h, --help   help for describe
```

### Options inherited from parent commands

```
  -c, --config string   config file (default is $HOME/.koyeb.yaml)
  -d, --debug           debug
      --full            show full id
  -o, --output output   output format (yaml,json,table)
      --token string    API token
      --url string      url of the api (default "https://app.koyeb.com")
```



* [koyeb secrets](#koyeb-secrets)	 - Secrets

## koyeb secrets get

Get secret

```
koyeb secrets get NAME [flags]
```

### Options

```
  -h, --help   help for get
```

### Options inherited from parent commands

```
  -c, --config string   config file (default is $HOME/.koyeb.yaml)
  -d, --debug           debug
      --full            show full id
  -o, --output output   output format (yaml,json,table)
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
  -c, --config string   config file (default is $HOME/.koyeb.yaml)
  -d, --debug           debug
      --full            show full id
  -o, --output output   output format (yaml,json,table)
      --token string    API token
      --url string      url of the api (default "https://app.koyeb.com")
```



* [koyeb secrets](#koyeb-secrets)	 - Secrets

## koyeb secrets update

Update secret

```
koyeb secrets update NAME [flags]
```

### Options

```
  -h, --help               help for update
  -v, --value string       Secret Value
      --value-from-stdin   Secret Value from stdin
```

### Options inherited from parent commands

```
  -c, --config string   config file (default is $HOME/.koyeb.yaml)
  -d, --debug           debug
      --full            show full id
  -o, --output output   output format (yaml,json,table)
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
  -c, --config string   config file (default is $HOME/.koyeb.yaml)
  -d, --debug           debug
      --full            show full id
  -o, --output output   output format (yaml,json,table)
      --token string    API token
      --url string      url of the api (default "https://app.koyeb.com")
```



* [koyeb](#koyeb)	 - Koyeb CLI
* [koyeb services create](#koyeb-services-create)	 - Create service
* [koyeb services delete](#koyeb-services-delete)	 - Delete service
* [koyeb services describe](#koyeb-services-describe)	 - Describe service
* [koyeb services exec](#koyeb-services-exec)	 - Run a command in the context of an instance selected among the service instances
* [koyeb services get](#koyeb-services-get)	 - Get service
* [koyeb services list](#koyeb-services-list)	 - List services
* [koyeb services logs](#koyeb-services-logs)	 - Get the service logs
* [koyeb services pause](#koyeb-services-pause)	 - Pause service
* [koyeb services redeploy](#koyeb-services-redeploy)	 - Redeploy service
* [koyeb services resume](#koyeb-services-resume)	 - Resume service
* [koyeb services update](#koyeb-services-update)	 - Update service

## koyeb services create

Create service

```
koyeb services create NAME [flags]
```

### Options

```
  -a, --app string                              App
      --checks strings                          HTTP healthcheck (<port>:http:<path>) and TCP healthcheck (<port>:tcp) - Available for "WEB" service only
      --docker string                           Docker image
      --docker-args strings                     Docker args
      --docker-command string                   Docker command
      --docker-entrypoint strings               Docker entrypoint
      --docker-private-registry-secret string   Docker private registry secret
      --env strings                             Env
      --git string                              Git repository
      --git-branch string                       Git branch
      --git-build-command string                Buid command (legacy, prefer git-buildpack-build-command)
      --git-builder string                      Builder to use, either "buildpack" (default) or "docker" (default "buildpack")
      --git-buildpack-build-command string      Buid command
      --git-buildpack-run-command string        Run command
      --git-docker-args strings                 Arguments for the Docker CMD
      --git-docker-command string               Docker CMD
      --git-docker-dockerfile string            Dockerfile path
      --git-docker-entrypoint strings           Docker entrypoint
      --git-docker-target string                Docker target
      --git-no-deploy-on-push                   Disable new deployments creation when code changes are pushed on the configured branch
      --git-run-command string                  Run command (legacy, prefer git-buildpack-run-command)
      --git-workdir string                      Path to the sub-directory containing the code to build and deploy
  -h, --help                                    help for create
      --instance-type string                    Instance type (default "nano")
      --max-scale int                           Max scale (default 1)
      --min-scale int                           Min scale (default 1)
      --ports strings                           Ports - Available for "WEB" service only (default [80:http])
      --regions strings                         Regions (default [fra])
      --routes strings                          Routes - Available for "WEB" service only (default [/:80])
      --type string                             Service type, either "WEB" or "WORKER" (default "WEB")
```

### Options inherited from parent commands

```
  -c, --config string   config file (default is $HOME/.koyeb.yaml)
  -d, --debug           debug
      --full            show full id
  -o, --output output   output format (yaml,json,table)
      --token string    API token
      --url string      url of the api (default "https://app.koyeb.com")
```



* [koyeb services](#koyeb-services)	 - Services

## koyeb services delete

Delete service

```
koyeb services delete NAME [flags]
```

### Options

```
  -h, --help   help for delete
```

### Options inherited from parent commands

```
  -c, --config string   config file (default is $HOME/.koyeb.yaml)
  -d, --debug           debug
      --full            show full id
  -o, --output output   output format (yaml,json,table)
      --token string    API token
      --url string      url of the api (default "https://app.koyeb.com")
```



* [koyeb services](#koyeb-services)	 - Services

## koyeb services describe

Describe service

```
koyeb services describe NAME [flags]
```

### Options

```
  -h, --help   help for describe
```

### Options inherited from parent commands

```
  -c, --config string   config file (default is $HOME/.koyeb.yaml)
  -d, --debug           debug
      --full            show full id
  -o, --output output   output format (yaml,json,table)
      --token string    API token
      --url string      url of the api (default "https://app.koyeb.com")
```



* [koyeb services](#koyeb-services)	 - Services

## koyeb services exec

Run a command in the context of an instance selected among the service instances

```
koyeb services exec NAME CMD -- [args...] [flags]
```

### Options

```
  -h, --help   help for exec
```

### Options inherited from parent commands

```
  -c, --config string   config file (default is $HOME/.koyeb.yaml)
  -d, --debug           debug
      --full            show full id
  -o, --output output   output format (yaml,json,table)
      --token string    API token
      --url string      url of the api (default "https://app.koyeb.com")
```



* [koyeb services](#koyeb-services)	 - Services

## koyeb services get

Get service

```
koyeb services get NAME [flags]
```

### Options

```
  -h, --help   help for get
```

### Options inherited from parent commands

```
  -c, --config string   config file (default is $HOME/.koyeb.yaml)
  -d, --debug           debug
      --full            show full id
  -o, --output output   output format (yaml,json,table)
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
  -a, --app string    App
  -h, --help          help for list
  -n, --name string   Service name
```

### Options inherited from parent commands

```
  -c, --config string   config file (default is $HOME/.koyeb.yaml)
  -d, --debug           debug
      --full            show full id
  -o, --output output   output format (yaml,json,table)
      --token string    API token
      --url string      url of the api (default "https://app.koyeb.com")
```



* [koyeb services](#koyeb-services)	 - Services

## koyeb services logs

Get the service logs

```
koyeb services logs NAME [flags]
```

### Options

```
  -h, --help              help for logs
      --instance string   Instance
  -t, --type string       Type (runtime,build)
```

### Options inherited from parent commands

```
  -c, --config string   config file (default is $HOME/.koyeb.yaml)
  -d, --debug           debug
      --full            show full id
  -o, --output output   output format (yaml,json,table)
      --token string    API token
      --url string      url of the api (default "https://app.koyeb.com")
```



* [koyeb services](#koyeb-services)	 - Services

## koyeb services pause

Pause service

```
koyeb services pause NAME [flags]
```

### Options

```
  -h, --help   help for pause
```

### Options inherited from parent commands

```
  -c, --config string   config file (default is $HOME/.koyeb.yaml)
  -d, --debug           debug
      --full            show full id
  -o, --output output   output format (yaml,json,table)
      --token string    API token
      --url string      url of the api (default "https://app.koyeb.com")
```



* [koyeb services](#koyeb-services)	 - Services

## koyeb services redeploy

Redeploy service

```
koyeb services redeploy NAME [flags]
```

### Options

```
  -h, --help        help for redeploy
      --use-cache   Use cache to redeploy
```

### Options inherited from parent commands

```
  -c, --config string   config file (default is $HOME/.koyeb.yaml)
  -d, --debug           debug
      --full            show full id
  -o, --output output   output format (yaml,json,table)
      --token string    API token
      --url string      url of the api (default "https://app.koyeb.com")
```



* [koyeb services](#koyeb-services)	 - Services

## koyeb services resume

Resume service

```
koyeb services resume NAME [flags]
```

### Options

```
  -h, --help   help for resume
```

### Options inherited from parent commands

```
  -c, --config string   config file (default is $HOME/.koyeb.yaml)
  -d, --debug           debug
      --full            show full id
  -o, --output output   output format (yaml,json,table)
      --token string    API token
      --url string      url of the api (default "https://app.koyeb.com")
```



* [koyeb services](#koyeb-services)	 - Services

## koyeb services update

Update service

```
koyeb services update NAME [flags]
```

### Options

```
      --checks strings                          HTTP healthcheck (<port>:http:<path>) and TCP healthcheck (<port>:tcp) - Available for "WEB" service only
      --docker string                           Docker image
      --docker-args strings                     Docker args
      --docker-command string                   Docker command
      --docker-entrypoint strings               Docker entrypoint
      --docker-private-registry-secret string   Docker private registry secret
      --env strings                             Env
      --git string                              Git repository
      --git-branch string                       Git branch
      --git-build-command string                Buid command (legacy, prefer git-buildpack-build-command)
      --git-builder string                      Builder to use, either "buildpack" (default) or "docker" (default "buildpack")
      --git-buildpack-build-command string      Buid command
      --git-buildpack-run-command string        Run command
      --git-docker-args strings                 Arguments for the Docker CMD
      --git-docker-command string               Docker CMD
      --git-docker-dockerfile string            Dockerfile path
      --git-docker-entrypoint strings           Docker entrypoint
      --git-docker-target string                Docker target
      --git-no-deploy-on-push                   Disable new deployments creation when code changes are pushed on the configured branch
      --git-run-command string                  Run command (legacy, prefer git-buildpack-run-command)
      --git-workdir string                      Path to the sub-directory containing the code to build and deploy
  -h, --help                                    help for update
      --instance-type string                    Instance type (default "nano")
      --max-scale int                           Max scale (default 1)
      --min-scale int                           Min scale (default 1)
      --ports strings                           Ports - Available for "WEB" service only (default [80:http])
      --regions strings                         Regions (default [fra])
      --routes strings                          Routes - Available for "WEB" service only (default [/:80])
      --type string                             Service type, either "WEB" or "WORKER" (default "WEB")
```

### Options inherited from parent commands

```
  -c, --config string   config file (default is $HOME/.koyeb.yaml)
  -d, --debug           debug
      --full            show full id
  -o, --output output   output format (yaml,json,table)
      --token string    API token
      --url string      url of the api (default "https://app.koyeb.com")
```



* [koyeb services](#koyeb-services)	 - Services

## koyeb deployments

Deployments

### Options

```
  -h, --help   help for deployments
```

### Options inherited from parent commands

```
  -c, --config string   config file (default is $HOME/.koyeb.yaml)
  -d, --debug           debug
      --full            show full id
  -o, --output output   output format (yaml,json,table)
      --token string    API token
      --url string      url of the api (default "https://app.koyeb.com")
```



* [koyeb](#koyeb)	 - Koyeb CLI
* [koyeb deployments cancel](#koyeb-deployments-cancel)	 - Cancel deployment
* [koyeb deployments describe](#koyeb-deployments-describe)	 - Describe deployment
* [koyeb deployments get](#koyeb-deployments-get)	 - Get deployment
* [koyeb deployments list](#koyeb-deployments-list)	 - List deployments
* [koyeb deployments logs](#koyeb-deployments-logs)	 - Get deployment logs

## koyeb deployments cancel

Cancel deployment

```
koyeb deployments cancel NAME [flags]
```

### Options

```
  -h, --help   help for cancel
```

### Options inherited from parent commands

```
  -c, --config string   config file (default is $HOME/.koyeb.yaml)
  -d, --debug           debug
      --full            show full id
  -o, --output output   output format (yaml,json,table)
      --token string    API token
      --url string      url of the api (default "https://app.koyeb.com")
```



* [koyeb deployments](#koyeb-deployments)	 - Deployments

## koyeb deployments describe

Describe deployment

```
koyeb deployments describe NAME [flags]
```

### Options

```
  -h, --help   help for describe
```

### Options inherited from parent commands

```
  -c, --config string   config file (default is $HOME/.koyeb.yaml)
  -d, --debug           debug
      --full            show full id
  -o, --output output   output format (yaml,json,table)
      --token string    API token
      --url string      url of the api (default "https://app.koyeb.com")
```



* [koyeb deployments](#koyeb-deployments)	 - Deployments

## koyeb deployments get

Get deployment

```
koyeb deployments get NAME [flags]
```

### Options

```
  -h, --help   help for get
```

### Options inherited from parent commands

```
  -c, --config string   config file (default is $HOME/.koyeb.yaml)
  -d, --debug           debug
      --full            show full id
  -o, --output output   output format (yaml,json,table)
      --token string    API token
      --url string      url of the api (default "https://app.koyeb.com")
```



* [koyeb deployments](#koyeb-deployments)	 - Deployments

## koyeb deployments list

List deployments

```
koyeb deployments list [flags]
```

### Options

```
  -h, --help   help for list
```

### Options inherited from parent commands

```
  -c, --config string   config file (default is $HOME/.koyeb.yaml)
  -d, --debug           debug
      --full            show full id
  -o, --output output   output format (yaml,json,table)
      --token string    API token
      --url string      url of the api (default "https://app.koyeb.com")
```



* [koyeb deployments](#koyeb-deployments)	 - Deployments

## koyeb deployments logs

Get deployment logs

```
koyeb deployments logs NAME [flags]
```

### Options

```
  -h, --help          help for logs
  -t, --type string   Type of log (runtime,build)
```

### Options inherited from parent commands

```
  -c, --config string   config file (default is $HOME/.koyeb.yaml)
  -d, --debug           debug
      --full            show full id
  -o, --output output   output format (yaml,json,table)
      --token string    API token
      --url string      url of the api (default "https://app.koyeb.com")
```



* [koyeb deployments](#koyeb-deployments)	 - Deployments

## koyeb instances

Instances

### Options

```
  -h, --help   help for instances
```

### Options inherited from parent commands

```
  -c, --config string   config file (default is $HOME/.koyeb.yaml)
  -d, --debug           debug
      --full            show full id
  -o, --output output   output format (yaml,json,table)
      --token string    API token
      --url string      url of the api (default "https://app.koyeb.com")
```



* [koyeb](#koyeb)	 - Koyeb CLI
* [koyeb instances describe](#koyeb-instances-describe)	 - Describe instance
* [koyeb instances exec](#koyeb-instances-exec)	 - Run a command in the context of an instance
* [koyeb instances get](#koyeb-instances-get)	 - Get instance
* [koyeb instances list](#koyeb-instances-list)	 - List instances
* [koyeb instances logs](#koyeb-instances-logs)	 - Get instance logs

## koyeb instances describe

Describe instance

```
koyeb instances describe NAME [flags]
```

### Options

```
  -h, --help   help for describe
```

### Options inherited from parent commands

```
  -c, --config string   config file (default is $HOME/.koyeb.yaml)
  -d, --debug           debug
      --full            show full id
  -o, --output output   output format (yaml,json,table)
      --token string    API token
      --url string      url of the api (default "https://app.koyeb.com")
```



* [koyeb instances](#koyeb-instances)	 - Instances

## koyeb instances exec

Run a command in the context of an instance

```
koyeb instances exec NAME CMD -- [args...] [flags]
```

### Options

```
  -h, --help   help for exec
```

### Options inherited from parent commands

```
  -c, --config string   config file (default is $HOME/.koyeb.yaml)
  -d, --debug           debug
      --full            show full id
  -o, --output output   output format (yaml,json,table)
      --token string    API token
      --url string      url of the api (default "https://app.koyeb.com")
```



* [koyeb instances](#koyeb-instances)	 - Instances

## koyeb instances get

Get instance

```
koyeb instances get NAME [flags]
```

### Options

```
  -h, --help   help for get
```

### Options inherited from parent commands

```
  -c, --config string   config file (default is $HOME/.koyeb.yaml)
  -d, --debug           debug
      --full            show full id
  -o, --output output   output format (yaml,json,table)
      --token string    API token
      --url string      url of the api (default "https://app.koyeb.com")
```



* [koyeb instances](#koyeb-instances)	 - Instances

## koyeb instances list

List instances

```
koyeb instances list [flags]
```

### Options

```
      --app string       Filter on App id or name
  -h, --help             help for list
      --service string   Filter on Service id or name
```

### Options inherited from parent commands

```
  -c, --config string   config file (default is $HOME/.koyeb.yaml)
  -d, --debug           debug
      --full            show full id
  -o, --output output   output format (yaml,json,table)
      --token string    API token
      --url string      url of the api (default "https://app.koyeb.com")
```



* [koyeb instances](#koyeb-instances)	 - Instances

## koyeb instances logs

Get instance logs

```
koyeb instances logs NAME [flags]
```

### Options

```
  -h, --help   help for logs
```

### Options inherited from parent commands

```
  -c, --config string   config file (default is $HOME/.koyeb.yaml)
  -d, --debug           debug
      --full            show full id
  -o, --output output   output format (yaml,json,table)
      --token string    API token
      --url string      url of the api (default "https://app.koyeb.com")
```



* [koyeb instances](#koyeb-instances)	 - Instances

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
  -c, --config string   config file (default is $HOME/.koyeb.yaml)
  -d, --debug           debug
      --full            show full id
  -o, --output output   output format (yaml,json,table)
      --token string    API token
      --url string      url of the api (default "https://app.koyeb.com")
```



* [koyeb](#koyeb)	 - Koyeb CLI

