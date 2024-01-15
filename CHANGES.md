## unreleased

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
