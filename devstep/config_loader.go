package devstep

import (
	"bytes"
	"errors"
	"gopkg.in/yaml.v1"
	"os"
	"path/filepath"
	"strings"
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
	RepositoryName *string           `yaml:"repository"`
	SourceImage    *string           `yaml:"source_image"`
	CacheDir       *string           `yaml:"cache_dir"`
	GuestDir       *string           `yaml:"working_dir"`
	Privileged     *bool             `yaml:"privileged"`
	Links          []string          `yaml:"links"`
	Volumes        []string          `yaml:"volumes"`
	Env            map[string]string `yaml:"environment"`
	Hack           *yamlConfig       `yaml:"hack"`
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
		SourceImage:    "fgrehm/devstep:v1.0.0",
		RepositoryName: repositoryName,
		HostDir:        l.projectRoot,
		GuestDir:       "/workspace",
		CacheDir:       "/tmp/devstep/cache",
		Defaults: &DockerRunOpts{
			Name: projectDirName + "-" + suffix,
			// Refactor: This should live somewhere else
			Env:      map[string]string{"DEVSTEP_CONTAINER_NAME": (projectDirName + "-" + suffix)},
			Hostname: projectDirName,
		},
		HackOpts: &DockerRunOpts{Env: make(map[string]string)},
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
		volumes := yamlConf.Volumes
		for i, vol := range volumes {
			hostAndGuestDirs := strings.SplitN(vol, ":", 2)
			hostDir, err := filepath.Abs(hostAndGuestDirs[0])
			if err != nil {
				panic(err)
			}
			guestDir := hostAndGuestDirs[1]
			volumes[i] = hostDir + ":" + guestDir
		}
		config.Defaults.Volumes = append(config.Defaults.Volumes, volumes...)
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
}
