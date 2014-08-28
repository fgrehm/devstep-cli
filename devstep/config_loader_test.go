package devstep_test

import (
	"errors"
	"github.com/fgrehm/devstep-cli/devstep"
	"io/ioutil"
	"os"
	"testing"
)

func Test_Defaults(t *testing.T) {
	projectRoot := "/path/to/a-project-dir"
	loader, _ := newConfigLoader("", projectRoot)

	config, err := loader.Load()

	ok(t, err)

	equals(t, "fgrehm/devstep:v0.1.0", config.SourceImage)
	equals(t, "fgrehm/devstep:v0.1.0", config.BaseImage)
	equals(t, projectRoot, config.HostDir)
	equals(t, "/workspace", config.GuestDir)
	equals(t, "/tmp/devstep/cache", config.CacheDir)
	equals(t, "devstep/a-project-dir", config.RepositoryName)
}

func Test_SourceImageGetsSetWhenRepositoryTagExists(t *testing.T) {
	loader, client := newConfigLoader("", "/path/to/a-project")

	var repositoryNameSearched string
	client.ListTagsFunc = func(r string) ([]string, error) {
		repositoryNameSearched = r
		return []string{"a-tag", "other-tag"}, nil
	}

	config, err := loader.Load()

	ok(t, err)
	equals(t, "devstep/a-project", repositoryNameSearched)
	equals(t, "devstep/a-project:a-tag", config.BaseImage)
}

func Test_ErrorWhenListTagsFails(t *testing.T) {
	loader, client := newConfigLoader("", "")

	listError := errors.New("Some Error!")
	client.ListTagsFunc = func(repositoryName string) ([]string, error) {
		return nil, listError
	}

	_, err := loader.Load()

	equals(t, listError, err)
}

func Test_LoadConfigFromHomeDir(t *testing.T) {
	tempDir, _ := ioutil.TempDir("", "devstep-project-")
	writeFile(tempDir+"/devstep.yml", `
source_image: 'source/image:tag'
cache_dir:    '/custom/cache/dir'
`)
	defer os.RemoveAll(tempDir)

	loader, _ := newConfigLoader(tempDir, "")
	config, err := loader.Load()

	ok(t, err)

	equals(t, "source/image:tag", config.SourceImage)
	equals(t, "/custom/cache/dir", config.CacheDir)
}

func Test_RepositoryNameCantBeSetFromHomeDir(t *testing.T) {
	tempDir, _ := ioutil.TempDir("", "devstep-project-")
	writeFile(tempDir+"/devstep.yml", "repository: 'custom/repository'")
	defer os.RemoveAll(tempDir)

	loader, _ := newConfigLoader(tempDir, "")

	_, err := loader.Load()
	assert(t, err != nil, "Repository name was allowed from home dir")
}

func newConfigLoader(homeDir, projectDir string) (devstep.ConfigLoader, *MockClient) {
	client := NewMockClient()
	loader := devstep.NewConfigLoader(client, homeDir, projectDir)

	return loader, client
}

func writeFile(path, yaml string) {
	file, _ := os.Create(path)
	defer file.Close()
	file.WriteString(yaml)
	file.Sync()
}
