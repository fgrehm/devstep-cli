package devstep

import (
	"bytes"
	"errors"
	"gopkg.in/yaml.v1"
	"os"
	"path/filepath"
	"text/template"
)

type ConfigLoader interface {
	Load() (*ProjectConfig, error)
}

type configLoader struct {
	client        DockerClient
	homeDirectory string
	projectRoot   string
}

type yamlConfig struct {
	RepositoryName string `yaml:"repository"`
	SourceImage    string `yaml:"source_image"`
	CacheDir       string `yaml:"cache_dir"`
}

func NewConfigLoader(client DockerClient, homeDirectory, projectRoot string) ConfigLoader {
	return &configLoader{
		client:        client,
		homeDirectory: homeDirectory,
		projectRoot:   projectRoot,
	}
}

func (l *configLoader) Load() (*ProjectConfig, error) {
	log.Info("Loading configuration for %s", l.projectRoot)
	// TODO: Load config from project directory
	// TODO: Handle errors

	config, err := l.buildDefaultConfig()
	if err != nil {
		return nil, err
	}

	yamlConf, err := loadConfig(l.homeDirectory + "/devstep.yml")
	if err != nil {
		return nil, err
	}
	if yamlConf != nil {
		log.Info("Loaded config from home dir")
		log.Debug("Home dir config: %+v", yamlConf)
		if yamlConf.RepositoryName != "" {
			return nil, errors.New("Repository name can't be set globally")
		}

		config.SourceImage = yamlConf.SourceImage
		config.CacheDir = yamlConf.CacheDir
	}

	log.Info("Config loaded")
	log.Debug("Final config: %+v", config)

	return config, nil
}

func (l *configLoader) buildDefaultConfig() (*ProjectConfig, error) {
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

	return config, nil
}

func loadConfig(configPath string) (*yamlConfig, error) {
	configInfo, err := os.Stat(configPath)
	// File does not exist or is a directory
	if err != nil || configInfo.IsDir() {
		return nil, nil
	}

	// Parse yaml
	file, err := os.Open(configPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	data := make([]byte, configInfo.Size())
	_, err = file.Read(data)

	funcMap := template.FuncMap{
		"env": os.Getenv,
	}

	tmpl, err := template.New("config").Funcs(funcMap).Parse(string(data))
	if err != nil {
		return nil, err
	}

	var b bytes.Buffer
	tmpl.ExecuteTemplate(&b, "config", struct{}{})

	c := &yamlConfig{}
	err = yaml.Unmarshal(b.Bytes(), &c)

	if err != nil {
		return nil, err
	}

	return c, nil
}
