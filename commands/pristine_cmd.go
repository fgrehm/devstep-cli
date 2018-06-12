package commands

import (
	"fmt"
	"github.com/urfave/cli"
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
	Action: func(c *cli.Context) error {
		CleanCmd.Run(c)
		reloadProject()
		if c.Bool("bootstrap") {
			BootstrapCmd.Run(c)
		} else {
			BuildCmd.Run(c)
		}
        return nil
	},
}
