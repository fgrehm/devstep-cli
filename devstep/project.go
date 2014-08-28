package devstep

import (
	"errors"
	"fmt"
	"time"
)

// The project interface provides access to the configuration, state and
// lifecycle of a Project.
type Project interface {
	Build(DockerClient) error
	Clean(DockerClient) error
	Hack(DockerClient) error
}

// Project specific configuration, usually parsed from an yaml file
type ProjectConfig struct {
	SourceImage    string         // image used when starting environments from scratch
	BaseImage      string         // starting point for the project
	RepositoryName string         // name of the docker repository this project should be commited
	HostDir        string         // root directory of the project on the host machine
	GuestDir       string         // directory where the project sources will be mounted on the container
	CacheDir       string         // a directory on the host machine were we can place downloaded packages
	Defaults       *DockerRunOpts // default options passed on to docker for all commands
	HackOpts       *DockerRunOpts // `devstep hack` specific options passed to the container
}

// An implementation of a Project.
type project struct {
	*ProjectConfig
}

// This creates a new project
func NewProject(config *ProjectConfig) (Project, error) {
	// TODO: This seems a bit weird
	project := &project{config}
	return project, nil
}

// Build the project and commit it to an image
func (p *project) Build(client DockerClient) error {
	volumes := []string{
		p.HostDir + ":" + p.GuestDir,
		"/tmp/devstep/cache:/.devstep/cache",
	}

	fmt.Printf("==> Building project from '%s'\n", p.BaseImage)

	result, err := client.Run(&DockerRunOpts{
		Image:      p.BaseImage,
		AutoRemove: false,
		Pty:        true,
		Cmd:        []string{"/.devstep/bin/build-project", p.GuestDir},
		Volumes:    volumes,
		Workdir:    p.GuestDir,
	})
	log.Debug("Docker run result: %+v", result)

	if err != nil {
		return err
	}

	if result.ExitCode != 0 {
		return errors.New("Container exited with status != 0")
	}

	tag := "latest"
	fmt.Printf("==> Commiting container to '%s:%s'\n", p.RepositoryName, tag)
	err = client.Commit(&DockerCommitOpts{
		ContainerID:    result.ContainerID,
		RepositoryName: p.RepositoryName,
		Tag:            tag,
	})
	if err != nil {
		return errors.New("Error commiting container:\n  " + err.Error())
	}

	tag = time.Now().Local().Format("20060102150405")
	fmt.Printf("==> Commiting container to '%s:%s'\n", p.RepositoryName, tag)
	err = client.Commit(&DockerCommitOpts{
		ContainerID:    result.ContainerID,
		RepositoryName: p.RepositoryName,
		Tag:            tag,
	})
	if err != nil {
		return errors.New("Error commiting container:\n  " + err.Error())
	}

	fmt.Println("==> Removing container used for build")
	return client.RemoveContainer(result.ContainerID)
}

// Starts a hacking session on the project
func (p *project) Hack(client DockerClient) error {
	volumes := []string{
		p.HostDir + ":" + p.GuestDir,
		"/tmp/devstep/cache:/.devstep/cache",
	}

	fmt.Printf("==> Creating container using '%s'\n", p.BaseImage)

	_, err := client.Run(&DockerRunOpts{
		Image:      p.BaseImage,
		AutoRemove: true,
		Pty:        true,
		Cmd:        []string{"/.devstep/bin/hack"},
		Volumes:    volumes,
		Workdir:    p.GuestDir,
	})
	return err
}

func (p *project) Clean(client DockerClient) error {
	fmt.Printf("==> Removing tags for '%s'\n", p.RepositoryName)

	tags, err := client.ListTags(p.RepositoryName)
	if err != nil {
		return err
	}

	for _, tag := range tags {
		image := p.RepositoryName + ":" + tag
		if err = client.RemoveImage(image); err != nil {
			return err
		}
	}

	return nil
}
