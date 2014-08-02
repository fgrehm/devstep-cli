package devstep

import (
	"errors"
	"github.com/fsouza/go-dockerclient"
)

type DockerClient interface {
	Run(*DockerRunOpts) (*DockerRunResult, error)
	RemoveContainer(string) error
	Commit(*DockerCommitOpts) error
	RemoveImage(string) error
	ListTags(string) ([]string, error)
}

type DockerRunOpts struct {
	AutoRemove bool
	Pty        bool
	Env        map[string]string
	Volumes    []string
	Links      []string
	Image      string
	Cmd        []string
}
type DockerRunResult struct {
	ContainerID string
	ExitStatus  int8
}

type DockerCommitOpts struct {
	ContainerID string
	ImageName   string
}

type dockerClient struct {
	client *docker.Client
}

func (*dockerClient) Run(*DockerRunOpts) (*DockerRunResult, error) {
	return nil, errors.New("NotImplemented")
}

func (*dockerClient) RemoveContainer(string) error {
	return errors.New("NotImplemented")
}

func (*dockerClient) Commit(*DockerCommitOpts) error {
	return errors.New("NotImplemented")
}

func (*dockerClient) RemoveImage(string) error {
	return errors.New("NotImplemented")
}

func (*dockerClient) ListTags(string) ([]string, error) {
	return nil, errors.New("NotImplemented")
}

func NewClient(endpoint string) DockerClient {
	// TODO: Error handling
	innerClient, _ := docker.NewClient(endpoint)
	return &dockerClient{innerClient}
}
