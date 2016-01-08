package devstep_test

import (
	"github.com/fgrehm/devstep-cli/devstep"
)

type MockClient struct {
	ExecuteFunc                          func(*devstep.DockerExecOpts) error
	RunFunc                              func(*devstep.DockerRunOpts) (*devstep.DockerRunResult, error)
	RemoveContainerFunc                  func(string) error
	ContainerChangedFunc                 func(string) (bool, error)
	ContainerHasExecInstancesRunningFunc func(string) bool
	CommitFunc                           func(*devstep.DockerCommitOpts) error
	RemoveImageFunc                      func(string) error
	ListTagsFunc                         func(string) ([]string, error)
	ListContainersFunc                   func(string) ([]string, error)
	LookupContainerIDFunc                func(string) (string, error)
}

func (c *MockClient) Execute(execOpts *devstep.DockerExecOpts) error {
	return c.ExecuteFunc(execOpts)
}

func (c *MockClient) Run(runOpts *devstep.DockerRunOpts) (*devstep.DockerRunResult, error) {
	return c.RunFunc(runOpts)
}

func (c *MockClient) RemoveContainer(containerID string) error {
	return c.RemoveContainerFunc(containerID)
}

func (c *MockClient) ContainerChanged(containerID string) (bool, error) {
	return c.ContainerChangedFunc(containerID)
}

func (c *MockClient) ContainerHasExecInstancesRunning(containerID string) bool {
	return c.ContainerHasExecInstancesRunningFunc(containerID)
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

func (c *MockClient) ListContainers(repositoryName string) ([]string, error) {
	return c.ListContainersFunc(repositoryName)
}

func (c *MockClient) LookupContainerID(containerName string) (string, error) {
	return c.LookupContainerIDFunc(containerName)
}

func NewMockClient() *MockClient {
	return &MockClient{
		ListTagsFunc: func(repositoryName string) ([]string, error) {
			return []string{}, nil
		},
		ListContainersFunc: func(repositoryName string) ([]string, error) {
			return []string{}, nil
		},
		RunFunc: func(runOpts *devstep.DockerRunOpts) (*devstep.DockerRunResult, error) {
			return &devstep.DockerRunResult{}, nil
		},
		ContainerChangedFunc: func(containerID string) (bool, error) {
			return true, nil
		},
		CommitFunc: func(commitOpts *devstep.DockerCommitOpts) error {
			return nil
		},
		RemoveContainerFunc: func(containerID string) error {
			return nil
		},
	}
}
