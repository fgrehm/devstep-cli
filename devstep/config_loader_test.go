package devstep_test

import (
	"errors"
	"github.com/fgrehm/devstep-cli/devstep"
	"io/ioutil"
	"os"
	"testing"
)

func Test_Defaults(t *testing.T) {
	client := NewMockClient()
	projectRoot := "/path/to/a-project-dir"
	loader := devstep.NewConfigLoader(client, "", projectRoot)

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
	client := NewMockClient()
	loader := devstep.NewConfigLoader(client, "", "/path/to/a-project")

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
	client := NewMockClient()

	listError := errors.New("Some Error!")
	client.ListTagsFunc = func(repositoryName string) ([]string, error) {
		return nil, listError
	}

	loader := devstep.NewConfigLoader(client, "", "")
	_, err := loader.Load()

	equals(t, listError, err)
}

func Test_LoadConfigFromHomeDir(t *testing.T) {
	tempDir, _ := ioutil.TempDir("", "devstep-project-")
	configFile, _ := os.Create(tempDir + "/devstep.yml")
	defer configFile.Close()
	defer os.RemoveAll(tempDir)

	configFile.WriteString("source_image: 'source/image:tag'")
	configFile.Sync()

	client := NewMockClient()
	loader := devstep.NewConfigLoader(client, tempDir, "")

	config, err := loader.Load()

	ok(t, err)

	equals(t, "source/image:tag", config.SourceImage)
}

func Test_RepositoryNameCantBeSetFromHomeDir(t *testing.T) {
	tempDir, _ := ioutil.TempDir("", "devstep-project-")
	configFile, _ := os.Create(tempDir + "/devstep.yml")
	defer configFile.Close()
	defer os.RemoveAll(tempDir)

	configFile.WriteString("repository: 'custom/repository'")
	configFile.Sync()

	client := NewMockClient()
	loader := devstep.NewConfigLoader(client, tempDir, "")

	_, err := loader.Load()
	assert(t, err != nil, "Repository name was allowed from home dir")
}
