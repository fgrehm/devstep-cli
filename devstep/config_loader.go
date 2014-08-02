package devstep

type ConfigLoader interface {
	Load() *ProjectConfig
}

type configLoader struct {
	client        DockerClient
	homeDirectory string
	projectRoot   string
}

func (l *configLoader) Load() *ProjectConfig {
	return &ProjectConfig{
		BaseImage: "fgrehm/devstep:v0.1.0",
		HostDir:   l.projectRoot,
		GuestDir:  "/workspace",
		CacheDir:  "/tmp/devstep/cache",
	}
}

func NewConfigLoader(client DockerClient, homeDirectory, projectRoot string) ConfigLoader {
	return &configLoader{
		client:        client,
		homeDirectory: homeDirectory,
		projectRoot:   projectRoot,
	}
}
