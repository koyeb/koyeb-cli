**v3.3.0**

* Add the flag `--privileged` to `koyeb service create` and `koyeb service update`
  - https://github.com/koyeb/koyeb-cli/issues/137

**v3.2.0**

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

**Prior v3.2.0**

Unfortunately, we didn't keep a changelog before v3.0.2. We will try to do better in the future.
