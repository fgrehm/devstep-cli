package commands

import (
	"fmt"
	"github.com/urfave/cli"
	"os"
)

var HackCmd = cli.Command{
	Name:  "hack",
	Usage: "start a hacking session for the current project",
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
		runOpts := parseRunOpts(c)
		err := project.Hack(client, runOpts)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
        return nil
	},
}
