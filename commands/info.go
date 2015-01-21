package commands

import (
	"fmt"
	"github.com/codegangsta/cli"
	"github.com/fgrehm/devstep-cli/devstep"
)

var InfoCmd = cli.Command{
	Name:  "info",
	Usage: "show information about the current environment",
	Action: func(c *cli.Context) {
		config := loadConfig()
		printConfig(config)
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
	if config.Commands != nil {
		fmt.Println("\n==> Commands:")
		for _, cmd := range config.Commands {
			fmt.Printf("* %s\n", cmd.Name)
			fmt.Printf("  Cmd:        %v\n", cmd.Cmd)
			fmt.Printf("  Publish:    %v\n", cmd.Publish)
			printDockerRunOpts(&cmd.DockerRunOpts, "  ")
		}
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
