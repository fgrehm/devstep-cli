package devstep

// TODO: Figure out how this can be unit tested

import (
	"errors"
	"github.com/fgrehm/go-dockerpty"
	"github.com/fsouza/go-dockerclient"
	"strings"
)

type DockerClient interface {
	Execute(*DockerExecOpts) error
	Run(*DockerRunOpts) (*DockerRunResult, error)
	RemoveContainer(string) error
	ContainerChanged(string) (bool, error)
	ContainerHasExecInstancesRunning(string) bool
	Commit(*DockerCommitOpts) error
	RemoveImage(string) error
	ListTags(string) ([]string, error)
	ListContainers(string) ([]string, error)
	LookupContainerID(string) (string, error)
}

type DockerExecOpts struct {
	ContainerID string
	User        string
	Cmd         []string
}

type DockerRunResult struct {
	ContainerID string
	ExitCode    int
}

type DockerCommitOpts struct {
	ContainerID    string
	RepositoryName string
	Tag            string
}

type dockerClient struct {
	client *docker.Client
}

func (c *dockerClient) Execute(opts *DockerExecOpts) error {
	log.Info("Creating exec instance")
	log.Debug("%+v", opts)

	exec, err := c.client.CreateExec(docker.CreateExecOptions{
		Container:    opts.ContainerID,
		User:         opts.User,
		AttachStdin:  true,
		AttachStdout: true,
		AttachStderr: true,
		Tty:          true,
		Cmd:          opts.Cmd,
	})

	if err != nil {
		return err
	}

	log.Info("Starting Exec instance with pseudo terminal")
	return dockerpty.StartExec(c.client, exec)
}

func (c *dockerClient) Run(opts *DockerRunOpts) (*DockerRunResult, error) {
	createOpts := opts.toCreateOpts()

	log.Info("Creating container")
	log.Debug("%+v", createOpts.Config)
	container, err := c.client.CreateContainer(createOpts)
	if err != nil {
		return nil, errors.New("Error creating container: \n  " + err.Error())
	}
	log.Info("Container created (ID='%s')", container.ID)

	if opts.AutoRemove {
		defer c.RemoveContainer(container.ID)
	}

	if opts.Detach {
		err = c.client.StartContainer(container.ID, &docker.HostConfig{})
	} else if opts.Pty {
		log.Info("Starting container with pseudo terminal")
		err = dockerpty.Start(c.client, container, &docker.HostConfig{})
	} else {
		return nil, errors.New("Starting daemon containers without Pty was not needed until this moment, please implement :)")
	}

	if err != nil {
		return nil, errors.New("Error starting container:\n  " + err.Error())
	}

	container, err = c.client.InspectContainer(container.ID)
	if err != nil {
		return nil, errors.New("Error inspecting container:\n  " + err.Error())
	}
	result := &DockerRunResult{
		ContainerID: container.ID,
		ExitCode:    container.State.ExitCode,
	}

	return result, nil
}

func (c *dockerClient) RemoveContainer(containerID string) error {
	log.Info("Removing container '%s'", containerID)
	return c.client.RemoveContainer(docker.RemoveContainerOptions{
		ID:            containerID,
		Force:         true,
		RemoveVolumes: true,
	})
}

func (c *dockerClient) ContainerChanged(containerID string) (bool, error) {
	changes, err := c.client.ContainerChanges(containerID)
	log.Debug("Container changes '%v'", changes)

	if len(changes) == 0 {
		return false, err
	}

	blankChanges := []docker.Change{
		docker.Change{
			Path: "/home",
			Kind: docker.ChangeModify,
		},
		docker.Change{
			Path: "/home/devstep",
			Kind: docker.ChangeModify,
		},
		docker.Change{
			Path: "/home/devstep/.rnd",
			Kind: docker.ChangeModify,
		},
	}
	for i, v := range blankChanges {
		if v != changes[i] {
			return true, err
		}
	}
	return len(changes) > 3, err
}

func (c *dockerClient) Commit(opts *DockerCommitOpts) error {
	_, err := c.client.CommitContainer(docker.CommitContainerOptions{
		Container:  opts.ContainerID,
		Repository: opts.RepositoryName,
		Tag:        opts.Tag,
	})

	return err
}

func (c *dockerClient) RemoveImage(name string) error {
	log.Info("Removing image '%s'", name)
	return c.client.RemoveImage(name)
}

// List tags for a given repository
func (c *dockerClient) ListTags(repositoryName string) ([]string, error) {
	if repositoryName == "" {
		return nil, errors.New("Repository name can't be blank")
	}

	log.Debug("Fetching tags for '%s'", repositoryName)

	// TODO: Use go-dockerclient's support for filtering images
	apiImages, err := c.client.ListImages(docker.ListImagesOptions{All: true})
	tags := []string{}
	for _, img := range apiImages {
		for _, repoTag := range img.RepoTags {
			repositoryAndTag := strings.Split(repoTag, ":")
			if repositoryAndTag[0] == repositoryName {
				tags = append(tags, repositoryAndTag[1])
			}
		}
	}

	log.Debug("Tags found %s", tags)

	return tags, err
}

func (c *dockerClient) ContainerHasExecInstancesRunning(containerID string) bool {
	container, err := c.client.InspectContainer(containerID)
	if err != nil {
		panic(err)
	}
	log.Debug("ExecIDs %+v", container.ExecIDs)
	for _, execID := range container.ExecIDs {
		execInspect, err := c.client.InspectExec(execID)
		if err != nil {
			panic(err)
		}
		log.Debug("execInspect ID=%s RUNNING=%+v", execInspect.ID, execInspect.Running)
		if execInspect.Running {
			return true
		}
	}
	return false
}

// List Containers for a given image
func (c *dockerClient) ListContainers(image string) ([]string, error) {
	if image == "" {
		return nil, errors.New("Image name can't be blank")
	}

	log.Info("Fetching containers for '%s'", image)

	containers, err := c.client.ListContainers(docker.ListContainersOptions{
		Filters: map[string][]string{"status": []string{"running"}},
	})
	containerIds := []string{}
	for _, container := range containers {
		log.Debug("Found '%+v'", container)
		if container.Image == image {
			containerIds = append(containerIds, container.ID)
		}
	}

	log.Info("Containers found %v", containerIds)

	return containerIds, err
}

func (c *dockerClient) LookupContainerID(containerName string) (string, error) {
	container, err := c.client.InspectContainer(containerName)
	if err != nil {
		return "", errors.New("Error inspecting container:\n  " + err.Error())
	}
	return container.Name, nil
}

func NewClient() DockerClient {
	// TODO: Error handling
	innerClient, _ := docker.NewClientFromEnv()
	return &dockerClient{innerClient}
}
