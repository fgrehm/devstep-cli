package devstep_test

import (
	"github.com/fgrehm/devstep-cli/devstep"
)

type MockClient struct {
	RunFunc             func(*devstep.DockerRunOpts) (*devstep.DockerRunResult, error)
	RemoveContainerFunc func(string) error
	CommitFunc          func(*devstep.DockerCommitOpts) error
	RemoveImageFunc     func(string) error
	ListTagsFunc        func(string) ([]string, error)
}

func (c *MockClient) Run(runOpts *devstep.DockerRunOpts) (*devstep.DockerRunResult, error) {
	return c.RunFunc(runOpts)
}

func (c *MockClient) RemoveContainer(containerID string) error {
	return c.RemoveContainerFunc(containerID)
}

func (c *MockClient) Commit(commitOpts *devstep.DockerCommitOpts) error {
	return c.CommitFunc(commitOpts)
}

func (c *MockClient) RemoveImage(imageName string) error {
	return c.RemoveImageFunc(imageName)
}

func (c *MockClient) ListTags(repositoryName string) ([]string, error) {
	return c.ListTagsFunc(repositoryName)
}

func NewMockClient() *MockClient {
	return &MockClient{}
}
