## v3.11.0 (2024-04-15)

* Add warning when accessing logs of a deployment with skipped build.
  - https://github.com/koyeb/koyeb-cli/pull/201
* Add instance ID to `koyeb service logs`, `koyeb deployment logs` and `koyeb instance logs`
  - https://github.com/koyeb/koyeb-cli/pull/202
* Add the `--save-only` flag to the `koyeb services update` command, that only saves the changes and does not trigger an immediate deploy.
  - https://github.com/koyeb/koyeb-cli/pull/200
* Support `-o json` for `koyeb service logs`, `koyeb deployment logs` and `koyeb instance logs`
  - https://github.com/koyeb/koyeb-cli/issues/203
* Add command `koyeb metrics get` to get the metrics of a service or an instance
  - https://github.com/koyeb/koyeb-cli/issues/162
* It is now possible to create services only available from the mesh, for example with `koyeb service create app/service --docker nginx --port 80:tcp`. With only `--port 80` (equivalent to `--port 80:http`), the service is also exposed to the internet.
  - https://github.com/koyeb/koyeb-cli/issues/207

## v3.10.0 (2024-03-28)

* Add the `--skip-build` flag for the `koyeb services redeploy` and `koyeb services update` commands.
* Add `--app` and `--service` flags to `koyeb deployments list` to filter deployments by app and service.
  - https://github.com/koyeb/koyeb-cli/issues/197
* Add `--deployment` to `koyeb regional-deployments list` to filter regional deployments by deployment ID.
  - https://github.com/koyeb/koyeb-cli/issues/198

## v3.9.0 (2024-03-14)

* Display date in `koyeb service logs` and `koyeb instance logs`
  - https://github.com/koyeb/koyeb-cli/issues/190
* Display the log stream in `koyeb service logs` and `koyeb instance logs`
  - https://github.com/koyeb/koyeb-cli/issues/192

## v3.8.1

* Better help message for `koyeb cp`
* The version 3.8.0 had never been released because the github action has not been triggered. This version is a re-release of the version 3.8.0.

## v3.8.0

* `koyeb service update`: remove autoscaling targets when --min-scale is equal to --max-scale
  - https://github.com/koyeb/koyeb-cli/issues/182
* `koyeb deployment get` and `koyeb deployment describe` now display the GIT commit hash for git services types
  - https://github.com/koyeb/koyeb-cli/pull/184
* Check the validity of the docker image when creating or updating a service
  - https://github.com/koyeb/koyeb-cli/pull/185

## v3.7.1

* Add `koyeb db update <name> --instance-type <type>` to update the instance type of a database
  - https://github.com/koyeb/koyeb-cli/issues/180

## v3.7.0

* Add `koyeb regional-deployments list` and `koyeb regional-deployments get`. Also works with the aliases `rd`, `rdeployment` and `rdeployments`.
  - https://github.com/koyeb/koyeb-cli/issues/176
* Stop hardcoding the maximum usage time and the database size displayed by `koyeb db list` and `koyeb db get`
  - https://github.com/koyeb/koyeb-cli/issues/169

## v3.6.1

* Always fetch the latest git commit with `koyeb service update`. It should fix the issue where an old commit is deployed instead of the latest one.
  - https://github.com/koyeb/koyeb-cli/pull/175

## v3.6.0

* `koyeb service create` and `koyeb service update` accept the parameters `--autoscaling-average-cpu`, `--autoscaling-average-mem` and `--autoscaling-requests-per-second` to set the autoscaling policy.
  - https://github.com/koyeb/koyeb-cli/issues/170
* Add the option `--skip-cache` to `koyeb service update`
  - https://github.com/koyeb/koyeb-cli/issues/172

## v3.5.2

* Fix `koyeb service update --git-build-command` and `koyeb service update --git-run-command` when the service has already a build command or a run command configured.
  - https://github.com/koyeb/koyeb-cli/issues/168

## v3.5.1

* Fix build. See v3.5.0 for the other changes.

## v3.5.0

* Fix nil pointer dereference when `--url` is invalid
  - https://github.com/koyeb/koyeb-cli/issues/155
* Fix error when the token is invalid and the user tries to switch organization
  - https://github.com/koyeb/koyeb-cli/issues/154
* Stop rendering partial connection string during the provisioning of a database
  - https://github.com/koyeb/koyeb-cli/issues/159
