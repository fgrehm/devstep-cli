package main

import (
	"bufio"
	"fmt"
	"github.com/codegangsta/cli"
	"github.com/fgrehm/devstep-cli/devstep"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var (
	project devstep.Project
	client  devstep.DockerClient
)

var commands = []cli.Command{
	hackCmd,
	buildCmd,
	bootstrapCmd,
	infoCmd,
	runCmd,
	binstubsCmd,
	cleanCmd,
	pristineCmd,
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

	pluginsToLoad, err := filepath.Glob(homeDir() + "/devstep/plugins/*/plugin.js")
	if err != nil {
		fmt.Printf("Error searching for plugins under '%s'\n%s\n", homeDir(), err.Error())
		os.Exit(1)
	}

	if len(pluginsToLoad) > 0 {
		// fmt.Printf("Loading plugins from home dir: %+v\n", pluginsToLoad)
		runtime := devstep.NewPluginRuntime(config)
		for _, pluginPath := range(pluginsToLoad) {
			file, err := os.Open(pluginPath)
			if err != nil {
				fmt.Printf("Error loading plugin '%s'\n%s\n", pluginPath, err.Error())
				os.Exit(1)
			}
			runtime.Load(filepath.Dir(pluginPath), file)
		}
		// log.Info("Triggering plugins:configLoaded")
		runtime.Trigger("configLoaded")
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
	app.EnableBashCompletion = true
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

var bootstrapCmd = cli.Command{
	Name:  "bootstrap",
	Usage: "bootstrap an environment for the current project",
	Action: func(c *cli.Context) {
		devstep.SetLogLevel(c.GlobalString("log-level"))

		err := newProject().Bootstrap(client)
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

var binstubsCmd = cli.Command{
	Name:  "binstubs",
	Usage: "Generate binstubs for the commands specified on devstep.yml",
	Action: func(c *cli.Context) {
		devstep.SetLogLevel(c.GlobalString("log-level"))

		project := newProject()
		commands := project.Config().Commands

		if len(commands) == 0 {
			fmt.Println("No binstubs specified!")
			os.Exit(0)
		}

		binstubsPath := ".devstep/bin"
		os.MkdirAll("./"+binstubsPath, 0700)

		for _, cmd := range commands {
			script := []byte("#!/usr/bin/env bash\neval \"devstep-cli run -- " + cmd.Name + " $@\"")
			err := ioutil.WriteFile(binstubsPath+"/"+cmd.Name, script, 0755)
			if err != nil {
				fmt.Printf("Error creating binstub '%s'\n%s\n", binstubsPath+"/"+cmd.Name, err)
				os.Exit(1)
			}
			fmt.Printf("Generated binstub for '%s' in '%s'\n", cmd.Name, binstubsPath)
		}
	},
}

var cleanCmd = cli.Command{
	Name:  "clean",
	Usage: "remove previously built images for the current environment",
	Flags: []cli.Flag{
		cli.BoolFlag{Name: "force, f", Usage: "skip confirmation"},
	},
	Action: func(c *cli.Context) {
		devstep.SetLogLevel(c.GlobalString("log-level"))

		if !c.Bool("force") {
			fmt.Print("Are you sure? [N/y] ")

			reader := bufio.NewReader(os.Stdin)
			line, _, err := reader.ReadLine()
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}

			if strings.ToUpper(string(line)) != "Y" {
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

var pristineCmd = cli.Command{
	Name: "pristine",
	Usage: "rebuild project image from scratch",
	Flags: []cli.Flag{
		cli.BoolFlag{Name: "force, f", Usage: "skip clean confirmation"},
	},
	Action: func(c *cli.Context) {
		// TODO: Figure out if this is the right way to invoke other CLI actions
		cleanCmd.Action(c)
		buildCmd.Action(c)
	},
}

var infoCmd = cli.Command{
	Name:  "info",
	Usage: "show information about the current environment",
	Action: func(c *cli.Context) {
		devstep.SetLogLevel(c.GlobalString("log-level"))

		config := loadConfig()
		printConfig(config)
	},
}

func printConfig(config *devstep.ProjectConfig) {
	fmt.Println("==> Project info")
	fmt.Printf("Repository:   %s\n", config.RepositoryName)
	fmt.Printf("Source image: %s\n", config.SourceImage)
	fmt.Printf("Base image:   %s\n", config.BaseImage)
	fmt.Printf("Host dir:     %s\n", config.HostDir)
	fmt.Printf("Guest dir:    %s\n", config.GuestDir)
	fmt.Printf("Cache dir:    %s\n", config.CacheDir)

	if config.Defaults != nil {
		fmt.Println("\n==> Default options:")
		printDockerRunOpts(config.Defaults, "")
	}

	if config.HackOpts != nil {
		fmt.Println("\n==> Hack options:")
		printDockerRunOpts(config.HackOpts, "")
	}
	if config.Commands != nil {
		fmt.Println("\n==> Commands:")
		for _, cmd := range config.Commands {
			fmt.Printf("* %s\n", cmd.Name)
			fmt.Printf("  Cmd:        %v\n", cmd.Cmd)
			fmt.Printf("  Publish:    %v\n", cmd.Publish)
			printDockerRunOpts(&cmd.DockerRunOpts, "  ")
		}
	}
}

func printDockerRunOpts(opts *devstep.DockerRunOpts, prefix string) {
	fmt.Printf("%sPrivileged: %v\n", prefix, opts.Privileged)
	fmt.Printf("%sLinks:      %v\n", prefix, opts.Links)
	fmt.Printf("%sVolumes:    %v\n", prefix, opts.Volumes)
	fmt.Printf("%sEnv:        %v\n", prefix, opts.Env)
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
