package commands

import (
	"fmt"
	"github.com/codegangsta/cli"
	"os"
)

var BootstrapCmd = cli.Command{
	Name:  "bootstrap",
	Usage: "bootstrap an environment for the current project",
	Flags: append(
		[]cli.Flag{
			cli.StringFlag{Name: "repository, r", Usage: "set the container repository name"},
		},
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
			fmt.Println("--repository")
		}
	},
	Action: func(c *cli.Context) {
		project := newProject()

		if repo := c.String("repository"); repo != "" {
			project.Config().RepositoryName = repo
		}

		runOpts := parseRunOpts(c)
		err := project.Bootstrap(client, runOpts)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	},
}
