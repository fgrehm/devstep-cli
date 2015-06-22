package commands

import (
	"fmt"
	"github.com/codegangsta/cli"
	"os"
)

var ExecCmd = cli.Command{
	Name:  "exec",
	Usage: "Run a one off command against the last container created for the current project",
	Action: func(c *cli.Context) {
		execCmd := c.Args()

		// Validate command
		if len(execCmd) == 0 {
			fmt.Printf("No command provided to `devstep exec`\n\n")
			cli.ShowCommandHelp(c, "exec")
			os.Exit(1)
		}

		commands := project.Config().Commands
		if cmd, ok := commands[execCmd[0]]; ok {
			 execCmd = append(cmd.Cmd, execCmd[1:]...)
		}

		// Prepend a `--` so that it doesn't interfere with the current init
		// process args
		execCmd = append([]string{"--"}, execCmd...)

		err := project.Exec(client, execCmd)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	},
}
