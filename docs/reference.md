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

* [koyeb compose](#koyeb-compose)	 - Create Koyeb resources from a koyeb-compose.yaml file
* [koyeb databases](#koyeb-databases)	 - Databases
* [koyeb deploy](#koyeb-deploy)	 - Deploy a directory to Koyeb
* [koyeb deployments](#koyeb-deployments)	 - Deployments
* [koyeb domains](#koyeb-domains)	 - Domains
* [koyeb instances](#koyeb-instances)	 - Instances
* [koyeb login](#koyeb-login)	 - Login to your Koyeb account
* [koyeb metrics](#koyeb-metrics)	 - Metrics
* [koyeb organizations](#koyeb-organizations)	 - Organization
* [koyeb regional-deployments](#koyeb-regional-deployments)	 - Regional deployments
* [koyeb sandbox](#koyeb-sandbox)	 - Sandbox - interactive execution environments
* [koyeb secrets](#koyeb-secrets)	 - Secrets
* [koyeb services](#koyeb-services)	 - Services
* [koyeb snapshots](#koyeb-snapshots)	 - Manage snapshots
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
      --delete-when-empty   Automatically delete the app after the last service is deleted. Empty apps created without services are not deleted.
  -h, --help                help for create
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
      --archive-ignore-dir strings               Set directories to ignore when building the archive.
                                                 To ignore multiple directories, use the flag multiple times.
                                                 To include all directories, set the flag to an empty string. (default [.git,node_modules,vendor])
      --auth strings                             Add security policies to all routes. Use --auth USERNAME:PASSWORD for basic auth, or --auth API_KEY for API key auth.
                                                 You can reference secrets for passwords and API keys using the syntax {{secret.SECRET_NAME}},
                                                 e.g. --auth 'admin:{{secret.my_pass}}' or --auth '{{secret.my_api_key}}'.
                                                 The referenced secrets must exist before deployment, otherwise the deployment will fail.
                                                 Can be specified multiple times to add multiple credentials.
                                                 
      --auth-disable                             Remove all security policies from routes
      --autoscaling-average-cpu int              Target CPU usage (in %) to trigger a scaling event. Set to 0 to disable CPU autoscaling.
      --autoscaling-average-mem int              Target memory usage (in %) to trigger a scaling event. Set to 0 to disable memory autoscaling.
      --autoscaling-concurrent-requests int      Target concurrent requests to trigger a scaling event. Set to 0 to disable concurrent requests autoscaling.
      --autoscaling-requests-per-second int      Target requests per second to trigger a scaling event. Set to 0 to disable requests per second autoscaling.
      --autoscaling-requests-response-time int   Target p95 response time to trigger a scaling event (in ms). Set to 0 to disable concurrent response time autoscaling.
      --checks strings                           Update service healthchecks (available for services of type "web" only)
                                                 For HTTP healthchecks, use the format <PORT>:http:<PATH>, for example --checks 8080:http:/health
                                                 For TCP healthchecks, use the format <PORT>:tcp, for example --checks 8080:tcp
                                                 To delete a healthcheck, use !PORT, for example --checks '!8080'
                                                 
      --checks-grace-period strings              Set healthcheck grace period in seconds.
                                                 Use the format <healthcheck>=<seconds>, for example --checks-grace-period 8080=10
                                                 
      --config-file strings                      Copy a local file to your service container using the format LOCAL_FILE:PATH:[PERMISSIONS]
                                                 for example --config-file /etc/data.yaml:/etc/data.yaml:0644
                                                 To delete a config file, use !PATH, for example --config-file !/etc/data.yaml
                                                 
      --deep-sleep-delay duration                Delay after which an idle service is put to deep sleep. Use duration format (e.g., '5m', '30m', '1h'). Set to 0 to disable.
      --delete-after-delay duration              Automatically delete the service after this duration from creation. Use duration format (e.g., '1h', '30m', '24h'). Set to 0 to disable.
      --delete-after-inactivity-delay duration   Automatically delete the service after being inactive (sleeping) for this duration. Use duration format (e.g., '1h', '30m', '24h'). Set to 0 to disable.
      --deployment-strategy STRATEGY             Deployment strategy, either "rolling" (default), "blue-green" or "immediate".
      --docker string                            Docker image
      --docker-args strings                      Set arguments to the docker command. To provide multiple arguments, use the --docker-args flag multiple times.
      --docker-command string                    Set the docker CMD explicitly. To provide arguments to the command, use the --docker-args flag.
      --docker-entrypoint strings                Docker entrypoint. To provide multiple arguments, use the --docker-entrypoint flag multiple times.
      --docker-private-registry-secret string    Docker private registry secret
      --env strings                              Update service environment variables using the format KEY=VALUE, for example --env FOO=bar
                                                 To use the value of a secret as an environment variable, use the following syntax: --env FOO={{secret.bar}}
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
      --light-sleep-delay duration               Delay after which an idle service is put to light sleep. Use duration format (e.g., '1m', '5m', '1h'). Set to 0 to disable.
      --max-scale int                            Max scale (default 1)
      --min-scale int                            Min scale (default 1)
      --ports strings                            Update service ports (available for services of type "web" only) using the format PORT[:PROTOCOL], for example --port 8080:http
                                                 PROTOCOL defaults to "http". Supported protocols are "http", "http2" and "tcp"
                                                 To delete an exposed port, prefix its number with '!', for example --port '!80'
                                                 
      --privileged                               Whether the service container should run in privileged mode
      --proxy-ports strings                      Update service proxy ports (available for services of type "web" only) using format PORT[:PROTOCOL], for example --proxy-ports 22:tcp
                                                 PROTOCOL defaults to "tcp". Supported protocols are "tcp".To delete a proxy port, prefix its number with '!', for example --proxy-ports '!80'
                                                 
      --regions strings                          Add a region where the service is deployed. You can specify this flag multiple times to deploy the service in multiple regions.
                                                 To update a service and remove a region, prefix the region name with '!', for example --region '!par'
                                                 If the region is not specified on service creation, the service is deployed in was
                                                 
      --routes strings                           Update service routes (available for services of type "web" only) using the format PATH[:PORT], for example '/foo:8080'
                                                 PORT defaults to 8000
                                                 To delete a route, use '!PATH', for example --route '!/foo'
                                                 
      --scale int                                Set both min-scale and max-scale (default 1)
      --skip-cache                               Whether to use the cache when building the service
      --type string                              Service type, one of "web", "worker" or "sandbox" (default "web")
      --volumes strings                          Update service volumes using the format VOLUME:PATH, for example --volume myvolume:/data.To delete a volume, use !VOLUME, for example --volume '!myvolume'
                                                 
      --wait                                     Waits until app deployment is done
      --wait-timeout duration                    Duration the wait will last until timeout (default 5m0s)
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
      --delete-when-empty   Automatically delete the app after the last service is deleted. Empty apps created without services are not deleted.
  -D, --domain string       Change the subdomain of the app (only specify the subdomain, skipping ".koyeb.app")
  -h, --help                help for update
  -n, --name string         Change the name of the app
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
  -h, --help                 help for create
      --ignore-dir strings   Set directories to ignore when building the archive.
                             To ignore multiple directories, use the flag multiple times.
                             To include all directories, set the flag to an empty string. (default [.git,node_modules,vendor])
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
      --archive-ignore-dir strings               Set directories to ignore when building the archive.
                                                 To ignore multiple directories, use the flag multiple times.
                                                 To include all directories, set the flag to an empty string. (default [.git,node_modules,vendor])
      --auth strings                             Add security policies to all routes. Use --auth USERNAME:PASSWORD for basic auth, or --auth API_KEY for API key auth.
                                                 You can reference secrets for passwords and API keys using the syntax {{secret.SECRET_NAME}},
                                                 e.g. --auth 'admin:{{secret.my_pass}}' or --auth '{{secret.my_api_key}}'.
                                                 The referenced secrets must exist before deployment, otherwise the deployment will fail.
                                                 Can be specified multiple times to add multiple credentials.
                                                 
      --auth-disable                             Remove all security policies from routes
      --autoscaling-average-cpu int              Target CPU usage (in %) to trigger a scaling event. Set to 0 to disable CPU autoscaling.
      --autoscaling-average-mem int              Target memory usage (in %) to trigger a scaling event. Set to 0 to disable memory autoscaling.
      --autoscaling-concurrent-requests int      Target concurrent requests to trigger a scaling event. Set to 0 to disable concurrent requests autoscaling.
      --autoscaling-requests-per-second int      Target requests per second to trigger a scaling event. Set to 0 to disable requests per second autoscaling.
      --autoscaling-requests-response-time int   Target p95 response time to trigger a scaling event (in ms). Set to 0 to disable concurrent response time autoscaling.
      --checks strings                           Update service healthchecks (available for services of type "web" only)
                                                 For HTTP healthchecks, use the format <PORT>:http:<PATH>, for example --checks 8080:http:/health
                                                 For TCP healthchecks, use the format <PORT>:tcp, for example --checks 8080:tcp
                                                 To delete a healthcheck, use !PORT, for example --checks '!8080'
                                                 
      --checks-grace-period strings              Set healthcheck grace period in seconds.
                                                 Use the format <healthcheck>=<seconds>, for example --checks-grace-period 8080=10
                                                 
      --config-file strings                      Copy a local file to your service container using the format LOCAL_FILE:PATH:[PERMISSIONS]
                                                 for example --config-file /etc/data.yaml:/etc/data.yaml:0644
                                                 To delete a config file, use !PATH, for example --config-file !/etc/data.yaml
                                                 
      --deep-sleep-delay duration                Delay after which an idle service is put to deep sleep. Use duration format (e.g., '5m', '30m', '1h'). Set to 0 to disable.
      --delete-after-delay duration              Automatically delete the service after this duration from creation. Use duration format (e.g., '1h', '30m', '24h'). Set to 0 to disable.
      --delete-after-inactivity-delay duration   Automatically delete the service after being inactive (sleeping) for this duration. Use duration format (e.g., '1h', '30m', '24h'). Set to 0 to disable.
      --deployment-strategy STRATEGY             Deployment strategy, either "rolling" (default), "blue-green" or "immediate".
      --env strings                              Update service environment variables using the format KEY=VALUE, for example --env FOO=bar
                                                 To use the value of a secret as an environment variable, use the following syntax: --env FOO={{secret.bar}}
                                                 To delete an environment variable, prefix its name with '!', for example --env '!FOO'
                                                 
  -h, --help                                     help for deploy
      --instance-type string                     Instance type (default "nano")
      --light-sleep-delay duration               Delay after which an idle service is put to light sleep. Use duration format (e.g., '1m', '5m', '1h'). Set to 0 to disable.
      --max-scale int                            Max scale (default 1)
      --min-scale int                            Min scale (default 1)
      --ports strings                            Update service ports (available for services of type "web" only) using the format PORT[:PROTOCOL], for example --port 8080:http
                                                 PROTOCOL defaults to "http". Supported protocols are "http", "http2" and "tcp"
                                                 To delete an exposed port, prefix its number with '!', for example --port '!80'
                                                 
      --privileged                               Whether the service container should run in privileged mode
      --proxy-ports strings                      Update service proxy ports (available for services of type "web" only) using format PORT[:PROTOCOL], for example --proxy-ports 22:tcp
                                                 PROTOCOL defaults to "tcp". Supported protocols are "tcp".To delete a proxy port, prefix its number with '!', for example --proxy-ports '!80'
                                                 
      --regions strings                          Add a region where the service is deployed. You can specify this flag multiple times to deploy the service in multiple regions.
                                                 To update a service and remove a region, prefix the region name with '!', for example --region '!par'
                                                 If the region is not specified on service creation, the service is deployed in was
                                                 
      --routes strings                           Update service routes (available for services of type "web" only) using the format PATH[:PORT], for example '/foo:8080'
                                                 PORT defaults to 8000
                                                 To delete a route, use '!PATH', for example --route '!/foo'
                                                 
      --scale int                                Set both min-scale and max-scale (default 1)
      --skip-cache                               Whether to use the cache when building the service
      --type string                              Service type, one of "web", "worker" or "sandbox" (default "web")
      --volumes strings                          Update service volumes using the format VOLUME:PATH, for example --volume myvolume:/data.To delete a volume, use !VOLUME, for example --volume '!myvolume'
                                                 
      --wait                                     Waits until the deployment is done
      --wait-timeout duration                    Duration the wait will last until timeout (default 5m0s)
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
* [koyeb services scale](#koyeb-services-scale)	 - Set manual scaling configuration for service (replaces existing configuration)
* [koyeb services unapplied-changes](#koyeb-services-unapplied-changes)	 - Show unapplied changes saved with the --save-only flag, which will be applied in the next deployment
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
      --archive-ignore-dir strings               Set directories to ignore when building the archive.
                                                 To ignore multiple directories, use the flag multiple times.
                                                 To include all directories, set the flag to an empty string. (default [.git,node_modules,vendor])
      --auth strings                             Add security policies to all routes. Use --auth USERNAME:PASSWORD for basic auth, or --auth API_KEY for API key auth.
                                                 You can reference secrets for passwords and API keys using the syntax {{secret.SECRET_NAME}},
                                                 e.g. --auth 'admin:{{secret.my_pass}}' or --auth '{{secret.my_api_key}}'.
                                                 The referenced secrets must exist before deployment, otherwise the deployment will fail.
                                                 Can be specified multiple times to add multiple credentials.
                                                 
      --auth-disable                             Remove all security policies from routes
      --autoscaling-average-cpu int              Target CPU usage (in %) to trigger a scaling event. Set to 0 to disable CPU autoscaling.
      --autoscaling-average-mem int              Target memory usage (in %) to trigger a scaling event. Set to 0 to disable memory autoscaling.
      --autoscaling-concurrent-requests int      Target concurrent requests to trigger a scaling event. Set to 0 to disable concurrent requests autoscaling.
      --autoscaling-requests-per-second int      Target requests per second to trigger a scaling event. Set to 0 to disable requests per second autoscaling.
      --autoscaling-requests-response-time int   Target p95 response time to trigger a scaling event (in ms). Set to 0 to disable concurrent response time autoscaling.
      --checks strings                           Update service healthchecks (available for services of type "web" only)
                                                 For HTTP healthchecks, use the format <PORT>:http:<PATH>, for example --checks 8080:http:/health
                                                 For TCP healthchecks, use the format <PORT>:tcp, for example --checks 8080:tcp
                                                 To delete a healthcheck, use !PORT, for example --checks '!8080'
                                                 
      --checks-grace-period strings              Set healthcheck grace period in seconds.
                                                 Use the format <healthcheck>=<seconds>, for example --checks-grace-period 8080=10
                                                 
      --config-file strings                      Copy a local file to your service container using the format LOCAL_FILE:PATH:[PERMISSIONS]
                                                 for example --config-file /etc/data.yaml:/etc/data.yaml:0644
                                                 To delete a config file, use !PATH, for example --config-file !/etc/data.yaml
                                                 
      --deep-sleep-delay duration                Delay after which an idle service is put to deep sleep. Use duration format (e.g., '5m', '30m', '1h'). Set to 0 to disable.
      --delete-after-delay duration              Automatically delete the service after this duration from creation. Use duration format (e.g., '1h', '30m', '24h'). Set to 0 to disable.
      --delete-after-inactivity-delay duration   Automatically delete the service after being inactive (sleeping) for this duration. Use duration format (e.g., '1h', '30m', '24h'). Set to 0 to disable.
      --deployment-strategy STRATEGY             Deployment strategy, either "rolling" (default), "blue-green" or "immediate".
      --docker string                            Docker image
      --docker-args strings                      Set arguments to the docker command. To provide multiple arguments, use the --docker-args flag multiple times.
      --docker-command string                    Set the docker CMD explicitly. To provide arguments to the command, use the --docker-args flag.
      --docker-entrypoint strings                Docker entrypoint. To provide multiple arguments, use the --docker-entrypoint flag multiple times.
      --docker-private-registry-secret string    Docker private registry secret
      --env strings                              Update service environment variables using the format KEY=VALUE, for example --env FOO=bar
                                                 To use the value of a secret as an environment variable, use the following syntax: --env FOO={{secret.bar}}
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
      --light-sleep-delay duration               Delay after which an idle service is put to light sleep. Use duration format (e.g., '1m', '5m', '1h'). Set to 0 to disable.
      --max-scale int                            Max scale (default 1)
      --min-scale int                            Min scale (default 1)
      --ports strings                            Update service ports (available for services of type "web" only) using the format PORT[:PROTOCOL], for example --port 8080:http
                                                 PROTOCOL defaults to "http". Supported protocols are "http", "http2" and "tcp"
                                                 To delete an exposed port, prefix its number with '!', for example --port '!80'
                                                 
      --privileged                               Whether the service container should run in privileged mode
      --proxy-ports strings                      Update service proxy ports (available for services of type "web" only) using format PORT[:PROTOCOL], for example --proxy-ports 22:tcp
                                                 PROTOCOL defaults to "tcp". Supported protocols are "tcp".To delete a proxy port, prefix its number with '!', for example --proxy-ports '!80'
                                                 
      --regions strings                          Add a region where the service is deployed. You can specify this flag multiple times to deploy the service in multiple regions.
                                                 To update a service and remove a region, prefix the region name with '!', for example --region '!par'
                                                 If the region is not specified on service creation, the service is deployed in was
                                                 
      --routes strings                           Update service routes (available for services of type "web" only) using the format PATH[:PORT], for example '/foo:8080'
                                                 PORT defaults to 8000
                                                 To delete a route, use '!PATH', for example --route '!/foo'
                                                 
      --scale int                                Set both min-scale and max-scale (default 1)
      --skip-cache                               Whether to use the cache when building the service
      --type string                              Service type, one of "web", "worker" or "sandbox" (default "web")
      --volumes strings                          Update service volumes using the format VOLUME:PATH, for example --volume myvolume:/data.To delete a volume, use !VOLUME, for example --volume '!myvolume'
                                                 
      --wait                                     Waits until service deployment is done
      --wait-timeout duration                    Duration the wait will last until timeout (default 5m0s)
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
  -a, --app string                Service application
      --end-time string           Return logs before this date
  -h, --help                      help for logs
      --instance string           Instance
      --order asc                 Order logs by asc or `desc` (default "asc")
      --regex-search string       Filter logs returned with this regex
      --since HumanFriendlyDate   DEPRECATED. Use --tail --start-time instead. (default 0001-01-01 00:00:00 +0000 UTC)
      --start-time string         Return logs after this date
      --tail                      Tail logs if no --end-time is provided.
      --text-search string        Filter logs returned with this text
  -t, --type string               Type (runtime, build)
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
  -a, --app string              Service application
  -h, --help                    help for redeploy
      --skip-build              If there has been at least one past successfully build deployment, use the last one instead of rebuilding. WARNING: this can lead to unexpected behavior if the build depends, for example, on environment variables.
      --use-cache               Use cache to redeploy
      --wait                    Waits until service deployment is done.
      --wait-timeout duration   Duration the wait will last until timeout (default 5m0s)
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

## koyeb services unapplied-changes

Show unapplied changes saved with the --save-only flag, which will be applied in the next deployment

```
koyeb services unapplied-changes SERVICE_NAME [flags]
```

### Options

```
  -a, --app string   Service application
  -h, --help         help for unapplied-changes
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
      --archive-ignore-dir strings               Set directories to ignore when building the archive.
                                                 To ignore multiple directories, use the flag multiple times.
                                                 To include all directories, set the flag to an empty string. (default [.git,node_modules,vendor])
      --auth strings                             Add security policies to all routes. Use --auth USERNAME:PASSWORD for basic auth, or --auth API_KEY for API key auth.
                                                 You can reference secrets for passwords and API keys using the syntax {{secret.SECRET_NAME}},
                                                 e.g. --auth 'admin:{{secret.my_pass}}' or --auth '{{secret.my_api_key}}'.
                                                 The referenced secrets must exist before deployment, otherwise the deployment will fail.
                                                 Can be specified multiple times to add multiple credentials.
                                                 
      --auth-disable                             Remove all security policies from routes
      --autoscaling-average-cpu int              Target CPU usage (in %) to trigger a scaling event. Set to 0 to disable CPU autoscaling.
      --autoscaling-average-mem int              Target memory usage (in %) to trigger a scaling event. Set to 0 to disable memory autoscaling.
      --autoscaling-concurrent-requests int      Target concurrent requests to trigger a scaling event. Set to 0 to disable concurrent requests autoscaling.
      --autoscaling-requests-per-second int      Target requests per second to trigger a scaling event. Set to 0 to disable requests per second autoscaling.
      --autoscaling-requests-response-time int   Target p95 response time to trigger a scaling event (in ms). Set to 0 to disable concurrent response time autoscaling.
      --checks strings                           Update service healthchecks (available for services of type "web" only)
                                                 For HTTP healthchecks, use the format <PORT>:http:<PATH>, for example --checks 8080:http:/health
                                                 For TCP healthchecks, use the format <PORT>:tcp, for example --checks 8080:tcp
                                                 To delete a healthcheck, use !PORT, for example --checks '!8080'
                                                 
      --checks-grace-period strings              Set healthcheck grace period in seconds.
                                                 Use the format <healthcheck>=<seconds>, for example --checks-grace-period 8080=10
                                                 
      --config-file strings                      Copy a local file to your service container using the format LOCAL_FILE:PATH:[PERMISSIONS]
                                                 for example --config-file /etc/data.yaml:/etc/data.yaml:0644
                                                 To delete a config file, use !PATH, for example --config-file !/etc/data.yaml
                                                 
      --deep-sleep-delay duration                Delay after which an idle service is put to deep sleep. Use duration format (e.g., '5m', '30m', '1h'). Set to 0 to disable.
      --delete-after-delay duration              Automatically delete the service after this duration from creation. Use duration format (e.g., '1h', '30m', '24h'). Set to 0 to disable.
      --delete-after-inactivity-delay duration   Automatically delete the service after being inactive (sleeping) for this duration. Use duration format (e.g., '1h', '30m', '24h'). Set to 0 to disable.
      --deployment-strategy STRATEGY             Deployment strategy, either "rolling" (default), "blue-green" or "immediate".
      --docker string                            Docker image
      --docker-args strings                      Set arguments to the docker command. To provide multiple arguments, use the --docker-args flag multiple times.
      --docker-command string                    Set the docker CMD explicitly. To provide arguments to the command, use the --docker-args flag.
      --docker-entrypoint strings                Docker entrypoint. To provide multiple arguments, use the --docker-entrypoint flag multiple times.
      --docker-private-registry-secret string    Docker private registry secret
      --env strings                              Update service environment variables using the format KEY=VALUE, for example --env FOO=bar
                                                 To use the value of a secret as an environment variable, use the following syntax: --env FOO={{secret.bar}}
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
      --light-sleep-delay duration               Delay after which an idle service is put to light sleep. Use duration format (e.g., '1m', '5m', '1h'). Set to 0 to disable.
      --max-scale int                            Max scale (default 1)
      --min-scale int                            Min scale (default 1)
      --name string                              Specify to update the service name
      --override                                 Override the service configuration with the new configuration instead of merging them
      --ports strings                            Update service ports (available for services of type "web" only) using the format PORT[:PROTOCOL], for example --port 8080:http
                                                 PROTOCOL defaults to "http". Supported protocols are "http", "http2" and "tcp"
                                                 To delete an exposed port, prefix its number with '!', for example --port '!80'
                                                 
      --privileged                               Whether the service container should run in privileged mode
      --proxy-ports strings                      Update service proxy ports (available for services of type "web" only) using format PORT[:PROTOCOL], for example --proxy-ports 22:tcp
                                                 PROTOCOL defaults to "tcp". Supported protocols are "tcp".To delete a proxy port, prefix its number with '!', for example --proxy-ports '!80'
                                                 
      --regions strings                          Add a region where the service is deployed. You can specify this flag multiple times to deploy the service in multiple regions.
                                                 To update a service and remove a region, prefix the region name with '!', for example --region '!par'
                                                 If the region is not specified on service creation, the service is deployed in was
                                                 
      --routes strings                           Update service routes (available for services of type "web" only) using the format PATH[:PORT], for example '/foo:8080'
                                                 PORT defaults to 8000
                                                 To delete a route, use '!PATH', for example --route '!/foo'
                                                 
      --save-only                                Save the new configuration without deploying it
      --scale int                                Set both min-scale and max-scale (default 1)
      --skip-build                               If there has been at least one past successfully build deployment, use the last one instead of rebuilding. WARNING: this can lead to unexpected behavior if the build depends, for example, on environment variables.
      --skip-cache                               Whether to use the cache when building the service
      --type string                              Service type, one of "web", "worker" or "sandbox" (default "web")
      --volumes strings                          Update service volumes using the format VOLUME:PATH, for example --volume myvolume:/data.To delete a volume, use !VOLUME, for example --volume '!myvolume'
                                                 
      --wait                                     Waits until the service deployment is done
      --wait-timeout duration                    Duration the wait will last until timeout (default 5m0s)
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

## koyeb services scale

Set manual scaling configuration for service (replaces existing configuration)

```
koyeb services scale NAME [flags]
```

### Examples

```

# Scale a service to 3 instances across all regions
$> koyeb service scale app/podinfo --instances 3

# Scale a service with different instance counts per region
$> koyeb service scale app/podinfo --scale fra:3 --scale was:2

# Scale a service in specific regions with same instance count (legacy syntax)
$> koyeb service scale app/podinfo --instances 2 --regions fra --regions was

# Set specific scaling per region
$> koyeb service scale app/podinfo --scale fra:5 --scale was:3 --scale sin:2

```

### Options

```
  -a, --app string        Service application
  -h, --help              help for scale
      --instances int     Number of instances to scale to (used with --regions or alone for all regions) (default 1)
      --regions strings   Regions to apply --instances count to (e.g., 'fra', 'was')
      --scale strings     Scale configuration per region in format 'region:instances' (e.g., 'fra:3'). Can be specified multiple times.
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
* [koyeb services scale delete](#koyeb-services-scale-delete)	 - Delete manual scaling configuration for service
* [koyeb services scale get](#koyeb-services-scale-get)	 - Get manual scaling configuration for service
* [koyeb services scale update](#koyeb-services-scale-update)	 - Update manual scaling configuration for service (patches existing configuration)

## koyeb services scale delete

Delete manual scaling configuration for service

```
koyeb services scale delete NAME [flags]
```

### Examples

```

# Remove all manual scaling configuration from a service
$> koyeb service scale delete app/podinfo

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



* [koyeb services scale](#koyeb-services-scale)	 - Set manual scaling configuration for service (replaces existing configuration)

## koyeb services scale get

Get manual scaling configuration for service

```
koyeb services scale get NAME [flags]
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



* [koyeb services scale](#koyeb-services-scale)	 - Set manual scaling configuration for service (replaces existing configuration)

## koyeb services scale update

Update manual scaling configuration for service (patches existing configuration)

```
koyeb services scale update NAME [flags]
```

### Examples

```

# Update instance count for specific regions, keeping other regions unchanged
$> koyeb service scale update app/podinfo --scale fra:5

# Update multiple regions
$> koyeb service scale update app/podinfo --scale fra:3 --scale was:2

# Remove scaling configuration for a specific region
$> koyeb service scale update app/podinfo --scale '!fra'

# Remove global scaling configuration
$> koyeb service scale update app/podinfo --scale '!'

# Remove scaling using --regions flag
$> koyeb service scale update app/podinfo --regions '!par'

# Update some regions and remove others
$> koyeb service scale update app/podinfo --scale fra:5 --scale '!was'

```

### Options

```
  -a, --app string        Service application
  -h, --help              help for update
      --instances int     Number of instances to scale to (used with --regions or alone for all regions) (default 1)
      --regions strings   Regions to apply --instances count to (e.g., 'fra', 'was') or '!region' to remove (e.g., '!fra')
      --scale strings     Scale configuration per region in format 'region:instances' (e.g., 'fra:3') or '!region' to remove (e.g., '!fra'). Can be specified multiple times.
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



* [koyeb services scale](#koyeb-services-scale)	 - Set manual scaling configuration for service (replaces existing configuration)

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
  -e, --end-time string           Return logs before this date
  -h, --help                      help for logs
      --order asc                 Order logs by asc or `desc` (default "asc")
      --regex-search string       Filter logs returned with this regex
      --since HumanFriendlyDate   DEPRECATED. Use --tail --start-time instead. (default 0001-01-01 00:00:00 +0000 UTC)
  -s, --start-time string         Return logs after this date
      --tail                      Tail logs if no --end-time is provided.
      --text-search string        Filter logs returned with this text
  -t, --type string               Type of log (runtime, build)
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
  -e, --end-time string           Return logs before this date
  -h, --help                      help for logs
      --order asc                 Order logs by asc or `desc` (default "asc")
      --regex-search string       Filter logs returned with this regex
      --since HumanFriendlyDate   DEPRECATED. Use --tail --start-time instead. (default 0001-01-01 00:00:00 +0000 UTC)
  -s, --start-time string         Return logs after this date
      --tail                      Tail logs if no --end-time is provided.
      --text-search string        Filter logs returned with this text
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
      --app app-name/database-name   Database application. If the application does not exist, it will be created. Can also be provided in the database name with the format app-name/database-name
      --db-name string               Database name (default "koyebdb")
      --db-owner string              Database owner (default "koyeb-adm")
  -h, --help                         help for create
      --instance-type string         Instance type (free, small, medium or large) (default "free")
      --name string                  Specify to update the database name
      --pg-version int               PostgreSQL version (default 16)
      --region string                Region where the database is deployed (default "was")
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
      --app app-name/database-name   Database application. If the application does not exist, it will be created. Can also be provided in the database name with the format app-name/database-name
  -h, --help                         help for get
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
      --app app-name/database-name   Database application. If the application does not exist, it will be created. Can also be provided in the database name with the format app-name/database-name
  -h, --help                         help for update
      --instance-type string         Instance type (free, small, medium or large) (default "free")
      --name string                  Specify to update the database name
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

## koyeb sandbox

Sandbox - interactive execution environments

### Synopsis

Sandbox commands for interacting with sandbox services.

Sandboxes are created using 'koyeb service create --type=sandbox'.
These commands provide additional functionality for running commands,
managing processes, filesystem operations, and port exposure.

### Options

```
  -h, --help   help for sandbox
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
* [koyeb sandbox create](#koyeb-sandbox-create)	 - Create a new sandbox
* [koyeb sandbox expose-port](#koyeb-sandbox-expose-port)	 - Expose a port from the sandbox via TCP proxy
* [koyeb sandbox fs](#koyeb-sandbox-fs)	 - Filesystem operations
* [koyeb sandbox health](#koyeb-sandbox-health)	 - Check sandbox health status
* [koyeb sandbox kill](#koyeb-sandbox-kill)	 - Kill a background process in the sandbox
* [koyeb sandbox list](#koyeb-sandbox-list)	 - List sandboxes
* [koyeb sandbox logs](#koyeb-sandbox-logs)	 - Stream logs from a background process
* [koyeb sandbox ps](#koyeb-sandbox-ps)	 - List background processes in the sandbox
* [koyeb sandbox run](#koyeb-sandbox-run)	 - Execute a command in the sandbox
* [koyeb sandbox start](#koyeb-sandbox-start)	 - Start a background process in the sandbox
* [koyeb sandbox unexpose-port](#koyeb-sandbox-unexpose-port)	 - Unexpose the currently exposed port

## koyeb sandbox create

Create a new sandbox

```
koyeb sandbox create NAME [flags]
```

### Examples

```

# Create a sandbox in an app
$> koyeb sandbox create myapp/mysandbox

# Create with a custom docker image
$> koyeb sandbox create myapp/mysandbox --docker myregistry/myimage

# Create with custom secret
$> koyeb sandbox create myapp/mysandbox --env SANDBOX_SECRET=mysecret

# Create and wait for deployment
$> koyeb sandbox create myapp/mysandbox --wait

```

### Options

```
  -a, --app string                               Sandbox application
      --config-file strings                      Config files (LOCAL:REMOTE:PERMS)
      --deep-sleep-delay duration                Delay after which an idle service is put to deep sleep. Use duration format (e.g., '5m', '30m', '1h'). Set to 0 to disable.
      --delete-after-delay duration              Auto-delete after duration (e.g., '24h')
      --delete-after-inactivity-delay duration   Auto-delete after inactivity (e.g., '1h')
      --docker string                            Docker image (default: koyeb/sandbox)
      --docker-args strings                      Docker command arguments
      --docker-command string                    Docker command
      --docker-entrypoint strings                Docker entrypoint
      --docker-private-registry-secret string    Docker private registry secret
      --env strings                              Environment variables (KEY=VALUE)
  -h, --help                                     help for create
      --instance-type string                     Instance type (default "nano")
      --light-sleep-delay duration               Delay after which an idle service is put to light sleep. Use duration format (e.g., '1m', '5m', '1h'). Set to 0 to disable.
      --min-scale int                            Min scale (default 1)
      --privileged                               Run in privileged mode
      --regions strings                          Deployment regions
      --wait                                     Wait until sandbox deployment is done
      --wait-timeout duration                    Wait timeout duration (default 5m0s)
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



* [koyeb sandbox](#koyeb-sandbox)	 - Sandbox - interactive execution environments

## koyeb sandbox expose-port

Expose a port from the sandbox via TCP proxy

```
koyeb sandbox expose-port NAME PORT [flags]
```

### Examples

```

# Expose port 8080
$> koyeb sandbox expose-port myapp/mysandbox 8080

```

### Options

```
  -h, --help   help for expose-port
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



* [koyeb sandbox](#koyeb-sandbox)	 - Sandbox - interactive execution environments

## koyeb sandbox fs

Filesystem operations

### Options

```
  -h, --help   help for fs
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



* [koyeb sandbox](#koyeb-sandbox)	 - Sandbox - interactive execution environments
* [koyeb sandbox fs download](#koyeb-sandbox-fs-download)	 - Download a file from the sandbox
* [koyeb sandbox fs ls](#koyeb-sandbox-fs-ls)	 - List directory contents in the sandbox
* [koyeb sandbox fs mkdir](#koyeb-sandbox-fs-mkdir)	 - Create a directory in the sandbox
* [koyeb sandbox fs read](#koyeb-sandbox-fs-read)	 - Read a file from the sandbox
* [koyeb sandbox fs rm](#koyeb-sandbox-fs-rm)	 - Remove a file or directory from the sandbox
* [koyeb sandbox fs upload](#koyeb-sandbox-fs-upload)	 - Upload a local file or directory to the sandbox (max 1G per file)
* [koyeb sandbox fs write](#koyeb-sandbox-fs-write)	 - Write content to a file in the sandbox

## koyeb sandbox fs download

Download a file from the sandbox

```
koyeb sandbox fs download NAME REMOTE_PATH LOCAL_PATH [flags]
```

### Options

```
  -h, --help   help for download
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



* [koyeb sandbox fs](#koyeb-sandbox-fs)	 - Filesystem operations

## koyeb sandbox fs ls

List directory contents in the sandbox

```
koyeb sandbox fs ls NAME [PATH] [flags]
```

### Options

```
  -h, --help   help for ls
  -l, --long   Use long listing format with details
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



* [koyeb sandbox fs](#koyeb-sandbox-fs)	 - Filesystem operations

## koyeb sandbox fs mkdir

Create a directory in the sandbox

```
koyeb sandbox fs mkdir NAME PATH [flags]
```

### Options

```
  -h, --help   help for mkdir
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



* [koyeb sandbox fs](#koyeb-sandbox-fs)	 - Filesystem operations

## koyeb sandbox fs read

Read a file from the sandbox

```
koyeb sandbox fs read NAME PATH [flags]
```

### Options

```
  -h, --help   help for read
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



* [koyeb sandbox fs](#koyeb-sandbox-fs)	 - Filesystem operations

## koyeb sandbox fs rm

Remove a file or directory from the sandbox

```
koyeb sandbox fs rm NAME PATH [flags]
```

### Options

```
  -h, --help        help for rm
  -r, --recursive   Remove directories recursively
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



* [koyeb sandbox fs](#koyeb-sandbox-fs)	 - Filesystem operations

## koyeb sandbox fs upload

Upload a local file or directory to the sandbox (max 1G per file)

```
koyeb sandbox fs upload NAME LOCAL_PATH REMOTE_PATH [flags]
```

### Options

```
  -f, --force       Overwrite existing remote directory
  -h, --help        help for upload
  -r, --recursive   Upload directories recursively
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



* [koyeb sandbox fs](#koyeb-sandbox-fs)	 - Filesystem operations

## koyeb sandbox fs write

Write content to a file in the sandbox

### Synopsis

Write content to a file in the sandbox.
Content can be provided as an argument or via stdin with -f flag.

```
koyeb sandbox fs write NAME PATH [CONTENT] [flags]
```

### Examples

```

# Write inline content
$> koyeb sandbox fs write myapp/mysandbox /tmp/hello.txt "Hello World"

# Write from local file
$> koyeb sandbox fs write myapp/mysandbox /tmp/script.py -f ./local-script.py

```

### Options

```
  -f, --file string   Read content from local file
  -h, --help          help for write
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



* [koyeb sandbox fs](#koyeb-sandbox-fs)	 - Filesystem operations

## koyeb sandbox health

Check sandbox health status

```
koyeb sandbox health NAME [flags]
```

### Options

```
  -h, --help   help for health
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



* [koyeb sandbox](#koyeb-sandbox)	 - Sandbox - interactive execution environments

## koyeb sandbox kill

Kill a background process in the sandbox

```
koyeb sandbox kill NAME PROCESS_ID [flags]
```

### Options

```
  -h, --help   help for kill
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



* [koyeb sandbox](#koyeb-sandbox)	 - Sandbox - interactive execution environments

## koyeb sandbox list

List sandboxes

```
koyeb sandbox list [flags]
```

### Options

```
  -a, --app string    App
  -h, --help          help for list
  -n, --name string   Sandbox name
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



* [koyeb sandbox](#koyeb-sandbox)	 - Sandbox - interactive execution environments

## koyeb sandbox logs

Stream logs from a background process

```
koyeb sandbox logs NAME PROCESS_ID [flags]
```

### Options

```
  -f, --follow   Follow log output (like tail -f)
  -h, --help     help for logs
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



* [koyeb sandbox](#koyeb-sandbox)	 - Sandbox - interactive execution environments

## koyeb sandbox ps

List background processes in the sandbox

```
koyeb sandbox ps NAME [flags]
```

### Options

```
  -h, --help   help for ps
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



* [koyeb sandbox](#koyeb-sandbox)	 - Sandbox - interactive execution environments

## koyeb sandbox run

Execute a command in the sandbox

```
koyeb sandbox run NAME COMMAND [ARGS...] [flags]
```

### Examples

```

# Run a simple command
$> koyeb sandbox run myapp/mysandbox echo "Hello World"

# Run a command with arguments
$> koyeb sandbox run myapp/mysandbox ls -la /app

# Run a python script
$> koyeb sandbox run myapp/mysandbox python script.py

# Run with custom working directory
$> koyeb sandbox run myapp/mysandbox --cwd /app python main.py

# Run with streaming output
$> koyeb sandbox run myapp/mysandbox --stream long-running-command

# Run with custom timeout (in seconds)
$> koyeb sandbox run myapp/mysandbox --timeout 120 long-running-command

```

### Options

```
      --cwd string    Working directory for the command
      --env strings   Environment variables (KEY=VALUE)
  -h, --help          help for run
      --stream        Stream output in real-time
      --timeout int   Command timeout in seconds (default 30)
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



* [koyeb sandbox](#koyeb-sandbox)	 - Sandbox - interactive execution environments

## koyeb sandbox start

Start a background process in the sandbox

```
koyeb sandbox start NAME COMMAND [ARGS...] [flags]
```

### Examples

```

# Start a web server in background
$> koyeb sandbox start myapp/mysandbox python -m http.server 8080

# Start a process with custom working directory
$> koyeb sandbox start myapp/mysandbox --cwd /app npm start

```

### Options

```
      --cwd string    Working directory for the process
      --env strings   Environment variables (KEY=VALUE)
  -h, --help          help for start
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



* [koyeb sandbox](#koyeb-sandbox)	 - Sandbox - interactive execution environments

## koyeb sandbox unexpose-port

Unexpose the currently exposed port

```
koyeb sandbox unexpose-port NAME [flags]
```

### Options

```
  -h, --help   help for unexpose-port
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



* [koyeb sandbox](#koyeb-sandbox)	 - Sandbox - interactive execution environments

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

