package devstep

type ConfigLoader interface {
	Load() *ProjectConfig
}

type configLoader struct {
	homeDirectory string
	projectRoot   string
}

func (l *configLoader) Load() *ProjectConfig {
	return &ProjectConfig{}
}

func NewConfigLoader(client DockerClient, homeDirectory, projectRoot string) ConfigLoader {
	return &configLoader{
		homeDirectory: homeDirectory,
		projectRoot:   projectRoot,
	}
}
