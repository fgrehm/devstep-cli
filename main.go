package main

import (
	"os"

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
	app.Version = "1.0.0"
	app.EnableBashCompletion = true
	app.Flags = []cli.Flag{
		cli.StringFlag{Name: "log-level, l", Value: "warning", Usage: "log level", EnvVar: "DEVSTEP_LOG"},
	}
	app.Before = func(c *cli.Context) error {
		commands.InitDevstepEnv()
		return devstep.SetLogLevel(c.GlobalString("log-level"))
	}

	containerName := os.Getenv("DEVSTEP_CONTAINER_NAME")
	if containerName == "" {
		app.Commands = []cli.Command{
			commands.BootstrapCmd,
			commands.BuildCmd,
			commands.CleanCmd,
			commands.ExecCmd,
			commands.HackCmd,
			commands.InfoCmd,
			commands.InitCmd,
			commands.PristineCmd,
			commands.RunCmd,
		}
	} else { // inside container
		app.Commands = []cli.Command{
			commands.CommitCmd,
			commands.InfoCmd,
			commands.InitCmd,
		}
	}

	app.RunAndExitOnError()
}
