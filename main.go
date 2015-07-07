package main

import (
	"github.com/codegangsta/cli"
	"github.com/fgrehm/devstep-cli/commands"
	"github.com/fgrehm/devstep-cli/devstep"
)

func main() {
	app := cli.NewApp()
	app.Name = "devstep"
	app.Author = "Fábio Rehm"
	app.Email = "fgrehm@gmail.com"
	app.Usage = "development environments made easy"
	app.Version = "0.4.0.dev"
	app.EnableBashCompletion = true
	app.Flags = []cli.Flag{
		cli.StringFlag{Name: "log-level, l", Value: "warning", Usage: "log level", EnvVar: "DEVSTEP_LOG"},
	}
	app.Before = func(c *cli.Context) error {
		commands.InitDevstepEnv()
		return devstep.SetLogLevel(c.GlobalString("log-level"))
	}
	app.Commands = []cli.Command{
		commands.HackCmd,
		commands.BuildCmd,
		commands.BootstrapCmd,
		commands.InfoCmd,
		commands.ExecCmd,
		commands.RunCmd,
		commands.BinstubsCmd,
		commands.CleanCmd,
		commands.PristineCmd,
		commands.InitCmd,
	}

	app.RunAndExitOnError()
}