* Allow to manage registry secrets with `koyeb service create --type registry-<type>` and `koyeb service update`
  - https://github.com/koyeb/koyeb-cli/issues/157
* Add `koyeb instance cp` to copy files from and to an instance, for example with `koyeb instance cp file.txt <instance_id>:/tmp/` or `koyeb instance cp <instance_id>:/tmp/file.txt .`
  - https://github.com/koyeb/koyeb-cli/pull/161

This version has never been released and is replaced with v3.5.1.

## v3.4.0

* Accept the two syntaxes `--app xxx` and `<app>/<service_name>` for koyeb service commands
  - https://github.com/koyeb/koyeb-cli/issues/121
* Update the user agent from 'OpenAPI-Generator/1.0.0/go' to 'koyeb-cli/version'
  - https://github.com/koyeb/koyeb-cli/pull/149
* The new flag `--override` of `koyeb service update` allows to override the service configuration instead of merging the other options provided with the existing configuration.
  - https://github.com/koyeb/koyeb-cli/issues/147
* Support `koyeb secret reveal <secret_name>` (or `koyeb secret show <secret_name>`). *The API behind this command is not stable yet and may change in the future. This command can break at any time and will require an update of the CLI.*
  - https://github.com/koyeb/koyeb-cli/issues/150
* Fix an issue where running `koyeb organizations switch` without specifying an organization or specifying more than one organization would crash.
  - https://github.com/koyeb/koyeb-cli/issues/151
* The default --port and --route for `koyeb service create` is now the port 8000 to match the default of https://app.koyeb.com
  - https://github.com/koyeb/koyeb-cli/issues/152
* `koyeb service create`: set default branch to `main`
  - https://github.com/koyeb/koyeb-cli/issues/153
* Implement `koyeb database list`, `koyeb database create`, `koyeb database get` and `koyeb database delete`
  - https://github.com/koyeb/koyeb-cli/issues/144


## v3.3.2

* Dynamically set `--port` and `--route`. Now, `koyeb service create xxx --app yyy --port 8000` automatically creates the route `/:8000`. Similarly, `koyeb service create xxx --app yyy --route /:9999` automatically creates the port `9999:http`. If `--port` and `--route` are both omitted, as before, the default port `80:http` and route `/:80` are created.
  - https://github.com/koyeb/koyeb-cli/issues/101

## v3.3.1

* Fix a bug where `koyeb services exec` / `koyeb instances exec` would not work for users being members of multiple organizations
  - https://github.com/koyeb/koyeb-cli/pull/146

## v3.3.0

* Add the flag `--privileged` to `koyeb service create` and `koyeb service update`
  - https://github.com/koyeb/koyeb-cli/issues/137

## v3.2.0

* The option `--min-scale` and `--max-scale` of `koyeb service update` where ignored, and the service was always scaled to 1. This is now fixed. Also, add the option `--scale` which sets both `--min-scale` and `--max-scale` to the same value. If `--min-scale` or `--max-scale` is set, it overrides `--scale`.
  - https://github.com/koyeb/koyeb-cli/issues/122
  - https://github.com/koyeb/koyeb-cli/pull/124
* Add completion for fish shell, thanks to @razvanazamfirei
  - https://github.com/koyeb/koyeb-cli/pull/118
* The option `--git-builder` of `koyeb service update` was ignored when we wanted to update the git builder. This is now fixed.
  - https://github.com/koyeb/koyeb-cli/issues/119
* Add support for organizations. Use command `koyeb organization list` to list the organizations you are a member of, and `koyeb organization use` to switch to another organization.
  - https://github.com/koyeb/koyeb-cli/pull/113
* Fix `koyeb service logs` when using the non-default organization
  - https://github.com/koyeb/koyeb-cli/issues/120
* Fix `--regions` options of `koyeb service create` and `koyeb service update`. To deploy an existing service to a new region, use `koyeb service update <app>/<service> --region <region>`. To remove a region from a service, use `koyeb service update <app>/<service> --region '!<region>'`. On service creation, the service is still deployed to `fra` if no region is specified. If more than two regions are specified, a warning message is displayed to avoid billing surprises.
  - https://github.com/koyeb/koyeb-cli/issues/123

## Prior v3.2.0

Unfortunately, we didn't keep a changelog before v3.0.2. We will try to do better in the future.
