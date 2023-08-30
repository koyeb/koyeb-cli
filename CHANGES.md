(unreleased)

* The option `--min-scale` and `--max-scale` of `koyeb service update` where ignored, and the service was always scaled to 1. This is now fixed. Also, add the option `--scale` which sets both `--min-scale` and `--max-scale` to the same value. If `--min-scale` or `--max-scale` is set, it overrides `--scale`.
  - https://github.com/koyeb/koyeb-cli/issues/122
  - https://github.com/koyeb/koyeb-cli/pull/124
* Add completion for fish shell, thanks to @razvanazamfirei
  - https://github.com/koyeb/koyeb-cli/pull/118
* The option `--git-builder` of `koyeb service update` was ignored when we wanted to update the git builder. This is now fixed.
  - https://github.com/koyeb/koyeb-cli/issues/119

(v3.0.2 and before)

Unfortunatly, we didn't keep a changelog before v3.0.2. We will try to do better in the future.