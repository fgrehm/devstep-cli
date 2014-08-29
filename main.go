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
	runCmd,
	cleanCmd,
	infoCmd,
}

var dockerRunFlags = []cli.Flag{
	cli.StringSliceFlag{Name: "p, publish", Value: &cli.StringSlice{}, Usage: "Publish a container's port to the host"},
	cli.StringSliceFlag{Name: "link", Value: &cli.StringSlice{}, Usage: "Add link to another container (name:alias)"},
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

	if devstep.LogLevel != "" {
		config.Defaults.Env["DEVSTEP_LOG"] = devstep.LogLevel
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
		cli.StringFlag{Name: "log-level, l", Usage: "log level", EnvVar: "DEVSTEP_LOG"},
	}
	app.Commands = commands

	app.Run(os.Args)
}

var buildCmd = cli.Command{
	Name:  "build",
	Usage: "build a docker image for the current project",
	Action: func(c *cli.Context) {
		devstep.SetLogLevel(c.GlobalString("log-level"))

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
	Flags: dockerRunFlags,
	Action: func(c *cli.Context) {
		devstep.SetLogLevel(c.GlobalString("log-level"))

		runOpts := parseRunOpts(c)
		err := newProject().Hack(client, runOpts)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	},
}

var runCmd = cli.Command{
	Name:  "run",
	Usage: "Run a one off command against the current base image",
	Flags: dockerRunFlags,
	Action: func(c *cli.Context) {
		devstep.SetLogLevel(c.GlobalString("log-level"))

		runOpts := parseRunOpts(c)
		runOpts.Cmd = c.Args()

		// Validate command
		if len(runOpts.Cmd) == 0 {
			fmt.Println("No command provided to `devstep run`\n")
			cli.ShowCommandHelp(c, "run")
			os.Exit(1)
		}

		project := newProject()
		commands := project.Config().Commands

		if cmd, ok := commands[runOpts.Cmd[0]]; ok {
			runOpts = cmd.Merge(runOpts)
			runOpts.Cmd = append(cmd.Cmd, runOpts.Cmd[1:]...)
		}

		// Prepend a `--` so that it doesn't interfere with the current init
		// process args
		runOpts.Cmd = append([]string{"--"}, runOpts.Cmd...)

		err := project.Run(client, runOpts)
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
		devstep.SetLogLevel(c.GlobalString("log-level"))

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
		devstep.SetLogLevel(c.GlobalString("log-level"))

		config := loadConfig()
		fmt.Printf("\nConfig:\n\t%+v", config)
		if config.Defaults != nil {
			fmt.Printf("\n\nDefaults:\n\t%+v\n", config.Defaults)
		}
		if config.HackOpts != nil {
			fmt.Printf("\nHack:\n\t%+v\n", config.HackOpts)
		}
		if config.Commands != nil {
			fmt.Println("\nCommands:")
			for _, cmd := range config.Commands {
				fmt.Printf("\t%s -> %+v\n", cmd.Name, cmd.DockerRunOpts)
			}
		}
	},
}

func parseRunOpts(c *cli.Context) *devstep.DockerRunOpts {
	runOpts := &devstep.DockerRunOpts{
		Publish: c.StringSlice("publish"),
		Links:   c.StringSlice("link"),
	}
	// Validate ports
	validPort := regexp.MustCompile(`^\d+:\d+$`)
	for _, port := range runOpts.Publish {
		if !validPort.MatchString(port) {
			fmt.Println("Invalid publish arg: " + port)
			os.Exit(1)
		}
	}
	// Validate links
	validLink := regexp.MustCompile(`[^:]+:[^:]+`)
	for _, link := range runOpts.Links {
		if !validLink.MatchString(link) {
			fmt.Println("Invalid link: " + link)
			os.Exit(1)
		}
	}
	return runOpts
}
