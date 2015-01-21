package commands

import (
	"fmt"
	"github.com/codegangsta/cli"
	"github.com/fgrehm/devstep-cli/devstep"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var client devstep.DockerClient

var project devstep.Project

var dockerRunFlags = []cli.Flag{
	cli.StringFlag{Name: "w, working_dir", Usage: "Working directory inside the container"},
	cli.StringSliceFlag{Name: "p, publish", Value: &cli.StringSlice{}, Usage: "Publish a container's port to the host (hostPort:containerPort)"},
	cli.StringSliceFlag{Name: "link", Value: &cli.StringSlice{}, Usage: "Add link to another container (name:alias)"},
	cli.StringSliceFlag{Name: "e, env", Value: &cli.StringSlice{}, Usage: "Set environment variables"},
	cli.BoolFlag{Name: "privileged", Usage: "Give extended privileges to this container"},
}

func newProject() devstep.Project {
	config := loadConfig()
	project, _ = devstep.NewProject(config)
	return project
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
