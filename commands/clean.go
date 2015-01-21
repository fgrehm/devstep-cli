package commands

import (
	"fmt"
	"github.com/codegangsta/cli"
	"github.com/segmentio/go-prompt"
	"os"
)

var CleanCmd = cli.Command{
	Name:  "clean",
	Usage: "remove previously built images for the current environment",
	Flags: []cli.Flag{
		cli.BoolFlag{Name: "force, f", Usage: "skip confirmation"},
	},
	BashComplete: func(c *cli.Context) {
		args := c.Args()
		if len(args) == 0 {
			fmt.Println("-f")
			fmt.Println("--force")
		}
	},
	Action: func(c *cli.Context) {
		if !c.Bool("force") {
			if ok := prompt.Confirm("Are you sure? [y/n]"); !ok {
				fmt.Println("Aborting")
				os.Exit(1)
			}
		}

		err := newProject().Clean(client)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	},
}
