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
		return &devstep.DockerRunResult{ContainerID: "cid", ExitCode: 0}, nil
	}
	clientMock.ListContainersFunc = func(string) ([]string, error) {
		return []string{}, nil
	}
	clientMock.ContainerHasExecInstancesRunningFunc = func(string) bool {
		return false
	}

	err = project.Hack(clientMock, &devstep.DockerRunOpts{
		Links:   []string{"foo:bar", "bar:foo"},
		Publish: []string{"1:2", "3:4"},
	})
	ok(t, err)

	equals(t, "repo/name:tag", runOpts.Image)
	assert(t, runOpts.Pty, "Pseudo tty allocation is disabled")
	equals(t, []string{"--"}, runOpts.Cmd)
	equals(t, "/path/on/guest", runOpts.Workdir)

	assert(t, inArray("/path/on/host:/path/on/guest", runOpts.Volumes), "Project dir was not shared")
	assert(t, inArray("/cache/path/on/host:/home/devstep/cache", runOpts.Volumes), "Cache dir was not shared")

	assert(t, inArray("foo:bar", runOpts.Links), "Link was not shared")
	assert(t, inArray("bar:foo", runOpts.Links), "Link was not shared")

	assert(t, inArray("1:2", runOpts.Publish), "CLI publish argument was not shared")
	assert(t, inArray("3:4", runOpts.Publish), "CLI publish argument was not shared")
}

func Test_HackUsesDockerConfigs(t *testing.T) {
	var privileged *bool
	{
		t := true
		privileged = &t
	}

	project, err := devstep.NewProject(&devstep.ProjectConfig{
		HostDir:  "/path/on/host",
		GuestDir: "/path/on/guest",
		CacheDir: "/cache/path/on/host",
		Defaults: &devstep.DockerRunOpts{
			Privileged: privileged,
			Links:      []string{"some:link"},
			Volumes:    []string{"/some:/volume"},
			Env:        map[string]string{"SOME": "ENV"},
		},
		HackOpts: &devstep.DockerRunOpts{
			Links:   []string{"other:link"},
			Volumes: []string{"/other:/volume"},
			Env:     map[string]string{"OTHER": "VALUE"},
		},
	})
	ok(t, err)

	var runOpts *devstep.DockerRunOpts
	clientMock := NewMockClient()
	clientMock.RunFunc = func(o *devstep.DockerRunOpts) (*devstep.DockerRunResult, error) {
		runOpts = o
		return nil, nil
	}

	err = project.Hack(clientMock, nil)
	ok(t, err)

	assert(t, *runOpts.Privileged, "Privileged is false")

	assert(t, inArray("/path/on/host:/path/on/guest", runOpts.Volumes), "Project dir was not shared")
	assert(t, inArray("/cache/path/on/host:/home/devstep/cache", runOpts.Volumes), "Cache dir was not shared")

	assert(t, inArray("/some:/volume", runOpts.Volumes), "Default volumes were not set")
	assert(t, inArray("some:link", runOpts.Links), "Default links were not set")
	equals(t, "ENV", runOpts.Env["SOME"])

	assert(t, inArray("/other:/volume", runOpts.Volumes), "Hack volumes were not set")
	assert(t, inArray("other:link", runOpts.Links), "Hack links were not set")
	equals(t, "VALUE", runOpts.Env["OTHER"])
}

func Test_Build(t *testing.T) {
	project, err := devstep.NewProject(&devstep.ProjectConfig{
		BaseImage:      "repo/name:tag",
		HostDir:        "/path/on/host",
		GuestDir:       "/path/on/guest",
		CacheDir:       "/cache/path/on/host",
		RepositoryName: "repo-name",
	})
	ok(t, err)

	var runOpts *devstep.DockerRunOpts
	clientMock := NewMockClient()
	clientMock.RunFunc = func(o *devstep.DockerRunOpts) (*devstep.DockerRunResult, error) {
		runOpts = o
		return &devstep.DockerRunResult{
			ExitCode:    0,
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

	err = project.Build(clientMock, &devstep.DockerRunOpts{})
	ok(t, err)

	equals(t, "repo/name:tag", runOpts.Image)
	assert(t, !runOpts.AutoRemove, "AutoRemove is true")
	assert(t, runOpts.Pty, "Pseudo tty allocation is disabled")
	equals(t, []string{"/opt/devstep/bin/build-project", "/path/on/guest"}, runOpts.Cmd)
	equals(t, "/path/on/guest", runOpts.Workdir)
	assert(t, inArray("/path/on/host:/path/on/guest", runOpts.Volumes), "Project dir was not shared")
	assert(t, inArray("/cache/path/on/host:/home/devstep/cache", runOpts.Volumes), "Cache dir was not shared")

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

func Test_BuildUsesGlobalDockerConfigs(t *testing.T) {
	var privilegedTrue *bool
	{
		t := true
		privilegedTrue = &t
	}
	var privilegedFalse *bool
	{
		f := false
		privilegedFalse = &f
	}

	project, err := devstep.NewProject(&devstep.ProjectConfig{
		HostDir:  "/path/on/host",
		GuestDir: "/path/on/guest",
		CacheDir: "/cache/path/on/host",
		Defaults: &devstep.DockerRunOpts{
			Privileged: privilegedTrue,
			Links:      []string{"some:link"},
			Volumes:    []string{"/some:/volume"},
			Env:        map[string]string{"SOME": "ENV"},
		},
		HackOpts: &devstep.DockerRunOpts{
			Privileged: privilegedFalse,
			Links:      []string{"other:link"},
			Volumes:    []string{"/other:/volume"},
			Env:        map[string]string{"OTHER": "VALUE"},
		},
	})
	ok(t, err)

	var runOpts *devstep.DockerRunOpts
	clientMock := NewMockClient()
	clientMock.RunFunc = func(o *devstep.DockerRunOpts) (*devstep.DockerRunResult, error) {
		runOpts = o
		return &devstep.DockerRunResult{
			ExitCode:    0,
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

	err = project.Build(clientMock, &devstep.DockerRunOpts{})
	ok(t, err)

	assert(t, *runOpts.Privileged, "Privileged is set to false")
	assert(t, inArray("/path/on/host:/path/on/guest", runOpts.Volumes), "Project dir was not shared")
	assert(t, inArray("/cache/path/on/host:/home/devstep/cache", runOpts.Volumes), "Cache dir was not shared")

	assert(t, inArray("/some:/volume", runOpts.Volumes), "Default volumes were not set")
	assert(t, inArray("some:link", runOpts.Links), "Default links were not set")
	equals(t, "ENV", runOpts.Env["SOME"])

	assert(t, !inArray("/other:/volume", runOpts.Volumes), "Hack volumes were set")
	assert(t, !inArray("other:link", runOpts.Links), "Hack links were set")
	equals(t, "", runOpts.Env["OTHER"])
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

	// TODO: Ensure the container gets removed

	err = project.Build(clientMock, &devstep.DockerRunOpts{})
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

	err = project.Build(clientMock, &devstep.DockerRunOpts{})
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

	// TODO: Ensure the container gets removed

	err = project.Build(clientMock, &devstep.DockerRunOpts{})
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
