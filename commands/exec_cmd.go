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
		// Validate command
		if len(c.Args()) == 0 {
			fmt.Printf("No command provided to `devstep exec`\n\n")
			cli.ShowCommandHelp(c, "exec")
			os.Exit(1)
		}

		err := project.Exec(client, c.Args())
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	},
}
