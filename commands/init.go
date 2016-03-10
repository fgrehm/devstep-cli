package commands

import (
	"fmt"
	"github.com/fgrehm/devstep-cli/devstep"
	"os"
	"path/filepath"
)

var (
	client  devstep.DockerClient
	project devstep.Project
)

func InitDevstepEnv() {
	client = devstep.NewClient()
	reloadProject()
}

func reloadProject() {
	project = newProject()
}

func newProject() devstep.Project {
	config := loadConfig()
	proj, _ := devstep.NewProject(config)
	return proj
}

func loadConfig() *devstep.ProjectConfig {
	projectRoot, err := os.Getwd()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	homeDir := os.Getenv("HOME")
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
