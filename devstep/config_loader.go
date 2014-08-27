package devstep

import "path/filepath"

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

	log.Info("Loaded config: %+v", config)

	return config, nil
}

func NewConfigLoader(client DockerClient, homeDirectory, projectRoot string) ConfigLoader {
	return &configLoader{
		client:        client,
		homeDirectory: homeDirectory,
		projectRoot:   projectRoot,
	}
}
