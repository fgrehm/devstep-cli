package commands

import (
	"bitbucket.org/kardianos/osext"
	"fmt"
	"github.com/codegangsta/cli"
	"io/ioutil"
	"os"
)

var BinstubsCmd = cli.Command{
	Name:  "binstubs",
	Usage: "Generate binstubs for the commands specified on devstep.yml",
	Action: func(c *cli.Context) {
		project := newProject()
		commands := project.Config().Commands

		if len(commands) == 0 {
			fmt.Println("No binstubs specified!")
			os.Exit(0)
		}

		binstubsPath := ".devstep/bin"
		os.MkdirAll("./"+binstubsPath, 0700)

		executable, _ := osext.Executable()

		for _, cmd := range commands {
			script := []byte("#!/usr/bin/env bash\neval \"" + executable + " run -- " + cmd.Name + " $@\"")
			err := ioutil.WriteFile(binstubsPath+"/"+cmd.Name, script, 0755)
			if err != nil {
				fmt.Printf("Error creating binstub '%s'\n%s\n", binstubsPath+"/"+cmd.Name, err)
				os.Exit(1)
			}
			fmt.Printf("Generated binstub for '%s' in '%s'\n", cmd.Name, binstubsPath)
		}
	},
}
