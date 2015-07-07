package commands

import (
	"fmt"
	"github.com/codegangsta/cli"
	"github.com/fgrehm/devstep-cli/devstep"
	"os"
	"regexp"
	"strings"
)

var dockerRunFlags = []cli.Flag{
	cli.StringFlag{Name: "v, volume", Usage: "Bind mount a volume"},
	cli.StringFlag{Name: "w, working_dir", Usage: "Working directory inside the container"},
	cli.StringSliceFlag{Name: "p, publish", Value: &cli.StringSlice{}, Usage: "Publish a container's port to the host (hostPort:containerPort)"},
	cli.StringSliceFlag{Name: "link", Value: &cli.StringSlice{}, Usage: "Add link to another container (name:alias)"},
	cli.StringSliceFlag{Name: "e, env", Value: &cli.StringSlice{}, Usage: "Set environment variables"},
	cli.BoolFlag{Name: "privileged", Usage: "Give extended privileges to this container"},
}

func bashCompleteRunArgs(c *cli.Context) {
	args := c.Args()
	if len(args) == 0 {
		fmt.Println("-p")
		fmt.Println("--publish")
		fmt.Println("--link")
		fmt.Println("-w")
		fmt.Println("--working_dir")
		fmt.Println("-e")
		fmt.Println("--env")
		fmt.Println("--privileged")
	}
}

func parseRunOpts(c *cli.Context) *devstep.DockerRunOpts {
	// Sane defaults
	runOpts := &devstep.DockerRunOpts{
		Publish: c.StringSlice("publish"),
		Links:   c.StringSlice("link"),
		Name:    c.String("name"),
		Env:     map[string]string{"DEVSTEP_LOG": c.GlobalString("log-level")},
	}

	// Only set the privileged config if it was provided
	if c.IsSet("privileged") {
		privileged := c.Bool("privileged")
		runOpts.Privileged = &privileged
	}

	// Set working dir directly on the project object so that it get passed
	// along properly
	if workingDir := c.String("working_dir"); workingDir != "" {
		project.Config().GuestDir = workingDir
	}

	// Env vars
	for _, envVar := range c.StringSlice("env") {
		varAndValue := strings.Split(envVar, "=")
		runOpts.Env[varAndValue[0]] = varAndValue[1]
	}

	// Validate ports
	validPort := regexp.MustCompile(`^\d+:\d+$`)
	for _, port := range runOpts.Publish {
		if !validPort.MatchString(port) {
			fmt.Println("Invalid publish arg: " + port)
			os.Exit(1)
		}
	}
	// Validate links
	validLink := regexp.MustCompile(`[^:]+:[^:]+`)
	for _, link := range runOpts.Links {
		if !validLink.MatchString(link) {
			fmt.Println("Invalid link: " + link)
			os.Exit(1)
		}
	}
	return runOpts
}
