## [0.4.1](https://github.com/fgrehm/devstep-cli/compare/v0.4.0...v0.4.1) (2015-08-03)

NEW FEATURES:
  - Support for using relative volume paths on on `devstep.yml` [[GH-8]]

[GH-8]: https://github.com/fgrehm/devstep-cli/issues/8

## [0.4.0](https://github.com/fgrehm/devstep-cli/compare/v0.3.1...v0.4.0) (2015-07-06)

NEW FEATURES:
  - New `devstep hack` behavior [[GH-35]]

[GH-35]: https://github.com/fgrehm/devstep-cli/issues/35

## [0.3.1](https://github.com/fgrehm/devstep-cli/compare/v0.3.0...v0.3.1) (2015-03-04)

IMPROVEMENTS:

  - Default source image to [fgrehm/devstep:0.3.1](https://github.com/fgrehm/devstep/releases/tag/v0.3.1)


## [0.3.0](https://github.com/fgrehm/devstep-cli/compare/v0.1.0...v0.3.0) (2015-02-12)

NEW FEATURES:

  - New command: `devstep exec` -> Run `docker exec`s on top of containers that are already running
  - New command: `devstep init` -> Generate an example config file for the current directory

IMPROVEMENTS:

  - Support `devstep run` options for `devstep bootstrap` / `devstep pristine` / `devstep build`
  - Default source image to `fgrehm/devstep:0.3.0`
  - Experimental support for skipping `docker commit`s in case the container filesystem does not get changed during a build
  - Add support for bootstraping from `devstep pristine` command with the `--bootstrap` flag
  - Add support for setting the container name to `devstep hack` and `devstep run` commands (defaults to `project-dir-name-TIMESTAMP`)
  - Add support for setting repository name with `--repository` when running `devstep bootstrap`

BUG FIXES:

  - Proper cleanup in case of errors during container creation on `devstep bootstrap` / `devstep build` / `devstep pristine`
  - Exit code of `devstep run CMD` is now the same as the exit code of running `CMD` inside the container

## 0.1.0 (2014-09-23)

First public release.
