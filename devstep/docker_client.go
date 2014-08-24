package devstep

import (
	"errors"
	"fmt"
	"github.com/fgrehm/go-dockerpty"
	"github.com/fsouza/go-dockerclient"
	"os"
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
	ExitStatus  int8
}

type DockerCommitOpts struct {
	ContainerID string
	ImageName   string
}

type dockerClient struct {
	client *docker.Client
}

func (c *dockerClient) Run(opts *DockerRunOpts) (*DockerRunResult, error) {
	// fmt.Println(opts)

	container, err := c.client.CreateContainer(docker.CreateContainerOptions{
		Name: "testing-new-devstep-cli",
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
	})

	if err != nil {
		fmt.Println("Error when creating container:")
		fmt.Println(err)
		os.Exit(1)
	}

	defer func() {
		c.client.RemoveContainer(docker.RemoveContainerOptions{
			ID:            container.ID,
			Force:         true,
			RemoveVolumes: true,
		})
	}()

	hostConfig := &docker.HostConfig{Binds: opts.Volumes}

	if opts.Pty {
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

			err = c.client.AttachToContainer(attachOpts)
		}
	}

	if err != nil {
		fmt.Println("Error when starting container:")
		fmt.Println(err)
		os.Exit(1)
	}

	return nil, nil
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
