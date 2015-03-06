package commands

import (
	"fmt"
	"github.com/codegangsta/cli"
)

var PristineCmd = cli.Command{
	Name:  "pristine",
	Usage: "rebuild project image from scratch",
	Flags: append(
		[]cli.Flag{
			cli.BoolFlag{Name: "force, f", Usage: "skip clean confirmation"},
			cli.BoolFlag{Name: "bootstrap, b", Usage: "manually bootstrap your environment"},
			cli.StringFlag{Name: "repository, r", Usage: "set the container repository name"},
		},
		dockerRunFlags...,
	),
	BashComplete: func(c *cli.Context) {
		bashCompleteRunArgs(c)
		args := c.Args()
		if len(args) == 0 {
			fmt.Println("-f")
			fmt.Println("--force")
			fmt.Println("-b")
			fmt.Println("--bootstrap")
			fmt.Println("--repository")
		}
	},
	Action: func(c *cli.Context) {
		// TODO: Figure out if this is the right way to invoke other CLI actions
		CleanCmd.Action(c)
		reloadProject()
		if c.Bool("bootstrap") {
			BootstrapCmd.Action(c)
		} else {
			BuildCmd.Action(c)
		}
	},
}
