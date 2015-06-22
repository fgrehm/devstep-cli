package devstep

import (
	"bytes"
	"errors"
	"gopkg.in/yaml.v1"
	"os"
	"path/filepath"
	"text/template"
	"time"
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
	RepositoryName *string                 `yaml:"repository"`
	SourceImage    *string                 `yaml:"source_image"`
	CacheDir       *string                 `yaml:"cache_dir"`
	GuestDir       *string                 `yaml:"working_dir"`
	Privileged     *bool                   `yaml:"privileged"`
	Links          []string                `yaml:"links"`
	Volumes        []string                `yaml:"volumes"`
	Env            map[string]string       `yaml:"environment"`
	Hack           *yamlConfig             `yaml:"hack"`
	Commands       map[string]*yamlCommand `yaml:"commands"`
}

type yamlCommand struct {
	Name       *string           `yaml:"name"`
	Cmd        []string          `yaml:"cmd"`
	Privileged *bool             `yaml:"privileged"`
	Links      []string          `yaml:"links"`
	Volumes    []string          `yaml:"volumes"`
	Publish    []string          `yaml:"publish"`
	Env        map[string]string `yaml:"environment"`
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

	config, err := l.buildDefaultConfig()
	if err != nil {
		return nil, err
	}

	yamlConf, err := parseYaml(l.homeDirectory + "/devstep.yml")
	if err != nil {
		return nil, err
	}
	if yamlConf != nil {
		log.Info("Loaded config from home dir")
		log.Debug("Home dir config: %+v", yamlConf)
		if yamlConf.RepositoryName != nil {
			return nil, errors.New("Repository name can't be set globally")
		}
		if yamlConf.Privileged != nil {
			return nil, errors.New("Privileged name can't be set globally")
		}
		assignYamlValues(yamlConf, config)
	}
	yamlConf, err = parseYaml(l.projectRoot + "/devstep.yml")
	if err != nil {
		return nil, err
	}
	if yamlConf != nil {
		log.Info("Loaded config from project dir")
		log.Debug("Project dir config: %+v", yamlConf)
		assignYamlValues(yamlConf, config)
		if yamlConf.Privileged != nil {
			config.Defaults.Privileged = yamlConf.Privileged
		}
	}

	tags, err := l.client.ListTags(config.RepositoryName)
	if err != nil {
		return nil, err
	}
	if len(tags) > 0 {
		config.BaseImage = config.RepositoryName + ":latest"
	} else {
		config.BaseImage = config.SourceImage
	}

	log.Info("Config loaded")
	log.Debug("Final config: %+v", config)

	return config, nil
}

func (l *configLoader) buildDefaultConfig() (*ProjectConfig, error) {
	projectDirName := filepath.Base(l.projectRoot)
	repositoryName := "devstep/" + projectDirName
	suffix := time.Now().Local().Format("20060102150405")

	config := &ProjectConfig{
		SourceImage:    "fgrehm/devstep:v0.4.0",
		RepositoryName: repositoryName,
		HostDir:        l.projectRoot,
		GuestDir:       "/workspace",
		CacheDir:       "/tmp/devstep/cache",
		Defaults: &DockerRunOpts{
			Name:     projectDirName + "-" + suffix,
			Env:      make(map[string]string),
			Hostname: projectDirName,
		},
		HackOpts: &DockerRunOpts{Env: make(map[string]string)},
		Commands: make(map[string]*ProjectCommand),
	}

	return config, nil
}

func parseYaml(configPath string) (*yamlConfig, error) {
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
		return nil, errors.New("Error parsing '" + configPath + "'\n  " + err.Error())
	}

	var b bytes.Buffer
	tmpl.ExecuteTemplate(&b, "config", struct{}{})

	c := &yamlConfig{}
	err = yaml.Unmarshal(b.Bytes(), &c)

	if err != nil {
		return nil, errors.New("Error parsing '" + configPath + "'\n  " + err.Error())
	}

	return c, nil
}

func assignYamlValues(yamlConf *yamlConfig, config *ProjectConfig) {
	if yamlConf.RepositoryName != nil {
		config.RepositoryName = *yamlConf.RepositoryName
	}
	if yamlConf.SourceImage != nil {
		config.SourceImage = *yamlConf.SourceImage
	}
	if yamlConf.CacheDir != nil {
		config.CacheDir = *yamlConf.CacheDir
	}
	if yamlConf.GuestDir != nil {
		config.GuestDir = *yamlConf.GuestDir
	}
	if yamlConf.Links != nil {
		config.Defaults.Links = append(config.Defaults.Links, yamlConf.Links...)
	}
	if yamlConf.Volumes != nil {
		config.Defaults.Volumes = append(config.Defaults.Volumes, yamlConf.Volumes...)
	}
	if yamlConf.Env != nil {
		for k, v := range yamlConf.Env {
			config.Defaults.Env[k] = v
		}
	}

	if yamlConf.Hack != nil {
		if yamlConf.Hack.Links != nil {
			config.HackOpts.Links = append(config.HackOpts.Links, yamlConf.Hack.Links...)
		}
		if yamlConf.Hack.Volumes != nil {
			config.HackOpts.Volumes = append(config.HackOpts.Volumes, yamlConf.Hack.Volumes...)
		}
		if yamlConf.Hack.Env != nil {
			for k, v := range yamlConf.Hack.Env {
				config.HackOpts.Env[k] = v
			}
		}
	}

	if yamlConf.Commands != nil && len(yamlConf.Commands) > 0 {
		for cmdName, yamlCmd := range yamlConf.Commands {
			if yamlCmd == nil {
				yamlCmd = &yamlCommand{}
			}

			cmd := &ProjectCommand{
				cmdName,
				DockerRunOpts{
					Links:   yamlCmd.Links,
					Volumes: yamlCmd.Volumes,
					Env:     yamlCmd.Env,
					Publish: yamlCmd.Publish,
				},
			}

			if yamlCmd.Privileged != nil {
				cmd.Privileged = yamlCmd.Privileged
			}

			if len(yamlCmd.Cmd) > 0 {
				cmd.Cmd = yamlCmd.Cmd
			} else {
				cmd.Cmd = []string{cmdName}
			}

			config.Commands[cmdName] = cmd
		}
	}
}
