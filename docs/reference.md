---
title: "Koyeb CLI Reference"
shortTitle: Reference
description: "Discover all the commands available via the Koyeb CLI and how to use them to interact with the Koyeb serverless platform directly from the terminal."
---

# Koyeb CLI Reference

The Koyeb CLI allows you to interact with Koyeb directly from the terminal. This documentation references all commands and options available in the CLI.

If you have not installed the Koyeb CLI yet, please read the [installation guide](/build-and-deploy/cli/installation).
## koyeb

Koyeb CLI

### Options

```
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
```



* [koyeb apps](#koyeb-apps)	 - Apps
* [koyeb archives](#koyeb-archives)	 - Archives

* [koyeb databases](#koyeb-databases)	 - Databases
* [koyeb deploy](#koyeb-deploy)	 - Deploy a directory to Koyeb
* [koyeb deployments](#koyeb-deployments)	 - Deployments
* [koyeb domains](#koyeb-domains)	 - Domains
* [koyeb instances](#koyeb-instances)	 - Instances
* [koyeb login](#koyeb-login)	 - Login to your Koyeb account
* [koyeb metrics](#koyeb-metrics)	 - Metrics
* [koyeb organizations](#koyeb-organizations)	 - Organization
* [koyeb regional-deployments](#koyeb-regional-deployments)	 - Regional deployments
* [koyeb secrets](#koyeb-secrets)	 - Secrets
* [koyeb services](#koyeb-services)	 - Services
* [koyeb version](#koyeb-version)	 - Get version
* [koyeb volumes](#koyeb-volumes)	 - Manage persistent volumes

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
  -c, --config string         config file (default is $HOME/.koyeb.yaml)
  -d, --debug                 enable the debug output
      --debug-full            do not hide sensitive information (tokens) in the debug output
      --force-ascii           only output ascii characters (no unicode emojis)
      --full                  do not truncate output
      --organization string   organization ID
  -o, --output output         output format (yaml,json,table)
      --token string          API token
      --url string            url of the api (default "https://app.koyeb.com")
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
  -c, --config string         config file (default is $HOME/.koyeb.yaml)
  -d, --debug                 enable the debug output
      --debug-full            do not hide sensitive information (tokens) in the debug output
      --force-ascii           only output ascii characters (no unicode emojis)
      --full                  do not truncate output
      --organization string   organization ID
  -o, --output output         output format (yaml,json,table)
      --token string          API token
      --url string            url of the api (default "https://app.koyeb.com")
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
  -c, --config string         config file (default is $HOME/.koyeb.yaml)
  -d, --debug                 enable the debug output
      --debug-full            do not hide sensitive information (tokens) in the debug output
      --force-ascii           only output ascii characters (no unicode emojis)
      --full                  do not truncate output
      --organization string   organization ID
  -o, --output output         output format (yaml,json,table)
      --token string          API token
      --url string            url of the api (default "https://app.koyeb.com")
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
  -c, --config string         config file (default is $HOME/.koyeb.yaml)
  -d, --debug                 enable the debug output
      --debug-full            do not hide sensitive information (tokens) in the debug output
      --force-ascii           only output ascii characters (no unicode emojis)
      --full                  do not truncate output
      --organization string   organization ID
  -o, --output output         output format (yaml,json,table)
      --token string          API token
      --url string            url of the api (default "https://app.koyeb.com")
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
  -c, --config string         config file (default is $HOME/.koyeb.yaml)
  -d, --debug                 enable the debug output
      --debug-full            do not hide sensitive information (tokens) in the debug output
      --force-ascii           only output ascii characters (no unicode emojis)
      --full                  do not truncate output
      --organization string   organization ID
  -o, --output output         output format (yaml,json,table)
      --token string          API token
      --url string            url of the api (default "https://app.koyeb.com")
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
  -c, --config string         config file (default is $HOME/.koyeb.yaml)
  -d, --debug                 enable the debug output
      --debug-full            do not hide sensitive information (tokens) in the debug output
      --force-ascii           only output ascii characters (no unicode emojis)
      --full                  do not truncate output
      --organization string   organization ID
  -o, --output output         output format (yaml,json,table)
      --token string          API token
      --url string            url of the api (default "https://app.koyeb.com")
```



* [koyeb apps](#koyeb-apps)	 - Apps

## koyeb apps init

Create app and service

```
koyeb apps init NAME [flags]
```

### Examples

```
See examples of koyeb service create --help
```

### Options

```
      --archive string                           Archive ID to deploy
      --archive-builder string                   Builder to use, either "buildpack" (default) or "docker" (default "buildpack")
      --archive-buildpack-build-command string   Buid command
      --archive-buildpack-run-command string     Run command
      --archive-docker-args strings              Set arguments to the docker command. To provide multiple arguments, use the --archive-docker-args flag multiple times.
      --archive-docker-command string            Set the docker CMD explicitly. To provide arguments to the command, use the --archive-docker-args flag.
      --archive-docker-dockerfile string         Dockerfile path
      --archive-docker-entrypoint strings        Docker entrypoint
      --archive-docker-target string             Docker target
      --autoscaling-average-cpu int              Target CPU usage (in %) to trigger a scaling event. Set to 0 to disable CPU autoscaling.
      --autoscaling-average-mem int              Target memory usage (in %) to trigger a scaling event. Set to 0 to disable memory autoscaling.
      --autoscaling-concurrent-requests int      Target concurrent requests to trigger a scaling event. Set to 0 to disable concurrent requests autoscaling.
      --autoscaling-requests-per-second int      Target requests per second to trigger a scaling event. Set to 0 to disable requests per second autoscaling.
      --autoscaling-requests-response-time int   Target p95 response time to trigger a scaling event (in ms). Set to 0 to disable concurrent response time autoscaling.
      --checks strings                           Update service healthchecks (available for services of type "web" only)
                                                 For HTTP healthchecks, use the format <PORT>:http:<PATH>, for example --checks 8080:http:/health
                                                 For TCP healthchecks, use the format <PORT>:tcp, for example --checks 8080:tcp
                                                 To delete a healthcheck, use !PORT, for example --checks '!8080'
                                                 
      --docker string                            Docker image
      --docker-args strings                      Set arguments to the docker command. To provide multiple arguments, use the --docker-args flag multiple times.
      --docker-command string                    Set the docker CMD explicitly. To provide arguments to the command, use the --docker-args flag.
      --docker-entrypoint strings                Docker entrypoint. To provide multiple arguments, use the --docker-entrypoint flag multiple times.
      --docker-private-registry-secret string    Docker private registry secret
      --env strings                              Update service environment variables using the format KEY=VALUE, for example --env FOO=bar
                                                 To use the value of a secret as an environment variable, specify the secret name preceded by @, for example --env FOO=@bar
                                                 To delete an environment variable, prefix its name with '!', for example --env '!FOO'
                                                 
      --git string                               Git repository
      --git-branch string                        Git branch (default "main")
      --git-build-command string                 Buid command (legacy, prefer git-buildpack-build-command)
      --git-builder string                       Builder to use, either "buildpack" (default) or "docker" (default "buildpack")
      --git-buildpack-build-command string       Buid command
      --git-buildpack-run-command string         Run command
      --git-docker-args strings                  Set arguments to the docker command. To provide multiple arguments, use the --git-docker-args flag multiple times.
      --git-docker-command string                Set the docker CMD explicitly. To provide arguments to the command, use the --git-docker-args flag.
      --git-docker-dockerfile string             Dockerfile path
      --git-docker-entrypoint strings            Docker entrypoint
      --git-docker-target string                 Docker target
      --git-no-deploy-on-push                    Disable new deployments creation when code changes are pushed on the configured branch
      --git-run-command string                   Run command (legacy, prefer git-buildpack-run-command)
      --git-sha string                           Git commit SHA to deploy
      --git-workdir string                       Path to the sub-directory containing the code to build and deploy
  -h, --help                                     help for init
      --instance-type string                     Instance type (default "nano")
      --max-scale int                            Max scale (default 1)
      --min-scale int                            Min scale (default 1)
      --ports strings                            Update service ports (available for services of type "web" only) using the format PORT[:PROTOCOL], for example --port 8080:http
                                                 PROTOCOL defaults to "http". Supported protocols are "http", "http2" and "tcp"
                                                 To delete an exposed port, prefix its number with '!', for example --port '!80'
                                                 
      --privileged                               Whether the service container should run in privileged mode
      --regions strings                          Add a region where the service is deployed. You can specify this flag multiple times to deploy the service in multiple regions.
                                                 To update a service and remove a region, prefix the region name with '!', for example --region '!par'
                                                 If the region is not specified on service creation, the service is deployed in fra
                                                 
      --routes strings                           Update service routes (available for services of type "web" only) using the format PATH[:PORT], for example '/foo:8080'
                                                 PORT defaults to 8000
                                                 To delete a route, use '!PATH', for example --route '!/foo'
                                                 
      --scale int                                Set both min-scale and max-scale (default 1)
      --skip-cache                               Whether to use the cache when building the service
      --type string                              Service type, either "web" or "worker" (default "web")
```

### Options inherited from parent commands

```
  -c, --config string         config file (default is $HOME/.koyeb.yaml)
  -d, --debug                 enable the debug output
      --debug-full            do not hide sensitive information (tokens) in the debug output
      --force-ascii           only output ascii characters (no unicode emojis)
      --full                  do not truncate output
      --organization string   organization ID
  -o, --output output         output format (yaml,json,table)
      --token string          API token
      --url string            url of the api (default "https://app.koyeb.com")
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
  -c, --config string         config file (default is $HOME/.koyeb.yaml)
  -d, --debug                 enable the debug output
      --debug-full            do not hide sensitive information (tokens) in the debug output
      --force-ascii           only output ascii characters (no unicode emojis)
      --full                  do not truncate output
      --organization string   organization ID
  -o, --output output         output format (yaml,json,table)
      --token string          API token
      --url string            url of the api (default "https://app.koyeb.com")
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
  -c, --config string         config file (default is $HOME/.koyeb.yaml)
  -d, --debug                 enable the debug output
      --debug-full            do not hide sensitive information (tokens) in the debug output
      --force-ascii           only output ascii characters (no unicode emojis)
      --full                  do not truncate output
      --organization string   organization ID
  -o, --output output         output format (yaml,json,table)
      --token string          API token
      --url string            url of the api (default "https://app.koyeb.com")
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
  -c, --config string         config file (default is $HOME/.koyeb.yaml)
  -d, --debug                 enable the debug output
      --debug-full            do not hide sensitive information (tokens) in the debug output
      --force-ascii           only output ascii characters (no unicode emojis)
      --full                  do not truncate output
      --organization string   organization ID
  -o, --output output         output format (yaml,json,table)
      --token string          API token
      --url string            url of the api (default "https://app.koyeb.com")
```



* [koyeb apps](#koyeb-apps)	 - Apps

## koyeb apps update

Update app

```
koyeb apps update NAME [flags]
```

### Options

```
  -D, --domain string   Change the subdomain of the app (only specify the subdomain, skipping ".koyeb.app")
  -h, --help            help for update
  -n, --name string     Change the name of the app
```

### Options inherited from parent commands

```
  -c, --config string         config file (default is $HOME/.koyeb.yaml)
  -d, --debug                 enable the debug output
      --debug-full            do not hide sensitive information (tokens) in the debug output
      --force-ascii           only output ascii characters (no unicode emojis)
      --full                  do not truncate output
      --organization string   organization ID
  -o, --output output         output format (yaml,json,table)
      --token string          API token
      --url string            url of the api (default "https://app.koyeb.com")
```



* [koyeb apps](#koyeb-apps)	 - Apps

## koyeb archives

Archives

### Options

```
  -h, --help   help for archives
```

### Options inherited from parent commands

```
  -c, --config string         config file (default is $HOME/.koyeb.yaml)
  -d, --debug                 enable the debug output
      --debug-full            do not hide sensitive information (tokens) in the debug output
      --force-ascii           only output ascii characters (no unicode emojis)
      --full                  do not truncate output
      --organization string   organization ID
  -o, --output output         output format (yaml,json,table)
      --token string          API token
      --url string            url of the api (default "https://app.koyeb.com")
```



* [koyeb](#koyeb)	 - Koyeb CLI
* [koyeb archives create](#koyeb-archives-create)	 - Create archive

## koyeb archives create

Create archive

```
koyeb archives create NAME [flags]
```

### Options

```
  -h, --help   help for create
```

### Options inherited from parent commands

```
  -c, --config string         config file (default is $HOME/.koyeb.yaml)
  -d, --debug                 enable the debug output
      --debug-full            do not hide sensitive information (tokens) in the debug output
      --force-ascii           only output ascii characters (no unicode emojis)
      --full                  do not truncate output
      --organization string   organization ID
  -o, --output output         output format (yaml,json,table)
      --token string          API token
      --url string            url of the api (default "https://app.koyeb.com")
```



* [koyeb archives](#koyeb-archives)	 - Archives

## koyeb deploy

Deploy a directory to Koyeb

```
koyeb deploy <path> <app>/<service> [flags]
```

### Options

```
      --app string                               Service application. Can also be provided in the service name with the format <app>/<service>
      --archive string                           Archive ID to deploy
      --archive-builder string                   Builder to use, either "buildpack" (default) or "docker" (default "buildpack")
      --archive-buildpack-build-command string   Buid command
      --archive-buildpack-run-command string     Run command
      --archive-docker-args strings              Set arguments to the docker command. To provide multiple arguments, use the --archive-docker-args flag multiple times.
      --archive-docker-command string            Set the docker CMD explicitly. To provide arguments to the command, use the --archive-docker-args flag.
      --archive-docker-dockerfile string         Dockerfile path
      --archive-docker-entrypoint strings        Docker entrypoint
      --archive-docker-target string             Docker target
      --autoscaling-average-cpu int              Target CPU usage (in %) to trigger a scaling event. Set to 0 to disable CPU autoscaling.
      --autoscaling-average-mem int              Target memory usage (in %) to trigger a scaling event. Set to 0 to disable memory autoscaling.
      --autoscaling-concurrent-requests int      Target concurrent requests to trigger a scaling event. Set to 0 to disable concurrent requests autoscaling.
      --autoscaling-requests-per-second int      Target requests per second to trigger a scaling event. Set to 0 to disable requests per second autoscaling.
      --autoscaling-requests-response-time int   Target p95 response time to trigger a scaling event (in ms). Set to 0 to disable concurrent response time autoscaling.
      --checks strings                           Update service healthchecks (available for services of type "web" only)
                                                 For HTTP healthchecks, use the format <PORT>:http:<PATH>, for example --checks 8080:http:/health
                                                 For TCP healthchecks, use the format <PORT>:tcp, for example --checks 8080:tcp
                                                 To delete a healthcheck, use !PORT, for example --checks '!8080'
                                                 
      --env strings                              Update service environment variables using the format KEY=VALUE, for example --env FOO=bar
                                                 To use the value of a secret as an environment variable, specify the secret name preceded by @, for example --env FOO=@bar
                                                 To delete an environment variable, prefix its name with '!', for example --env '!FOO'
                                                 
  -h, --help                                     help for deploy
      --instance-type string                     Instance type (default "nano")
      --max-scale int                            Max scale (default 1)
      --min-scale int                            Min scale (default 1)
      --ports strings                            Update service ports (available for services of type "web" only) using the format PORT[:PROTOCOL], for example --port 8080:http
                                                 PROTOCOL defaults to "http". Supported protocols are "http", "http2" and "tcp"
                                                 To delete an exposed port, prefix its number with '!', for example --port '!80'
                                                 
      --privileged                               Whether the service container should run in privileged mode
      --regions strings                          Add a region where the service is deployed. You can specify this flag multiple times to deploy the service in multiple regions.
                                                 To update a service and remove a region, prefix the region name with '!', for example --region '!par'
                                                 If the region is not specified on service creation, the service is deployed in fra
                                                 
      --routes strings                           Update service routes (available for services of type "web" only) using the format PATH[:PORT], for example '/foo:8080'
                                                 PORT defaults to 8000
                                                 To delete a route, use '!PATH', for example --route '!/foo'
                                                 
      --scale int                                Set both min-scale and max-scale (default 1)
      --skip-cache                               Whether to use the cache when building the service
      --type string                              Service type, either "web" or "worker" (default "web")
```

### Options inherited from parent commands

```
  -c, --config string         config file (default is $HOME/.koyeb.yaml)
  -d, --debug                 enable the debug output
      --debug-full            do not hide sensitive information (tokens) in the debug output
      --force-ascii           only output ascii characters (no unicode emojis)
      --full                  do not truncate output
      --organization string   organization ID
  -o, --output output         output format (yaml,json,table)
      --token string          API token
      --url string            url of the api (default "https://app.koyeb.com")
```



* [koyeb](#koyeb)	 - Koyeb CLI

## koyeb domains

Domains

### Options

```
  -h, --help   help for domains
```

### Options inherited from parent commands

```
  -c, --config string         config file (default is $HOME/.koyeb.yaml)
  -d, --debug                 enable the debug output
      --debug-full            do not hide sensitive information (tokens) in the debug output
      --force-ascii           only output ascii characters (no unicode emojis)
      --full                  do not truncate output
      --organization string   organization ID
  -o, --output output         output format (yaml,json,table)
      --token string          API token
      --url string            url of the api (default "https://app.koyeb.com")
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
  -c, --config string         config file (default is $HOME/.koyeb.yaml)
  -d, --debug                 enable the debug output
      --debug-full            do not hide sensitive information (tokens) in the debug output
      --force-ascii           only output ascii characters (no unicode emojis)
      --full                  do not truncate output
      --organization string   organization ID
  -o, --output output         output format (yaml,json,table)
      --token string          API token
      --url string            url of the api (default "https://app.koyeb.com")
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
  -c, --config string         config file (default is $HOME/.koyeb.yaml)
  -d, --debug                 enable the debug output
      --debug-full            do not hide sensitive information (tokens) in the debug output
      --force-ascii           only output ascii characters (no unicode emojis)
      --full                  do not truncate output
      --organization string   organization ID
  -o, --output output         output format (yaml,json,table)
      --token string          API token
      --url string            url of the api (default "https://app.koyeb.com")
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
  -c, --config string         config file (default is $HOME/.koyeb.yaml)
  -d, --debug                 enable the debug output
      --debug-full            do not hide sensitive information (tokens) in the debug output
      --force-ascii           only output ascii characters (no unicode emojis)
      --full                  do not truncate output
      --organization string   organization ID
  -o, --output output         output format (yaml,json,table)
      --token string          API token
      --url string            url of the api (default "https://app.koyeb.com")
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
  -c, --config string         config file (default is $HOME/.koyeb.yaml)
  -d, --debug                 enable the debug output
      --debug-full            do not hide sensitive information (tokens) in the debug output
      --force-ascii           only output ascii characters (no unicode emojis)
      --full                  do not truncate output
      --organization string   organization ID
  -o, --output output         output format (yaml,json,table)
      --token string          API token
      --url string            url of the api (default "https://app.koyeb.com")
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
  -c, --config string         config file (default is $HOME/.koyeb.yaml)
  -d, --debug                 enable the debug output
      --debug-full            do not hide sensitive information (tokens) in the debug output
      --force-ascii           only output ascii characters (no unicode emojis)
      --full                  do not truncate output
      --organization string   organization ID
  -o, --output output         output format (yaml,json,table)
      --token string          API token
      --url string            url of the api (default "https://app.koyeb.com")
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
  -c, --config string         config file (default is $HOME/.koyeb.yaml)
  -d, --debug                 enable the debug output
      --debug-full            do not hide sensitive information (tokens) in the debug output
      --force-ascii           only output ascii characters (no unicode emojis)
      --full                  do not truncate output
      --organization string   organization ID
  -o, --output output         output format (yaml,json,table)
      --token string          API token
      --url string            url of the api (default "https://app.koyeb.com")
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
  -c, --config string         config file (default is $HOME/.koyeb.yaml)
  -d, --debug                 enable the debug output
      --debug-full            do not hide sensitive information (tokens) in the debug output
      --force-ascii           only output ascii characters (no unicode emojis)
      --full                  do not truncate output
      --organization string   organization ID
  -o, --output output         output format (yaml,json,table)
      --token string          API token
      --url string            url of the api (default "https://app.koyeb.com")
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
  -c, --config string         config file (default is $HOME/.koyeb.yaml)
  -d, --debug                 enable the debug output
      --debug-full            do not hide sensitive information (tokens) in the debug output
      --force-ascii           only output ascii characters (no unicode emojis)
      --full                  do not truncate output
      --organization string   organization ID
  -o, --output output         output format (yaml,json,table)
      --token string          API token
      --url string            url of the api (default "https://app.koyeb.com")
```



* [koyeb domains](#koyeb-domains)	 - Domains

## koyeb organizations

Organization

### Options

```
  -h, --help   help for organizations
```

### Options inherited from parent commands

```
  -c, --config string         config file (default is $HOME/.koyeb.yaml)
  -d, --debug                 enable the debug output
      --debug-full            do not hide sensitive information (tokens) in the debug output
      --force-ascii           only output ascii characters (no unicode emojis)
      --full                  do not truncate output
      --organization string   organization ID
  -o, --output output         output format (yaml,json,table)
      --token string          API token
      --url string            url of the api (default "https://app.koyeb.com")
```



* [koyeb](#koyeb)	 - Koyeb CLI
* [koyeb organizations list](#koyeb-organizations-list)	 - List organizations
* [koyeb organizations switch](#koyeb-organizations-switch)	 - Switch the CLI context to another organization

## koyeb organizations list

List organizations

```
koyeb organizations list [flags]
```

### Options

```
  -h, --help   help for list
```

### Options inherited from parent commands

```
  -c, --config string         config file (default is $HOME/.koyeb.yaml)
  -d, --debug                 enable the debug output
      --debug-full            do not hide sensitive information (tokens) in the debug output
      --force-ascii           only output ascii characters (no unicode emojis)
      --full                  do not truncate output
      --organization string   organization ID
  -o, --output output         output format (yaml,json,table)
      --token string          API token
      --url string            url of the api (default "https://app.koyeb.com")
```



* [koyeb organizations](#koyeb-organizations)	 - Organization

## koyeb organizations switch

Switch the CLI context to another organization

```
koyeb organizations switch [flags]
```

### Options

```
  -h, --help   help for switch
```

### Options inherited from parent commands

```
  -c, --config string         config file (default is $HOME/.koyeb.yaml)
  -d, --debug                 enable the debug output
      --debug-full            do not hide sensitive information (tokens) in the debug output
      --force-ascii           only output ascii characters (no unicode emojis)
      --full                  do not truncate output
      --organization string   organization ID
  -o, --output output         output format (yaml,json,table)
      --token string          API token
      --url string            url of the api (default "https://app.koyeb.com")
```



* [koyeb organizations](#koyeb-organizations)	 - Organization

## koyeb secrets

Secrets

### Options

```
  -h, --help   help for secrets
```

### Options inherited from parent commands

```
  -c, --config string         config file (default is $HOME/.koyeb.yaml)
  -d, --debug                 enable the debug output
      --debug-full            do not hide sensitive information (tokens) in the debug output
      --force-ascii           only output ascii characters (no unicode emojis)
      --full                  do not truncate output
      --organization string   organization ID
  -o, --output output         output format (yaml,json,table)
      --token string          API token
      --url string            url of the api (default "https://app.koyeb.com")
```



* [koyeb](#koyeb)	 - Koyeb CLI
* [koyeb secrets create](#koyeb-secrets-create)	 - Create secret
* [koyeb secrets delete](#koyeb-secrets-delete)	 - Delete secret
* [koyeb secrets describe](#koyeb-secrets-describe)	 - Describe secret
* [koyeb secrets get](#koyeb-secrets-get)	 - Get secret
* [koyeb secrets list](#koyeb-secrets-list)	 - List secrets
* [koyeb secrets reveal](#koyeb-secrets-reveal)	 - Show secret value
* [koyeb secrets update](#koyeb-secrets-update)	 - Update secret

## koyeb secrets create

Create secret

```
koyeb secrets create NAME [flags]
```

### Options

```
  -h, --help                       help for create
      --registry-keyfile string    Registry URL. Only valid with --type=registry-gcp, otherwise ignored
      --registry-name string       Registry name. Only valid with --type=registry-azure, otherwise ignored
      --registry-url string        Registry URL. Only valid with --type=registry-private and --type=registry-gcp, otherwise ignored
      --registry-username string   Registry username. Only valid with --type=registry-*
      --type type                  Secret type (simple, registry-dockerhub, registry-private, registry-digital-ocean, registry-gitlab, registry-gcp, registry-azure) (default simple)
  -v, --value string               Secret Value
      --value-from-stdin           Secret Value from stdin
```

### Options inherited from parent commands

```
  -c, --config string         config file (default is $HOME/.koyeb.yaml)
  -d, --debug                 enable the debug output
      --debug-full            do not hide sensitive information (tokens) in the debug output
      --force-ascii           only output ascii characters (no unicode emojis)
      --full                  do not truncate output
      --organization string   organization ID
  -o, --output output         output format (yaml,json,table)
      --token string          API token
      --url string            url of the api (default "https://app.koyeb.com")
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
  -c, --config string         config file (default is $HOME/.koyeb.yaml)
  -d, --debug                 enable the debug output
      --debug-full            do not hide sensitive information (tokens) in the debug output
      --force-ascii           only output ascii characters (no unicode emojis)
      --full                  do not truncate output
      --organization string   organization ID
  -o, --output output         output format (yaml,json,table)
      --token string          API token
      --url string            url of the api (default "https://app.koyeb.com")
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
  -c, --config string         config file (default is $HOME/.koyeb.yaml)
  -d, --debug                 enable the debug output
      --debug-full            do not hide sensitive information (tokens) in the debug output
      --force-ascii           only output ascii characters (no unicode emojis)
      --full                  do not truncate output
      --organization string   organization ID
  -o, --output output         output format (yaml,json,table)
      --token string          API token
      --url string            url of the api (default "https://app.koyeb.com")
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
  -c, --config string         config file (default is $HOME/.koyeb.yaml)
  -d, --debug                 enable the debug output
      --debug-full            do not hide sensitive information (tokens) in the debug output
      --force-ascii           only output ascii characters (no unicode emojis)
      --full                  do not truncate output
      --organization string   organization ID
  -o, --output output         output format (yaml,json,table)
      --token string          API token
      --url string            url of the api (default "https://app.koyeb.com")
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
  -c, --config string         config file (default is $HOME/.koyeb.yaml)
  -d, --debug                 enable the debug output
      --debug-full            do not hide sensitive information (tokens) in the debug output
      --force-ascii           only output ascii characters (no unicode emojis)
      --full                  do not truncate output
      --organization string   organization ID
  -o, --output output         output format (yaml,json,table)
      --token string          API token
      --url string            url of the api (default "https://app.koyeb.com")
```



* [koyeb secrets](#koyeb-secrets)	 - Secrets

## koyeb secrets reveal

Show secret value

```
koyeb secrets reveal NAME [flags]
```

### Options

```
  -h, --help   help for reveal
```

### Options inherited from parent commands

```
  -c, --config string         config file (default is $HOME/.koyeb.yaml)
  -d, --debug                 enable the debug output
      --debug-full            do not hide sensitive information (tokens) in the debug output
      --force-ascii           only output ascii characters (no unicode emojis)
      --full                  do not truncate output
      --organization string   organization ID
  -o, --output output         output format (yaml,json,table)
      --token string          API token
      --url string            url of the api (default "https://app.koyeb.com")
```



* [koyeb secrets](#koyeb-secrets)	 - Secrets

## koyeb secrets update

Update secret

```
koyeb secrets update NAME [flags]
```

### Options

```
  -h, --help                       help for update
      --registry-keyfile string    Registry URL. Only valid with --type=registry-gcp, otherwise ignored
      --registry-name string       Registry name. Only valid with --type=registry-azure, otherwise ignored
      --registry-url string        Registry URL. Only valid with --type=registry-private and --type=registry-gcp, otherwise ignored
      --registry-username string   Registry username. Only valid with --type=registry-*
  -v, --value string               Secret Value
      --value-from-stdin           Secret Value from stdin
```

### Options inherited from parent commands

```
  -c, --config string         config file (default is $HOME/.koyeb.yaml)
  -d, --debug                 enable the debug output
      --debug-full            do not hide sensitive information (tokens) in the debug output
      --force-ascii           only output ascii characters (no unicode emojis)
      --full                  do not truncate output
      --organization string   organization ID
  -o, --output output         output format (yaml,json,table)
      --token string          API token
      --url string            url of the api (default "https://app.koyeb.com")
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
  -c, --config string         config file (default is $HOME/.koyeb.yaml)
  -d, --debug                 enable the debug output
      --debug-full            do not hide sensitive information (tokens) in the debug output
      --force-ascii           only output ascii characters (no unicode emojis)
      --full                  do not truncate output
      --organization string   organization ID
  -o, --output output         output format (yaml,json,table)
      --token string          API token
      --url string            url of the api (default "https://app.koyeb.com")
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

### Examples

```

# Deploy a nginx docker image, listening on port 80
$> koyeb service create myservice --app myapp --docker nginx --port 80

# Deploy a nginx docker image and set the docker CMD explicitly, equivalent to docker CMD ["nginx", "-g", "daemon off;"]
$> koyeb service create myservice --app myapp --docker nginx --port 80 --docker-command nginx --docker-args '-g' --docker-args 'daemon off;'

# Build and deploy a GitHub repository using buildpack (default), set the environment variable PORT, and expose the port 9000 to the root route
$> koyeb service create myservice --app myapp --git github.com/koyeb/example-flask --git-branch main --env PORT=9000 --port 9000:http --route /:9000

# Build and deploy a GitHub repository using docker
$> koyeb service create myservice --app myapp --git github.com/org/name --git-branch main --git-builder docker

# Create a docker service, only accessible from the mesh (--route is not automatically created for TCP ports)
$> koyeb service create myservice --app myapp --docker nginx --port 80:tcp

```

### Options

```
  -a, --app string                               Service application
      --archive string                           Archive ID to deploy
      --archive-builder string                   Builder to use, either "buildpack" (default) or "docker" (default "buildpack")
      --archive-buildpack-build-command string   Buid command
      --archive-buildpack-run-command string     Run command
      --archive-docker-args strings              Set arguments to the docker command. To provide multiple arguments, use the --archive-docker-args flag multiple times.
      --archive-docker-command string            Set the docker CMD explicitly. To provide arguments to the command, use the --archive-docker-args flag.
      --archive-docker-dockerfile string         Dockerfile path
      --archive-docker-entrypoint strings        Docker entrypoint
      --archive-docker-target string             Docker target
      --autoscaling-average-cpu int              Target CPU usage (in %) to trigger a scaling event. Set to 0 to disable CPU autoscaling.
      --autoscaling-average-mem int              Target memory usage (in %) to trigger a scaling event. Set to 0 to disable memory autoscaling.
      --autoscaling-concurrent-requests int      Target concurrent requests to trigger a scaling event. Set to 0 to disable concurrent requests autoscaling.
      --autoscaling-requests-per-second int      Target requests per second to trigger a scaling event. Set to 0 to disable requests per second autoscaling.
      --autoscaling-requests-response-time int   Target p95 response time to trigger a scaling event (in ms). Set to 0 to disable concurrent response time autoscaling.
      --checks strings                           Update service healthchecks (available for services of type "web" only)
                                                 For HTTP healthchecks, use the format <PORT>:http:<PATH>, for example --checks 8080:http:/health
                                                 For TCP healthchecks, use the format <PORT>:tcp, for example --checks 8080:tcp
                                                 To delete a healthcheck, use !PORT, for example --checks '!8080'
                                                 
      --docker string                            Docker image
      --docker-args strings                      Set arguments to the docker command. To provide multiple arguments, use the --docker-args flag multiple times.
      --docker-command string                    Set the docker CMD explicitly. To provide arguments to the command, use the --docker-args flag.
      --docker-entrypoint strings                Docker entrypoint. To provide multiple arguments, use the --docker-entrypoint flag multiple times.
      --docker-private-registry-secret string    Docker private registry secret
      --env strings                              Update service environment variables using the format KEY=VALUE, for example --env FOO=bar
                                                 To use the value of a secret as an environment variable, specify the secret name preceded by @, for example --env FOO=@bar
                                                 To delete an environment variable, prefix its name with '!', for example --env '!FOO'
                                                 
      --git string                               Git repository
      --git-branch string                        Git branch (default "main")
      --git-build-command string                 Buid command (legacy, prefer git-buildpack-build-command)
      --git-builder string                       Builder to use, either "buildpack" (default) or "docker" (default "buildpack")
      --git-buildpack-build-command string       Buid command
      --git-buildpack-run-command string         Run command
      --git-docker-args strings                  Set arguments to the docker command. To provide multiple arguments, use the --git-docker-args flag multiple times.
      --git-docker-command string                Set the docker CMD explicitly. To provide arguments to the command, use the --git-docker-args flag.
      --git-docker-dockerfile string             Dockerfile path
      --git-docker-entrypoint strings            Docker entrypoint
      --git-docker-target string                 Docker target
      --git-no-deploy-on-push                    Disable new deployments creation when code changes are pushed on the configured branch
      --git-run-command string                   Run command (legacy, prefer git-buildpack-run-command)
      --git-sha string                           Git commit SHA to deploy
      --git-workdir string                       Path to the sub-directory containing the code to build and deploy
  -h, --help                                     help for create
      --instance-type string                     Instance type (default "nano")
      --max-scale int                            Max scale (default 1)
      --min-scale int                            Min scale (default 1)
      --ports strings                            Update service ports (available for services of type "web" only) using the format PORT[:PROTOCOL], for example --port 8080:http
                                                 PROTOCOL defaults to "http". Supported protocols are "http", "http2" and "tcp"
                                                 To delete an exposed port, prefix its number with '!', for example --port '!80'
                                                 
      --privileged                               Whether the service container should run in privileged mode
      --regions strings                          Add a region where the service is deployed. You can specify this flag multiple times to deploy the service in multiple regions.
                                                 To update a service and remove a region, prefix the region name with '!', for example --region '!par'
                                                 If the region is not specified on service creation, the service is deployed in fra
                                                 
      --routes strings                           Update service routes (available for services of type "web" only) using the format PATH[:PORT], for example '/foo:8080'
                                                 PORT defaults to 8000
                                                 To delete a route, use '!PATH', for example --route '!/foo'
                                                 
      --scale int                                Set both min-scale and max-scale (default 1)
      --skip-cache                               Whether to use the cache when building the service
      --type string                              Service type, either "web" or "worker" (default "web")
```

### Options inherited from parent commands

```
  -c, --config string         config file (default is $HOME/.koyeb.yaml)
  -d, --debug                 enable the debug output
      --debug-full            do not hide sensitive information (tokens) in the debug output
      --force-ascii           only output ascii characters (no unicode emojis)
      --full                  do not truncate output
      --organization string   organization ID
  -o, --output output         output format (yaml,json,table)
      --token string          API token
      --url string            url of the api (default "https://app.koyeb.com")
```



* [koyeb services](#koyeb-services)	 - Services

## koyeb services delete

Delete service

```
koyeb services delete NAME [flags]
```

### Options

```
  -a, --app string   Service application
  -h, --help         help for delete
```

### Options inherited from parent commands

```
  -c, --config string         config file (default is $HOME/.koyeb.yaml)
  -d, --debug                 enable the debug output
      --debug-full            do not hide sensitive information (tokens) in the debug output
      --force-ascii           only output ascii characters (no unicode emojis)
      --full                  do not truncate output
      --organization string   organization ID
  -o, --output output         output format (yaml,json,table)
      --token string          API token
      --url string            url of the api (default "https://app.koyeb.com")
```



* [koyeb services](#koyeb-services)	 - Services

## koyeb services describe

Describe service

```
koyeb services describe NAME [flags]
```

### Options

```
  -a, --app string   Service application
  -h, --help         help for describe
```

### Options inherited from parent commands

```
  -c, --config string         config file (default is $HOME/.koyeb.yaml)
  -d, --debug                 enable the debug output
      --debug-full            do not hide sensitive information (tokens) in the debug output
      --force-ascii           only output ascii characters (no unicode emojis)
      --full                  do not truncate output
      --organization string   organization ID
  -o, --output output         output format (yaml,json,table)
      --token string          API token
      --url string            url of the api (default "https://app.koyeb.com")
```



* [koyeb services](#koyeb-services)	 - Services

## koyeb services exec

Run a command in the context of an instance selected among the service instances

```
koyeb services exec NAME CMD -- [args...] [flags]
```

### Options

```
  -a, --app string   Service application
  -h, --help         help for exec
```

### Options inherited from parent commands

```
  -c, --config string         config file (default is $HOME/.koyeb.yaml)
  -d, --debug                 enable the debug output
      --debug-full            do not hide sensitive information (tokens) in the debug output
      --force-ascii           only output ascii characters (no unicode emojis)
      --full                  do not truncate output
      --organization string   organization ID
  -o, --output output         output format (yaml,json,table)
      --token string          API token
      --url string            url of the api (default "https://app.koyeb.com")
```



* [koyeb services](#koyeb-services)	 - Services

## koyeb services get

Get service

```
koyeb services get NAME [flags]
```

### Options

```
  -a, --app string   Service application
  -h, --help         help for get
```

### Options inherited from parent commands

```
  -c, --config string         config file (default is $HOME/.koyeb.yaml)
  -d, --debug                 enable the debug output
      --debug-full            do not hide sensitive information (tokens) in the debug output
      --force-ascii           only output ascii characters (no unicode emojis)
      --full                  do not truncate output
      --organization string   organization ID
  -o, --output output         output format (yaml,json,table)
      --token string          API token
      --url string            url of the api (default "https://app.koyeb.com")
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
  -c, --config string         config file (default is $HOME/.koyeb.yaml)
  -d, --debug                 enable the debug output
      --debug-full            do not hide sensitive information (tokens) in the debug output
      --force-ascii           only output ascii characters (no unicode emojis)
      --full                  do not truncate output
      --organization string   organization ID
  -o, --output output         output format (yaml,json,table)
      --token string          API token
      --url string            url of the api (default "https://app.koyeb.com")
```



* [koyeb services](#koyeb-services)	 - Services

## koyeb services logs

Get the service logs

```
koyeb services logs NAME [flags]
```

### Options

```
  -a, --app string        Service application
  -h, --help              help for logs
      --instance string   Instance
  -t, --type string       Type (runtime, build)
```

### Options inherited from parent commands

```
  -c, --config string         config file (default is $HOME/.koyeb.yaml)
  -d, --debug                 enable the debug output
      --debug-full            do not hide sensitive information (tokens) in the debug output
      --force-ascii           only output ascii characters (no unicode emojis)
      --full                  do not truncate output
      --organization string   organization ID
  -o, --output output         output format (yaml,json,table)
      --token string          API token
      --url string            url of the api (default "https://app.koyeb.com")
```



* [koyeb services](#koyeb-services)	 - Services

## koyeb services pause

Pause service

```
koyeb services pause NAME [flags]
```

### Options

```
  -a, --app string   Service application
  -h, --help         help for pause
```

### Options inherited from parent commands

```
  -c, --config string         config file (default is $HOME/.koyeb.yaml)
  -d, --debug                 enable the debug output
      --debug-full            do not hide sensitive information (tokens) in the debug output
      --force-ascii           only output ascii characters (no unicode emojis)
      --full                  do not truncate output
      --organization string   organization ID
  -o, --output output         output format (yaml,json,table)
      --token string          API token
      --url string            url of the api (default "https://app.koyeb.com")
```



* [koyeb services](#koyeb-services)	 - Services

## koyeb services redeploy

Redeploy service

```
koyeb services redeploy NAME [flags]
```

### Options

```
  -a, --app string   Service application
  -h, --help         help for redeploy
      --skip-build   If there has been at least one past successfully build deployment, use the last one instead of rebuilding. WARNING: this can lead to unexpected behavior if the build depends, for example, on environment variables.
      --use-cache    Use cache to redeploy
```

### Options inherited from parent commands

```
  -c, --config string         config file (default is $HOME/.koyeb.yaml)
  -d, --debug                 enable the debug output
      --debug-full            do not hide sensitive information (tokens) in the debug output
      --force-ascii           only output ascii characters (no unicode emojis)
      --full                  do not truncate output
      --organization string   organization ID
  -o, --output output         output format (yaml,json,table)
      --token string          API token
      --url string            url of the api (default "https://app.koyeb.com")
```



* [koyeb services](#koyeb-services)	 - Services

## koyeb services resume

Resume service

```
koyeb services resume NAME [flags]
```

### Options

```
  -a, --app string   Service application
  -h, --help         help for resume
```

### Options inherited from parent commands

```
  -c, --config string         config file (default is $HOME/.koyeb.yaml)
  -d, --debug                 enable the debug output
      --debug-full            do not hide sensitive information (tokens) in the debug output
      --force-ascii           only output ascii characters (no unicode emojis)
      --full                  do not truncate output
      --organization string   organization ID
  -o, --output output         output format (yaml,json,table)
      --token string          API token
      --url string            url of the api (default "https://app.koyeb.com")
```



* [koyeb services](#koyeb-services)	 - Services

## koyeb services update

Update service

```
koyeb services update NAME [flags]
```

### Examples

```

# Update the service "myservice" in the app "myapp", upsert the environment variable PORT and delete the environment variable DEBUG
$> koyeb service update myapp/myservice --env PORT=8001 --env '!DEBUG'

# Update the docker command of the service "myservice" in the app "myapp", equivalent to docker CMD ["nginx", "-g", "daemon off;"]
$> koyeb service update myapp/myservice --docker-command nginx --docker-args '-g' --docker-args 'daemon off;'

# Given a public service configured with the port 80:http and the route /:80, update it to make the service private, ie. only
# accessible from the mesh, by changing the port's protocol and removing the route
$> koyeb service update myapp/myservice --port 80:tcp --route '!/'

```

### Options

```
  -a, --app string                               Service application
      --archive string                           Archive ID to deploy
      --archive-builder string                   Builder to use, either "buildpack" (default) or "docker" (default "buildpack")
      --archive-buildpack-build-command string   Buid command
      --archive-buildpack-run-command string     Run command
      --archive-docker-args strings              Set arguments to the docker command. To provide multiple arguments, use the --archive-docker-args flag multiple times.
      --archive-docker-command string            Set the docker CMD explicitly. To provide arguments to the command, use the --archive-docker-args flag.
      --archive-docker-dockerfile string         Dockerfile path
      --archive-docker-entrypoint strings        Docker entrypoint
      --archive-docker-target string             Docker target
      --autoscaling-average-cpu int              Target CPU usage (in %) to trigger a scaling event. Set to 0 to disable CPU autoscaling.
      --autoscaling-average-mem int              Target memory usage (in %) to trigger a scaling event. Set to 0 to disable memory autoscaling.
      --autoscaling-concurrent-requests int      Target concurrent requests to trigger a scaling event. Set to 0 to disable concurrent requests autoscaling.
      --autoscaling-requests-per-second int      Target requests per second to trigger a scaling event. Set to 0 to disable requests per second autoscaling.
      --autoscaling-requests-response-time int   Target p95 response time to trigger a scaling event (in ms). Set to 0 to disable concurrent response time autoscaling.
      --checks strings                           Update service healthchecks (available for services of type "web" only)
                                                 For HTTP healthchecks, use the format <PORT>:http:<PATH>, for example --checks 8080:http:/health
                                                 For TCP healthchecks, use the format <PORT>:tcp, for example --checks 8080:tcp
                                                 To delete a healthcheck, use !PORT, for example --checks '!8080'
                                                 
      --docker string                            Docker image
      --docker-args strings                      Set arguments to the docker command. To provide multiple arguments, use the --docker-args flag multiple times.
      --docker-command string                    Set the docker CMD explicitly. To provide arguments to the command, use the --docker-args flag.
      --docker-entrypoint strings                Docker entrypoint. To provide multiple arguments, use the --docker-entrypoint flag multiple times.
      --docker-private-registry-secret string    Docker private registry secret
      --env strings                              Update service environment variables using the format KEY=VALUE, for example --env FOO=bar
                                                 To use the value of a secret as an environment variable, specify the secret name preceded by @, for example --env FOO=@bar
                                                 To delete an environment variable, prefix its name with '!', for example --env '!FOO'
                                                 
      --git string                               Git repository
      --git-branch string                        Git branch (default "main")
      --git-build-command string                 Buid command (legacy, prefer git-buildpack-build-command)
      --git-builder string                       Builder to use, either "buildpack" (default) or "docker" (default "buildpack")
      --git-buildpack-build-command string       Buid command
      --git-buildpack-run-command string         Run command
      --git-docker-args strings                  Set arguments to the docker command. To provide multiple arguments, use the --git-docker-args flag multiple times.
      --git-docker-command string                Set the docker CMD explicitly. To provide arguments to the command, use the --git-docker-args flag.
      --git-docker-dockerfile string             Dockerfile path
      --git-docker-entrypoint strings            Docker entrypoint
      --git-docker-target string                 Docker target
      --git-no-deploy-on-push                    Disable new deployments creation when code changes are pushed on the configured branch
      --git-run-command string                   Run command (legacy, prefer git-buildpack-run-command)
      --git-sha string                           Git commit SHA to deploy
      --git-workdir string                       Path to the sub-directory containing the code to build and deploy
  -h, --help                                     help for update
      --instance-type string                     Instance type (default "nano")
      --max-scale int                            Max scale (default 1)
      --min-scale int                            Min scale (default 1)
      --name string                              Specify to update the service name
      --override                                 Override the service configuration with the new configuration instead of merging them
      --ports strings                            Update service ports (available for services of type "web" only) using the format PORT[:PROTOCOL], for example --port 8080:http
                                                 PROTOCOL defaults to "http". Supported protocols are "http", "http2" and "tcp"
                                                 To delete an exposed port, prefix its number with '!', for example --port '!80'
                                                 
      --privileged                               Whether the service container should run in privileged mode
      --regions strings                          Add a region where the service is deployed. You can specify this flag multiple times to deploy the service in multiple regions.
                                                 To update a service and remove a region, prefix the region name with '!', for example --region '!par'
                                                 If the region is not specified on service creation, the service is deployed in fra
                                                 
      --routes strings                           Update service routes (available for services of type "web" only) using the format PATH[:PORT], for example '/foo:8080'
                                                 PORT defaults to 8000
                                                 To delete a route, use '!PATH', for example --route '!/foo'
                                                 
      --save-only                                Save the new configuration without deploying it
      --scale int                                Set both min-scale and max-scale (default 1)
      --skip-build                               If there has been at least one past successfully build deployment, use the last one instead of rebuilding. WARNING: this can lead to unexpected behavior if the build depends, for example, on environment variables.
      --skip-cache                               Whether to use the cache when building the service
      --type string                              Service type, either "web" or "worker" (default "web")
```

### Options inherited from parent commands

```
  -c, --config string         config file (default is $HOME/.koyeb.yaml)
  -d, --debug                 enable the debug output
      --debug-full            do not hide sensitive information (tokens) in the debug output
      --force-ascii           only output ascii characters (no unicode emojis)
      --full                  do not truncate output
      --organization string   organization ID
  -o, --output output         output format (yaml,json,table)
      --token string          API token
      --url string            url of the api (default "https://app.koyeb.com")
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
  -c, --config string         config file (default is $HOME/.koyeb.yaml)
  -d, --debug                 enable the debug output
      --debug-full            do not hide sensitive information (tokens) in the debug output
      --force-ascii           only output ascii characters (no unicode emojis)
      --full                  do not truncate output
      --organization string   organization ID
  -o, --output output         output format (yaml,json,table)
      --token string          API token
      --url string            url of the api (default "https://app.koyeb.com")
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
  -c, --config string         config file (default is $HOME/.koyeb.yaml)
  -d, --debug                 enable the debug output
      --debug-full            do not hide sensitive information (tokens) in the debug output
      --force-ascii           only output ascii characters (no unicode emojis)
      --full                  do not truncate output
      --organization string   organization ID
  -o, --output output         output format (yaml,json,table)
      --token string          API token
      --url string            url of the api (default "https://app.koyeb.com")
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
  -c, --config string         config file (default is $HOME/.koyeb.yaml)
  -d, --debug                 enable the debug output
      --debug-full            do not hide sensitive information (tokens) in the debug output
      --force-ascii           only output ascii characters (no unicode emojis)
      --full                  do not truncate output
      --organization string   organization ID
  -o, --output output         output format (yaml,json,table)
      --token string          API token
      --url string            url of the api (default "https://app.koyeb.com")
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
  -c, --config string         config file (default is $HOME/.koyeb.yaml)
  -d, --debug                 enable the debug output
      --debug-full            do not hide sensitive information (tokens) in the debug output
      --force-ascii           only output ascii characters (no unicode emojis)
      --full                  do not truncate output
      --organization string   organization ID
  -o, --output output         output format (yaml,json,table)
      --token string          API token
      --url string            url of the api (default "https://app.koyeb.com")
```



* [koyeb deployments](#koyeb-deployments)	 - Deployments

## koyeb deployments list

List deployments

```
koyeb deployments list [flags]
```

### Options

```
      --app string       Limit the list to deployments of a specific app
  -h, --help             help for list
      --service string   Limit the list to deployments of a specific service
```

### Options inherited from parent commands

```
  -c, --config string         config file (default is $HOME/.koyeb.yaml)
  -d, --debug                 enable the debug output
      --debug-full            do not hide sensitive information (tokens) in the debug output
      --force-ascii           only output ascii characters (no unicode emojis)
      --full                  do not truncate output
      --organization string   organization ID
  -o, --output output         output format (yaml,json,table)
      --token string          API token
      --url string            url of the api (default "https://app.koyeb.com")
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
  -t, --type string   Type of log (runtime, build)
```

### Options inherited from parent commands

```
  -c, --config string         config file (default is $HOME/.koyeb.yaml)
  -d, --debug                 enable the debug output
      --debug-full            do not hide sensitive information (tokens) in the debug output
      --force-ascii           only output ascii characters (no unicode emojis)
      --full                  do not truncate output
      --organization string   organization ID
  -o, --output output         output format (yaml,json,table)
      --token string          API token
      --url string            url of the api (default "https://app.koyeb.com")
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
  -c, --config string         config file (default is $HOME/.koyeb.yaml)
  -d, --debug                 enable the debug output
      --debug-full            do not hide sensitive information (tokens) in the debug output
      --force-ascii           only output ascii characters (no unicode emojis)
      --full                  do not truncate output
      --organization string   organization ID
  -o, --output output         output format (yaml,json,table)
      --token string          API token
      --url string            url of the api (default "https://app.koyeb.com")
```



* [koyeb](#koyeb)	 - Koyeb CLI
* [koyeb instances cp](#koyeb-instances-cp)	 - Copy files and directories to and from instances.
* [koyeb instances describe](#koyeb-instances-describe)	 - Describe instance
* [koyeb instances exec](#koyeb-instances-exec)	 - Run a command in the context of an instance
* [koyeb instances get](#koyeb-instances-get)	 - Get instance
* [koyeb instances list](#koyeb-instances-list)	 - List instances
* [koyeb instances logs](#koyeb-instances-logs)	 - Get instance logs

## koyeb instances cp

Copy files and directories to and from instances.

```
koyeb instances cp SRC DST [flags]
```

### Examples

```

To copy a file called `hello.txt` from the current directory of your local machine to the `/tmp` directory of a remote Koyeb Instance, type:
$> koyeb instance cp hello.txt <instance_id>:/tmp/
To copy a `spreadsheet.csv` file from the `/tmp/` directory of your Koyeb Instance to the current directory on your local machine, type:
$> koyeb instance cp <instance_id>:/tmp/spreadsheet.csv .
```

### Options

```
  -h, --help   help for cp
```

### Options inherited from parent commands

```
  -c, --config string         config file (default is $HOME/.koyeb.yaml)
  -d, --debug                 enable the debug output
      --debug-full            do not hide sensitive information (tokens) in the debug output
      --force-ascii           only output ascii characters (no unicode emojis)
      --full                  do not truncate output
      --organization string   organization ID
  -o, --output output         output format (yaml,json,table)
      --token string          API token
      --url string            url of the api (default "https://app.koyeb.com")
```



* [koyeb instances](#koyeb-instances)	 - Instances

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
  -c, --config string         config file (default is $HOME/.koyeb.yaml)
  -d, --debug                 enable the debug output
      --debug-full            do not hide sensitive information (tokens) in the debug output
      --force-ascii           only output ascii characters (no unicode emojis)
      --full                  do not truncate output
      --organization string   organization ID
  -o, --output output         output format (yaml,json,table)
      --token string          API token
      --url string            url of the api (default "https://app.koyeb.com")
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
  -c, --config string         config file (default is $HOME/.koyeb.yaml)
  -d, --debug                 enable the debug output
      --debug-full            do not hide sensitive information (tokens) in the debug output
      --force-ascii           only output ascii characters (no unicode emojis)
      --full                  do not truncate output
      --organization string   organization ID
  -o, --output output         output format (yaml,json,table)
      --token string          API token
      --url string            url of the api (default "https://app.koyeb.com")
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
  -c, --config string         config file (default is $HOME/.koyeb.yaml)
  -d, --debug                 enable the debug output
      --debug-full            do not hide sensitive information (tokens) in the debug output
      --force-ascii           only output ascii characters (no unicode emojis)
      --full                  do not truncate output
      --organization string   organization ID
  -o, --output output         output format (yaml,json,table)
      --token string          API token
      --url string            url of the api (default "https://app.koyeb.com")
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
  -c, --config string         config file (default is $HOME/.koyeb.yaml)
  -d, --debug                 enable the debug output
      --debug-full            do not hide sensitive information (tokens) in the debug output
      --force-ascii           only output ascii characters (no unicode emojis)
      --full                  do not truncate output
      --organization string   organization ID
  -o, --output output         output format (yaml,json,table)
      --token string          API token
      --url string            url of the api (default "https://app.koyeb.com")
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
  -c, --config string         config file (default is $HOME/.koyeb.yaml)
  -d, --debug                 enable the debug output
      --debug-full            do not hide sensitive information (tokens) in the debug output
      --force-ascii           only output ascii characters (no unicode emojis)
      --full                  do not truncate output
      --organization string   organization ID
  -o, --output output         output format (yaml,json,table)
      --token string          API token
      --url string            url of the api (default "https://app.koyeb.com")
```



* [koyeb instances](#koyeb-instances)	 - Instances

## koyeb databases

Databases

### Options

```
  -h, --help   help for databases
```

### Options inherited from parent commands

```
  -c, --config string         config file (default is $HOME/.koyeb.yaml)
  -d, --debug                 enable the debug output
      --debug-full            do not hide sensitive information (tokens) in the debug output
      --force-ascii           only output ascii characters (no unicode emojis)
      --full                  do not truncate output
      --organization string   organization ID
  -o, --output output         output format (yaml,json,table)
      --token string          API token
      --url string            url of the api (default "https://app.koyeb.com")
```



* [koyeb](#koyeb)	 - Koyeb CLI
* [koyeb databases create](#koyeb-databases-create)	 - Create database
* [koyeb databases delete](#koyeb-databases-delete)	 - Delete database
* [koyeb databases get](#koyeb-databases-get)	 - Get database
* [koyeb databases list](#koyeb-databases-list)	 - List databases
* [koyeb databases update](#koyeb-databases-update)	 - Update database

## koyeb databases create

Create database

```
koyeb databases create NAME [flags]
```

### Options

```
      --db-name string         Database name (default "koyebdb")
      --db-owner string        Database owner (default "koyeb-adm")
  -h, --help                   help for create
      --instance-type string   Instance type (free, small, medium or large) (default "free")
      --pg-version int         PostgreSQL version (default 16)
      --region string          Region where the database is deployed (default "fra")
```

### Options inherited from parent commands

```
  -c, --config string         config file (default is $HOME/.koyeb.yaml)
  -d, --debug                 enable the debug output
      --debug-full            do not hide sensitive information (tokens) in the debug output
      --force-ascii           only output ascii characters (no unicode emojis)
      --full                  do not truncate output
      --organization string   organization ID
  -o, --output output         output format (yaml,json,table)
      --token string          API token
      --url string            url of the api (default "https://app.koyeb.com")
```



* [koyeb databases](#koyeb-databases)	 - Databases

## koyeb databases delete

Delete database

```
koyeb databases delete NAME [flags]
```

### Options

```
  -h, --help   help for delete
```

### Options inherited from parent commands

```
  -c, --config string         config file (default is $HOME/.koyeb.yaml)
  -d, --debug                 enable the debug output
      --debug-full            do not hide sensitive information (tokens) in the debug output
      --force-ascii           only output ascii characters (no unicode emojis)
      --full                  do not truncate output
      --organization string   organization ID
  -o, --output output         output format (yaml,json,table)
      --token string          API token
      --url string            url of the api (default "https://app.koyeb.com")
```



* [koyeb databases](#koyeb-databases)	 - Databases

## koyeb databases get

Get database

```
koyeb databases get NAME [flags]
```

### Options

```
  -h, --help   help for get
```

### Options inherited from parent commands

```
  -c, --config string         config file (default is $HOME/.koyeb.yaml)
  -d, --debug                 enable the debug output
      --debug-full            do not hide sensitive information (tokens) in the debug output
      --force-ascii           only output ascii characters (no unicode emojis)
      --full                  do not truncate output
      --organization string   organization ID
  -o, --output output         output format (yaml,json,table)
      --token string          API token
      --url string            url of the api (default "https://app.koyeb.com")
```



* [koyeb databases](#koyeb-databases)	 - Databases

## koyeb databases list

List databases

```
koyeb databases list [flags]
```

### Options

```
  -h, --help   help for list
```

### Options inherited from parent commands

```
  -c, --config string         config file (default is $HOME/.koyeb.yaml)
  -d, --debug                 enable the debug output
      --debug-full            do not hide sensitive information (tokens) in the debug output
      --force-ascii           only output ascii characters (no unicode emojis)
      --full                  do not truncate output
      --organization string   organization ID
  -o, --output output         output format (yaml,json,table)
      --token string          API token
      --url string            url of the api (default "https://app.koyeb.com")
```



* [koyeb databases](#koyeb-databases)	 - Databases

## koyeb databases update

Update database

```
koyeb databases update NAME [flags]
```

### Options

```
  -h, --help                   help for update
      --instance-type string   Instance type (free, small, medium or large) (default "free")
```

### Options inherited from parent commands

```
  -c, --config string         config file (default is $HOME/.koyeb.yaml)
  -d, --debug                 enable the debug output
      --debug-full            do not hide sensitive information (tokens) in the debug output
      --force-ascii           only output ascii characters (no unicode emojis)
      --full                  do not truncate output
      --organization string   organization ID
  -o, --output output         output format (yaml,json,table)
      --token string          API token
      --url string            url of the api (default "https://app.koyeb.com")
```



* [koyeb databases](#koyeb-databases)	 - Databases

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
  -c, --config string         config file (default is $HOME/.koyeb.yaml)
  -d, --debug                 enable the debug output
      --debug-full            do not hide sensitive information (tokens) in the debug output
      --force-ascii           only output ascii characters (no unicode emojis)
      --full                  do not truncate output
      --organization string   organization ID
  -o, --output output         output format (yaml,json,table)
      --token string          API token
      --url string            url of the api (default "https://app.koyeb.com")
```



* [koyeb](#koyeb)	 - Koyeb CLI

## koyeb volumes

Manage persistent volumes

### Options

```
  -h, --help   help for volumes
```

### Options inherited from parent commands

```
  -c, --config string         config file (default is $HOME/.koyeb.yaml)
  -d, --debug                 enable the debug output
      --debug-full            do not hide sensitive information (tokens) in the debug output
      --force-ascii           only output ascii characters (no unicode emojis)
      --full                  do not truncate output
      --organization string   organization ID
  -o, --output output         output format (yaml,json,table)
      --token string          API token
      --url string            url of the api (default "https://app.koyeb.com")
```



* [koyeb](#koyeb)	 - Koyeb CLI
* [koyeb volumes create](#koyeb-volumes-create)	 - Create a new volume
* [koyeb volumes delete](#koyeb-volumes-delete)	 - Delete a volume
* [koyeb volumes get](#koyeb-volumes-get)	 - Get a volume
* [koyeb volumes list](#koyeb-volumes-list)	 - List volumes
* [koyeb volumes update](#koyeb-volumes-update)	 - Update a volume

