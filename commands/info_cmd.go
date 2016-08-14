package commands

import (
	"fmt"
	"github.com/urfave/cli"
	"github.com/fgrehm/devstep-cli/devstep"
)

var InfoCmd = cli.Command{
	Name:  "info",
	Usage: "show information about the current environment",
	Action: func(c *cli.Context) error {
		printConfig(project.Config())
        return nil
	},
}

func printConfig(config *devstep.ProjectConfig) {
	fmt.Println("==> Project info")
	fmt.Printf("Repository:   %s\n", config.RepositoryName)
	fmt.Printf("Source image: %s\n", config.SourceImage)
	fmt.Printf("Base image:   %s\n", config.BaseImage)
	fmt.Printf("Host dir:     %s\n", config.HostDir)
	fmt.Printf("Guest dir:    %s\n", config.GuestDir)
	fmt.Printf("Cache dir:    %s\n", config.CacheDir)

	if config.Defaults != nil {
		fmt.Println("\n==> Default options:")
		printDockerRunOpts(config.Defaults, "")
	}

	if config.HackOpts != nil {
		fmt.Println("\n==> Hack options:")
		printDockerRunOpts(config.HackOpts, "")
	}
}

func printDockerRunOpts(opts *devstep.DockerRunOpts, prefix string) {
	privileged := false
	if opts.Privileged != nil {
		privileged = *opts.Privileged
	}
	fmt.Printf("%sPrivileged: %v\n", prefix, privileged)
	fmt.Printf("%sLinks:      %v\n", prefix, opts.Links)
	fmt.Printf("%sVolumes:    %v\n", prefix, opts.Volumes)
	fmt.Printf("%sEnv:        %v\n", prefix, opts.Env)
}
