package devstep_test

import (
	"github.com/fgrehm/devstep-cli/devstep"
	"testing"
)

func Test_Hack(t *testing.T) {
	project, err := devstep.NewProject(&devstep.ProjectConfig{
		BaseImage: "repo/name:tag",
		HostDir:   "/path/on/host",
		GuestDir:  "/path/on/guest",
		CacheDir:  "/cache/path/on/host",
	})
	ok(t, err)

	var runOpts *devstep.DockerRunOpts
	clientMock := NewMockClient()
	clientMock.RunFunc = func(o *devstep.DockerRunOpts) (*devstep.DockerRunResult, error) {
		runOpts = o
		return nil, nil
	}

	err = project.Hack(clientMock)
	ok(t, err)

	equals(t, "repo/name:tag", runOpts.Image)
	assert(t, runOpts.AutoRemove, "AutoRemove is false")
	assert(t, runOpts.Pty, "Pseudo tty allocation is disabled")
	equals(t, []string{"/.devstep/bin/hack"}, runOpts.Cmd)
	equals(t, "/path/on/guest", runOpts.Workdir)

	assert(t, inArray("/path/on/host:/path/on/guest", runOpts.Volumes), "Project dir was not shared")
	assert(t, inArray("/tmp/devstep/cache:/.devstep/cache", runOpts.Volumes), "Cache dir was not shared")
}

func inArray(str string, array[]string) bool {
	for index := range array {
		if str == array[index] {
			return true
		}
	}
	return false
}
