package devstep

import (
	"github.com/fsouza/go-dockerclient"
	"os"
	"strings"
)

type DockerRunOpts struct {
	Name       string
	Detach     bool
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

		if other.Detach {
			this.Detach = true
		}

		this.Publish = append(this.Publish, other.Publish...)
		this.Volumes = append(this.Volumes, other.Volumes...)
		this.Links = append(this.Links, other.Links...)
		if this.Env == nil {
			this.Env = map[string]string{}
		}
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

    hostConfig := opts.toHostConfig()
    log.Debug("HostConfig: %+v", hostConfig)

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
			AttachStdout: opts.Pty,
			AttachStderr: opts.Pty,
			Tty:          opts.Pty,
			WorkingDir:   opts.Workdir,
		},
        HostConfig: hostConfig,
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
