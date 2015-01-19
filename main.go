package main

import (
	"bitbucket.org/kardianos/osext"
	"fmt"
	"github.com/codegangsta/cli"
	"github.com/fgrehm/devstep-cli/devstep"
	"github.com/segmentio/go-prompt"
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

func main() {
	app := cli.NewApp()
	app.Name = "devstep"
	app.Author = "Fábio Rehm"
	app.Email = "fgrehm@gmail.com"
	app.Usage = "development environments made easy"
	app.Version = "0.2.0"
	app.EnableBashCompletion = true
	app.Flags = []cli.Flag{
		cli.StringFlag{Name: "log-level, l", Value: "warning", Usage: "log level", EnvVar: "DEVSTEP_LOG"},
	}
	app.Before = func(c *cli.Context) error {
		return devstep.SetLogLevel(c.GlobalString("log-level"))
	}
	app.Commands = commands

	app.RunAndExitOnError()
}

var dockerRunFlags = []cli.Flag{
	cli.StringFlag{Name: "w, working_dir", Usage: "Working directory inside the container"},
	cli.StringSliceFlag{Name: "p, publish", Value: &cli.StringSlice{}, Usage: "Publish a container's port to the host (hostPort:containerPort)"},
	cli.StringSliceFlag{Name: "link", Value: &cli.StringSlice{}, Usage: "Add link to another container (name:alias)"},
	cli.StringSliceFlag{Name: "e, env", Value: &cli.StringSlice{}, Usage: "Set environment variables"},
	cli.BoolFlag{Name: "privileged", Usage: "Give extended privileges to this container"},
}

func loadConfig() *devstep.ProjectConfig {
	projectRoot, err := os.Getwd()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	dockerHost := os.Getenv("DOCKER_HOST")
	if dockerHost == "" {
		dockerHost = "unix:///var/run/docker.sock"
	}

	homeDir := os.Getenv("HOME")
	client = devstep.NewClient(dockerHost)
	loader := devstep.NewConfigLoader(client, homeDir, projectRoot)

	config, err := loader.Load()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	pluginsToLoad, err := filepath.Glob(homeDir + "/devstep/plugins/*/plugin.js")
	if err != nil {
		fmt.Printf("Error searching for plugins under '%s'\n%s\n", homeDir, err.Error())
		os.Exit(1)
	}

	if len(pluginsToLoad) > 0 {
		runtime := devstep.NewPluginRuntime(config)
		for _, pluginPath := range pluginsToLoad {
			runtime.Load(pluginPath)
		}
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

var buildCmd = cli.Command{
	Name:  "build",
	Usage: "build a docker image for the current project",
	Flags: dockerRunFlags,
	BashComplete: func(c *cli.Context) {
		args := c.Args()
		if len(args) == 0 {
			fmt.Println("-p")
			fmt.Println("--publish")
			fmt.Println("--link")
			fmt.Println("-w")
			fmt.Println("--working_dir")
			fmt.Println("-e")
			fmt.Println("--env")
			fmt.Println("--privileged")
		}
	},
	Action: func(c *cli.Context) {
		runOpts := parseRunOpts(c)
		err := newProject().Build(client, runOpts)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	},
}

var bootstrapCmd = cli.Command{
	Name:  "bootstrap",
	Usage: "bootstrap an environment for the current project",
	Flags: append(
		[]cli.Flag{
			cli.StringFlag{Name: "repository, r", Usage: "set the container repository name"},
		},
		dockerRunFlags...,
	),
	BashComplete: func(c *cli.Context) {
		args := c.Args()
		if len(args) == 0 {
			fmt.Println("-p")
			fmt.Println("--publish")
			fmt.Println("--link")
			fmt.Println("-w")
			fmt.Println("--working_dir")
			fmt.Println("-e")
			fmt.Println("--env")
			fmt.Println("--privileged")
			fmt.Println("--repository")
		}
	},
	Action: func(c *cli.Context) {
		project := newProject()

		if repo := c.String("repository"); repo != "" {
			project.Config().RepositoryName = repo
		}

		runOpts := parseRunOpts(c)
		err := project.Bootstrap(client, runOpts)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	},
}

var hackCmd = cli.Command{
	Name:  "hack",
	Usage: "start a hacking session for the current project",
	Flags: append(
		[]cli.Flag{cli.StringFlag{Name: "name", Usage: "Name to be assigned to the container"}},
		dockerRunFlags...,
	),
	BashComplete: func(c *cli.Context) {
		args := c.Args()
		if len(args) == 0 {
			fmt.Println("-p")
			fmt.Println("--publish")
			fmt.Println("--link")
			fmt.Println("-w")
			fmt.Println("--working_dir")
			fmt.Println("-e")
			fmt.Println("--env")
			fmt.Println("--privileged")
			fmt.Println("--name")
		}
	},
	Action: func(c *cli.Context) {
		project := newProject()
		runOpts := parseRunOpts(c)
		err := project.Hack(client, runOpts)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	},
}

var runCmd = cli.Command{
	Name:  "run",
	Usage: "Run a one off command against the current base image",
	Flags: append(
		[]cli.Flag{cli.StringFlag{Name: "name", Usage: "Name to be assigned to the container"}},
		dockerRunFlags...,
	),
	BashComplete: func(c *cli.Context) {
		args := c.Args()
		if len(args) == 0 {
			fmt.Println("-p")
			fmt.Println("--publish")
			fmt.Println("--link")
			fmt.Println("-w")
			fmt.Println("--working_dir")
			fmt.Println("-e")
			fmt.Println("--env")
			fmt.Println("--privileged")
			fmt.Println("--name")
		}
	},
	Action: func(c *cli.Context) {
		project := newProject()
		commands := project.Config().Commands

		runOpts := parseRunOpts(c)
		runOpts.Cmd = c.Args()

		// Validate command
		if len(runOpts.Cmd) == 0 {
			fmt.Printf("No command provided to `devstep run`\n\n")
			cli.ShowCommandHelp(c, "run")
			os.Exit(1)
		}

		if cmd, ok := commands[runOpts.Cmd[0]]; ok {
			runOpts = cmd.Merge(runOpts)
			runOpts.Cmd = append(cmd.Cmd, runOpts.Cmd[1:]...)
		}

		// Prepend a `--` so that it doesn't interfere with the current init
		// process args
		runOpts.Cmd = append([]string{"--"}, runOpts.Cmd...)

		result, err := project.Run(client, runOpts)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		os.Exit(result.ExitCode)
	},
}

var binstubsCmd = cli.Command{
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

var cleanCmd = cli.Command{
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

var pristineCmd = cli.Command{
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
		args := c.Args()
		if len(args) == 0 {
			fmt.Println("-p")
			fmt.Println("--publish")
			fmt.Println("--link")
			fmt.Println("-w")
			fmt.Println("--working_dir")
			fmt.Println("-e")
			fmt.Println("--env")
			fmt.Println("--privileged")
			fmt.Println("-f")
			fmt.Println("--force")
			fmt.Println("-b")
			fmt.Println("--bootstrap")
			fmt.Println("--repository")
		}
	},
	Action: func(c *cli.Context) {
		// TODO: Figure out if this is the right way to invoke other CLI actions
		cleanCmd.Action(c)
		if c.Bool("bootstrap") {
			bootstrapCmd.Action(c)
		} else {
			buildCmd.Action(c)
		}
	},
}

var infoCmd = cli.Command{
	Name:  "info",
	Usage: "show information about the current environment",
	Action: func(c *cli.Context) {
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
	privileged := false
	if opts.Privileged != nil {
		privileged = *opts.Privileged
	}
	fmt.Printf("%sPrivileged: %v\n", prefix, privileged)
	fmt.Printf("%sLinks:      %v\n", prefix, opts.Links)
	fmt.Printf("%sVolumes:    %v\n", prefix, opts.Volumes)
	fmt.Printf("%sEnv:        %v\n", prefix, opts.Env)
}

func parseRunOpts(c *cli.Context) *devstep.DockerRunOpts {
	runOpts := &devstep.DockerRunOpts{
		Publish: c.StringSlice("publish"),
		Links:   c.StringSlice("link"),
		Name:    c.String("name"),
		Env:     make(map[string]string),
	}

	// Only set the privileged config if it was provided
	if c.IsSet("privileged") {
		privileged := c.Bool("privileged")
		runOpts.Privileged = &privileged
	}

	// Set working dir directly on the project object so that it get passed
	// along properly
	if workingDir := c.String("working_dir"); workingDir != "" {
		project.Config().GuestDir = workingDir
	}

	// Env vars
	for _, envVar := range c.StringSlice("env") {
		varAndValue := strings.Split(envVar, "=")
		runOpts.Env[varAndValue[0]] = varAndValue[1]
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
