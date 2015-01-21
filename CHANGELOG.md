## [0.3.0](https://github.com/fgrehm/devstep-cli/compare/v0.1.0...master) (unreleased)

NEW FEATURES:

  - New command: `devstep exec` -> Run `docker exec`s on top of containers that are already running
  - New command: `devstep init` -> Generate an example config file for the current directory
  - Support `devstep run` options for `devstep bootstrap` / `devstep pristine` / `devstep build`

IMPROVEMENTS:

  - Default source image to `fgrehm/devstep:0.3.0`
  - Experimental support for skipping `docker commit`s in case the container filesystem does not get changed during a build
  - Update default source image to `fgrehm/devstep:v0.3.0`
  - Add support for bootstraping from `devstep pristine` command with the `--bootstrap` flag
  - Add support for setting the container name to `devstep hack` and `devstep run` commands (defaults to `project-dir-name:TIMESTAMP`)
  - Add support for setting repository name with `--repository` when running `devstep bootstrap`

BUG FIXES:

  - Proper cleanup in case of errors during container creation on `devstep bootstrap` / `devstep build` / `devstep pristine`
  - Exit code of `devstep run CMD` is now the same as the exit code of running `CMD` inside the container

## 0.1.0 (2014-09-23)

First public release.
