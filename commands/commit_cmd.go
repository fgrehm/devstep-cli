package commands

import (
	"fmt"
	"github.com/codegangsta/cli"
	"os"
)

var CommitCmd = cli.Command{
	Name:  "commit",
	Usage: "commits the currently running container into a Docker image (EXPERIMENTAL)",
	Action: func(c *cli.Context) {
		containerName := os.Getenv("DEVSTEP_CONTAINER_NAME")
		if containerName == "" {
			fmt.Println("This command should be executed from within a Devstep environment")
			os.Exit(1)
		}

		err := project.Commit(client, containerName)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	},
}
