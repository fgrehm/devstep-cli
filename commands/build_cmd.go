package commands

import (
	"fmt"
	"github.com/codegangsta/cli"
	"os"
)

var BuildCmd = cli.Command{
	Name:         "build",
	Usage:        "build a docker image for the current project",
	Flags:        dockerRunFlags,
	BashComplete: bashCompleteRunArgs,
	Action: func(c *cli.Context) {
		runOpts := parseRunOpts(c)
		err := project.Build(client, runOpts)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	},
}
