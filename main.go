package main

import (
	"fmt"
	"github.com/codegangsta/cli"
	"github.com/fgrehm/devstep-cli/devstep"
	"os"
	"regexp"
)

var (
	project devstep.Project
	client  devstep.DockerClient
)

var commands = []cli.Command{
	buildCmd,
	hackCmd,
	cleanCmd,
	infoCmd,
}

func projectRoot() string {
	pwd, err := os.Getwd()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	return pwd
}

func homeDir() string {
	return os.Getenv("HOME")
}

func loadConfig() *devstep.ProjectConfig {
	client = devstep.NewClient("unix:///var/run/docker.sock")
	loader := devstep.NewConfigLoader(client, homeDir(), projectRoot())

	config, err := loader.Load()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	return config
}

func newProject() devstep.Project {
	config := loadConfig()
	project, _ = devstep.NewProject(config)
	return project
}

func main() {
	app := cli.NewApp()
	app.Name = "devstep"
	app.Author = "Fábio Rehm"
	app.Email = "fgrehm@gmail.com"
	app.Usage = "development environments made easy"
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
		devstep.Verbose(c.GlobalBool("debug"))

		err := newProject().Build(client)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	},
}

var hackCmd = cli.Command{
	Name:  "hack",
	Usage: "start a hacking session for the current project",
	Flags: []cli.Flag{
		cli.StringSliceFlag{Name: "p, publish", Value: &cli.StringSlice{}, Usage: "Publish a container's port to the host"},
	},
	Action: func(c *cli.Context) {
		devstep.Verbose(c.GlobalBool("debug"))

		runOpts := &devstep.DockerRunOpts{
			Publish: c.StringSlice("publish"),
		}

		// Validate ports
		validPort := regexp.MustCompile(`\d+:\d+`)
		for _, port := range runOpts.Publish {
			if ! validPort.MatchString(port) {
				fmt.Println("Invalid publish arg: " + port)
				os.Exit(1)
			}
		}

		err := newProject().Hack(client, runOpts)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	},
}

var cleanCmd = cli.Command{
	Name:  "clean",
	Usage: "remove previously built images for the current environment",
	Action: func(c *cli.Context) {
		devstep.Verbose(c.GlobalBool("debug"))

		err := newProject().Clean(client)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	},
}

var infoCmd = cli.Command{
	Name:  "info",
	Usage: "show information about the current environment",
	Action: func(c *cli.Context) {
		devstep.Verbose(c.GlobalBool("debug"))
		config := loadConfig()
		fmt.Printf("%+v\n", config)
		if config.Defaults != nil {
			fmt.Printf("%+v\n", config.Defaults)
		}
		if config.HackOpts != nil {
			fmt.Printf("%+v\n", config.HackOpts)
		}
	},
}
