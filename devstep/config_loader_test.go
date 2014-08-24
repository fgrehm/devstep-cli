package devstep_test

import (
	"github.com/fgrehm/devstep-cli/devstep"
	"testing"
)

func Test_Defaults(t *testing.T) {
	client := NewMockClient()
	projectRoot := "/path/to/a-project-dir"
	loader := devstep.NewConfigLoader(client, "", projectRoot)

	config := loader.Load()

	equals(t, "fgrehm/devstep:v0.1.0", config.BaseImage)
	equals(t, projectRoot, config.HostDir)
	equals(t, "/workspace", config.GuestDir)
	equals(t, "/tmp/devstep/cache", config.CacheDir)
	equals(t, "devstep/a-project-dir", config.RepositoryName)
}
