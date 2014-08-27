package devstep_test

import (
	"errors"
	"github.com/fgrehm/devstep-cli/devstep"
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

	client.ListTagsFunc = func(repositoryName string) ([]string, error) {
		return []string{"a-tag", "other-tag"}, nil
	}

	config, err := loader.Load()

	ok(t, err)
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
