package main

import (
	"github.com/codegangsta/cli"
	"github.com/fgrehm/devstep-cli/devstep"
	"os"
)

var project devstep.Project

var commands = []cli.Command{
	buildCmd,
	hackCmd,
}

func main() {
	// TODO: Error handling
	project, _ = devstep.NewProject()

	app := cli.NewApp()
	app.Name = "devstep"
	app.Author = "Fábio Rehm"
	app.Email = "fgrehm@gmail.com"
	app.Usage = "build development environments with ease"
	app.Version = "0.1.0"
	app.Flags = []cli.Flag{
		cli.BoolFlag{Name: "debug, d", Usage: "debug mode"},
	}
	app.Commands = commands

	app.Run(os.Args)
}

var buildCmd = cli.Command{
	Name:  "build",
	Usage: "build a docker image for the current project",
	Action: func(c *cli.Context) {
		project.Build()
	},
}

var hackCmd = cli.Command{
	Name:  "hack",
	Usage: "start a hacking session for the current project",
	Action: func(c *cli.Context) {
		project.Hack()
	},
}
