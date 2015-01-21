package commands

import (
	"fmt"
	"github.com/codegangsta/cli"
	"os"
)

var BuildCmd = cli.Command{
	Name:  "build",
	Usage: "build a docker image for the current project",
	Flags: dockerRunFlags,
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
		}
	},
	Action: func(c *cli.Context) {
		runOpts := parseRunOpts(c)
		err := newProject().Build(client, runOpts)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	},
}
