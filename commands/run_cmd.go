package commands

import (
	"fmt"
	"github.com/urfave/cli"
	"os"
)

var RunCmd = cli.Command{
	Name:  "run",
	Usage: "Run a one off command against the current base image",
	Flags: append(
		[]cli.Flag{cli.StringFlag{Name: "name", Usage: "Name to be assigned to the container"}},
		dockerRunFlags...,
	),
	BashComplete: func(c *cli.Context) {
		bashCompleteRunArgs(c)
		args := c.Args()
		if len(args) == 0 {
			fmt.Println("--name")
		}
	},
	Action: func(c *cli.Context) error {
		project := newProject()

		runOpts := parseRunOpts(c)
		runOpts.Cmd = c.Args()

		// Validate command
		if len(runOpts.Cmd) == 0 {
			fmt.Printf("No command provided to `devstep run`\n\n")
			cli.ShowCommandHelp(c, "run")
			os.Exit(1)
		}

		// Prepend a `--` so that it doesn't interfere with the current init
		// process args
		runOpts.Cmd = append([]string{"--"}, runOpts.Cmd...)

		result, err := project.Run(client, runOpts)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		os.Exit(result.ExitCode)
        return nil
	},
}
