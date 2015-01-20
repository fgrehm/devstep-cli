package devstep

// TODO: Figure out how this can be unit tested

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
	Name       string
	AutoRemove bool
	Pty        bool
	Workdir    string
	Hostname   string
	Privileged *bool
	Env        map[string]string
	Volumes    []string
	Links      []string
	Image      string
	Cmd        []string
	Publish    []string
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

	hostConfig := opts.toHostConfig()
	log.Debug("HostConfig: %+v", hostConfig)

	if opts.Pty {
		log.Info("Starting container with pseudo terminal")
		err = dockerpty.Start(c.client, container, hostConfig)
	} else {
		return nil, errors.New("Starting containers without Pty was not needed until this moment, please implement :)")
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

func NewClient(endpoint string) DockerClient {
	// TODO: Error handling
	innerClient, _ := docker.NewClient(endpoint)
	return &dockerClient{innerClient}
}

func (this DockerRunOpts) Merge(others ...*DockerRunOpts) *DockerRunOpts {
	for _, other := range others {
		if other == nil {
			continue
		}

		if other.Name != "" {
			this.Name = other.Name
		}
		if other.Image != "" {
			this.Image = other.Image
		}
		if len(other.Cmd) > 0 {
			this.Cmd = other.Cmd
		}

		this.AutoRemove = other.AutoRemove
		this.Pty = other.Pty

		if other.Workdir != "" {
			this.Workdir = other.Workdir
		}

		if other.Privileged != nil {
			this.Privileged = other.Privileged
		}

		this.Publish = append(this.Publish, other.Publish...)
		this.Volumes = append(this.Volumes, other.Volumes...)
		this.Links = append(this.Links, other.Links...)
		for k, v := range other.Env {
			this.Env[k] = v
		}
	}
	return &this
}

func (opts *DockerRunOpts) toCreateOpts() docker.CreateContainerOptions {
	env := []string{}
	for k, v := range opts.Env {
		env = append(env, k+"="+v)
	}

	exposedPorts := make(map[docker.Port]struct{})
	for _, port := range opts.Publish {
		containerPort := strings.Split(port, ":")[1]
		exposedPorts[docker.Port(containerPort)] = struct{}{}
	}

	return docker.CreateContainerOptions{
		Name: opts.Name,
		Config: &docker.Config{
			Image:        opts.Image,
			Cmd:          opts.Cmd,
			Env:          env,
			Hostname:     opts.Hostname,
			ExposedPorts: exposedPorts,
			OpenStdin:    opts.Pty,
			StdinOnce:    opts.Pty,
			AttachStdin:  opts.Pty,
			AttachStdout: true,
			AttachStderr: true,
			Tty:          opts.Pty,
			WorkingDir:   opts.Workdir,
		},
	}
}

func (opts *DockerRunOpts) toHostConfig() *docker.HostConfig {
	portBindings := make(map[docker.Port][]docker.PortBinding)
	for _, port := range opts.Publish {
		hostAndContainerPorts := strings.Split(port, ":")
		hostPort := hostAndContainerPorts[0]
		containerPort := docker.Port(hostAndContainerPorts[1])

		bindings := portBindings[containerPort]
		portBindings[containerPort] = append(bindings, docker.PortBinding{HostPort: hostPort})
	}

	privileged := false
	if opts.Privileged != nil {
		privileged = *opts.Privileged
	}

	return &docker.HostConfig{
		Binds:        opts.Volumes,
		Links:        opts.Links,
		Privileged:   privileged,
		PortBindings: portBindings,
	}
}

func (opts *DockerRunOpts) toAttachOpts(containerID string) docker.AttachToContainerOptions {
	attachOpts := docker.AttachToContainerOptions{
		Container:    containerID,
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

	return attachOpts
}
