package devstep

import "path/filepath"

type ConfigLoader interface {
	Load() *ProjectConfig
}

type configLoader struct {
	client        DockerClient
	homeDirectory string
	projectRoot   string
}

func (l *configLoader) Load() *ProjectConfig {
	repositoryName := "devstep/" + filepath.Base(l.projectRoot)

	return &ProjectConfig{
		SourceImage:    "fgrehm/devstep:v0.1.0",
		BaseImage:      "fgrehm/devstep:v0.1.0",
		RepositoryName: repositoryName,
		HostDir:        l.projectRoot,
		GuestDir:       "/workspace",
		CacheDir:       "/tmp/devstep/cache",
	}
}

func NewConfigLoader(client DockerClient, homeDirectory, projectRoot string) ConfigLoader {
	return &configLoader{
		client:        client,
		homeDirectory: homeDirectory,
		projectRoot:   projectRoot,
	}
}
