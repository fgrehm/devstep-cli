package devstep_test

import (
	"errors"
	"github.com/fgrehm/devstep-cli/devstep"
	"strings"
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

func Test_Build(t *testing.T) {
	project, err := devstep.NewProject(&devstep.ProjectConfig{
		BaseImage: "repo/name:tag",
		HostDir:   "/path/on/host",
		GuestDir:  "/path/on/guest",
		CacheDir:  "/cache/path/on/host",
		RepositoryName: "repo-name",
	})
	ok(t, err)

	var runOpts *devstep.DockerRunOpts
	clientMock := NewMockClient()
	clientMock.RunFunc = func(o *devstep.DockerRunOpts) (*devstep.DockerRunResult, error) {
		runOpts = o
		return &devstep.DockerRunResult{
			ExitCode: 0,
			ContainerID: "cid",
		}, nil
	}
	var commitOpts []*devstep.DockerCommitOpts
	clientMock.CommitFunc = func(o *devstep.DockerCommitOpts) error {
		commitOpts = append(commitOpts, o)
		return nil
	}
	var removeId string
	clientMock.RemoveContainerFunc = func(r string) error {
		removeId = r
		return nil
	}

	err = project.Build(clientMock)
	ok(t, err)

	equals(t, "repo/name:tag", runOpts.Image)
	assert(t, !runOpts.AutoRemove, "AutoRemove is true")
	assert(t, runOpts.Pty, "Pseudo tty allocation is disabled")
	equals(t, []string{"/.devstep/bin/build-project", "/path/on/guest"}, runOpts.Cmd)
	equals(t, "/path/on/guest", runOpts.Workdir)
	assert(t, inArray("/path/on/host:/path/on/guest", runOpts.Volumes), "Project dir was not shared")
	assert(t, inArray("/tmp/devstep/cache:/.devstep/cache", runOpts.Volumes), "Cache dir was not shared")

	equals(t, 2, len(commitOpts))

	latestTagCommit := commitOpts[0]
	equals(t, "cid", latestTagCommit.ContainerID)
	equals(t, "repo-name", latestTagCommit.RepositoryName)
	equals(t, "latest", latestTagCommit.Tag)

	timestampTagCommit := commitOpts[1]
	equals(t, "cid", timestampTagCommit.ContainerID)
	equals(t, "repo-name", timestampTagCommit.RepositoryName)
	// DISCUSS: How can we mock or test the tag value?
	assert(t, timestampTagCommit.Tag != "", "Tag not set")

	equals(t, "cid", removeId)
}

func Test_BuildWithErrorOnRun(t *testing.T) {
	project, err := devstep.NewProject(&devstep.ProjectConfig{
		BaseImage: "repo/name:tag",
		HostDir:   "/path/on/host",
		GuestDir:  "/path/on/guest",
		CacheDir:  "/cache/path/on/host",
	})
	ok(t, err)

	clientMock := NewMockClient()
	runError := errors.New("BOOM!")
	clientMock.RunFunc = func(o *devstep.DockerRunOpts) (*devstep.DockerRunResult, error) {
		return nil, runError
	}

	err = project.Build(clientMock)
	equals(t, runError, err)
}

func Test_BuildWithErrorOnCommit(t *testing.T) {
	project, err := devstep.NewProject(&devstep.ProjectConfig{
		BaseImage: "repo/name:tag",
		HostDir:   "/path/on/host",
		GuestDir:  "/path/on/guest",
		CacheDir:  "/cache/path/on/host",
	})
	ok(t, err)

	clientMock := NewMockClient()
	clientMock.CommitFunc = func(o *devstep.DockerCommitOpts) error {
		return errors.New("BOOM!")
	}

	err = project.Build(clientMock)
	assert(t, err != nil, "No error raised")
	assert(t, strings.HasPrefix(err.Error(), "Error commiting"), "Wrong message")
}

func Test_BuildWithBadExitCode(t *testing.T) {
	project, err := devstep.NewProject(&devstep.ProjectConfig{
		BaseImage: "repo/name:tag",
		HostDir:   "/path/on/host",
		GuestDir:  "/path/on/guest",
		CacheDir:  "/cache/path/on/host",
	})
	ok(t, err)

	clientMock := NewMockClient()
	clientMock.RunFunc = func(o *devstep.DockerRunOpts) (*devstep.DockerRunResult, error) {
		return &devstep.DockerRunResult{ExitCode: 1}, nil
	}

	err = project.Build(clientMock)
	assert(t, err != nil, "Did not error")
}

func Test_Clean(t *testing.T) {
	project, err := devstep.NewProject(&devstep.ProjectConfig{
		RepositoryName: "my/project",
	})
	ok(t, err)

	tags := []string{"a-tag", "other-tag"}
	clientMock := NewMockClient()

	var repositoryNameSearched string
	clientMock.ListTagsFunc = func(r string) ([]string, error) {
		repositoryNameSearched = r
		return tags, nil
	}

	removedImages := []string{}
	clientMock.RemoveImageFunc = func(t string) error {
		removedImages = append(removedImages, t)
		return nil
	}

	err = project.Clean(clientMock)
	ok(t, err)

	equals(t, "my/project", repositoryNameSearched)
	equals(t, 2, len(removedImages))
}

func inArray(str string, array []string) bool {
	for index := range array {
		if str == array[index] {
			return true
		}
	}
	return false
}
