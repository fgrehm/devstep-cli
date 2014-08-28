package devstep

import (
	"gopkg.in/yaml.v1"
	"os"
	"path/filepath"
)

type ConfigLoader interface {
	Load() (*ProjectConfig, error)
}

type configLoader struct {
	client        DockerClient
	homeDirectory string
	projectRoot   string
}

func (l *configLoader) Load() (*ProjectConfig, error) {
	log.Info("Loading configuration for %s", l.projectRoot)
	// TODO: Load config from home directory
	// TODO: Load config from project directory
	// TODO: Handle errors

	repositoryName := "devstep/" + filepath.Base(l.projectRoot)
	config := &ProjectConfig{
		SourceImage:    "fgrehm/devstep:v0.1.0",
		BaseImage:      "fgrehm/devstep:v0.1.0",
		RepositoryName: repositoryName,
		HostDir:        l.projectRoot,
		GuestDir:       "/workspace",
		CacheDir:       "/tmp/devstep/cache",
	}

	tags, err := l.client.ListTags(repositoryName)
	if err != nil {
		return nil, err
	}
	if len(tags) > 0 {
		config.BaseImage = config.RepositoryName + ":" + tags[0]
	}

	configFile := l.homeDirectory + "/devstep.yml"

	if err = loadConfig(config, configFile); err != nil {
		return nil, err
	}

	log.Info("Loaded config: %+v", config)

	return config, nil
}

func loadConfig(config *ProjectConfig, configPath string) error {
	configInfo, err := os.Stat(configPath)
	// File does not exist or is a directory
	if err != nil || configInfo.IsDir() {
		return nil
	}

	// Parse yaml
	file, err := os.Open(configPath)
	if err != nil {
		return err
	}
	defer file.Close()

	data := make([]byte, configInfo.Size())
	_, err = file.Read(data)

	c := &yamlConfig{}
	err = yaml.Unmarshal(data, &c)

	if err != nil {
		return err
	}

	config.RepositoryName = c.Repository

	return nil
}

type yamlConfig struct {
	Repository string
}

func NewConfigLoader(client DockerClient, homeDirectory, projectRoot string) ConfigLoader {
	return &configLoader{
		client:        client,
		homeDirectory: homeDirectory,
		projectRoot:   projectRoot,
	}
}
