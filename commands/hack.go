package commands

import (
	"fmt"
	"github.com/codegangsta/cli"
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
			fmt.Println("--name")
		}
	},
	Action: func(c *cli.Context) {
		project := newProject()
		runOpts := parseRunOpts(c)
		err := project.Hack(client, runOpts)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	},
}
