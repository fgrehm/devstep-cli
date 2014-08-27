package devstep

import (
	"errors"
	"github.com/fgrehm/go-dockerpty"
	"github.com/fsouza/go-dockerclient"
	"os"
	"strings"
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
	Workdir    string
	Env        map[string]string
	Volumes    []string
	Links      []string
	Image      string
	Cmd        []string
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

func (c *dockerClient) Run(opts *DockerRunOpts) (*DockerRunResult, error) {
	createOpts := docker.CreateContainerOptions{
		Config: &docker.Config{
			Image:        opts.Image,
			Cmd:          opts.Cmd,
			OpenStdin:    opts.Pty,
			StdinOnce:    opts.Pty,
			AttachStdin:  opts.Pty,
			AttachStdout: true,
			AttachStderr: true,
			Tty:          opts.Pty,
			WorkingDir:   opts.Workdir,
		},
	}
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

	hostConfig := &docker.HostConfig{Binds: opts.Volumes}
	log.Debug("HostConfig: %+v", hostConfig)

	if opts.Pty {
		log.Info("Starting container with pseudo terminal")
		err = dockerpty.Start(c.client, container, hostConfig)
	} else {
		err = c.client.StartContainer(container.ID, hostConfig)

		if err == nil {
			// Attach to the container
			attachOpts := docker.AttachToContainerOptions{
				Container:    container.ID,
				OutputStream: os.Stdout,
				ErrorStream:  os.Stderr,
				Stdin:        opts.Pty,
				Stdout:       true,
				Stderr:       true,
				Stream:       true,
				RawTerminal:  opts.Pty,
			}

			if opts.Pty {
				attachOpts.InputStream = os.Stdin
			}

			log.Info("Attaching to container '%s'", container.ID)
			log.Debug("Attach options: %+v", attachOpts)
			err = c.client.AttachToContainer(attachOpts)
		}
	}
	if err != nil {
		return nil, errors.New("Error starting container:\n  " + err.Error())
	}

	container, err = c.client.InspectContainer(container.ID)
	if err != nil {
		return nil, errors.New("Error inspecting container:\n  " + err.Error())
	}

	return &DockerRunResult{
		ContainerID: container.ID,
		ExitCode:  container.State.ExitCode,
	}, nil
}

func (c *dockerClient) RemoveContainer(containerID string) error {
	log.Info("Removing container '%s'", containerID)
	return c.client.RemoveContainer(docker.RemoveContainerOptions{
		ID:            containerID,
		Force:         true,
		RemoveVolumes: true,
	})
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

	apiImages, err := c.client.ListImages(false)
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

func NewClient(endpoint string) DockerClient {
	// TODO: Error handling
	innerClient, _ := docker.NewClient(endpoint)
	return &dockerClient{innerClient}
}
