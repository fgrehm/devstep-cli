package devstep

import (
	"errors"
	"fmt"
	"time"

	"github.com/kardianos/osext"
)

// The project interface provides access to the configuration, state and
// lifecycle of a Project.
type Project interface {
	Config() *ProjectConfig
	Build(DockerClient, *DockerRunOpts) error
	Commit(DockerClient, string) error
	Bootstrap(DockerClient, *DockerRunOpts) error
	Clean(DockerClient) error
	Hack(DockerClient, *DockerRunOpts) error
	Run(DockerClient, *DockerRunOpts) (*DockerRunResult, error)
	Exec(DockerClient, []string) error
}

// Project specific configuration, usually parsed from an yaml file
type ProjectConfig struct {
	SourceImage    string                     // image used when starting environments from scratch
	BaseImage      string                     // starting point for the project
	RepositoryName string                     // name of the docker repository this project should be commited
	HostDir        string                     // root directory of the project on the host machine
	GuestDir       string                     // directory where the project sources will be mounted on the container
	CacheDir       string                     // a directory on the host machine were we can place downloaded packages
	Defaults       *DockerRunOpts             // default options passed on to docker for all commands
	HackOpts       *DockerRunOpts             // `devstep hack` specific options passed to the container
	Commands       map[string]*ProjectCommand // shortcut for commands
}

type ProjectCommand struct {
	Name string
	DockerRunOpts
}

// An implementation of a Project.
type project struct {
	*ProjectConfig
}

// This creates a new project
func NewProject(config *ProjectConfig) (Project, error) {
	project := &project{config}
	if project.Defaults == nil {
		project.Defaults = &DockerRunOpts{Env: make(map[string]string)}
	}
	if project.HackOpts == nil {
		project.HackOpts = &DockerRunOpts{Env: make(map[string]string)}
	}
	return project, nil
}

func (p *project) Config() *ProjectConfig {
	return p.ProjectConfig
}

// Build the project and commit it to an image
func (p *project) Build(client DockerClient, cliOpts *DockerRunOpts) error {
	fmt.Printf("==> Building project from '%s'\n", p.BaseImage)

	result, err := p.buildWithCommand(client, cliOpts, []string{"/opt/devstep/bin/build-project", p.GuestDir})
	if err != nil {
		return err
	}

	fmt.Println("==> Removing container used for build")
	return client.RemoveContainer(result.ContainerID)
}

// Commit a container that is running
func (p *project) Commit(client DockerClient, containerName string) error {
	containerID, err := client.LookupContainerID(containerName)

	if err != nil {
		return err
	}

	if err = p.commit(client, containerID, "latest"); err != nil {
		return err
	}

	tag := time.Now().Local().Format("20060102150405")
	return p.commit(client, containerID, tag)
}

// Start a hacking session and commit it to an image if all goes well
func (p *project) Bootstrap(client DockerClient, cliOpts *DockerRunOpts) error {
	fmt.Printf("==> Creating container based on '%s'\n", p.BaseImage)

	result, err := p.buildWithCommand(client, cliOpts, []string{"bash"})
	if err != nil {
		return err
	}

	fmt.Println("==> Removing container used for bootstrapping")
	return client.RemoveContainer(result.ContainerID)
}

// Starts a hacking session on the project
func (p *project) Hack(client DockerClient, cliHackOpts *DockerRunOpts) error {
	if p.SourceImage != p.BaseImage {
		containerID := ""

		containers, err := client.ListContainers(p.BaseImage)
		if err != nil {
			return err
		}

		if len(containers) == 0 {
			log.Debug("==> No containers have been created for '%s', will start a new one\n", p.BaseImage)

			result, err := p.startContainer(client, cliHackOpts)
			if err != nil {
				return err
			}

			containerID = result.ContainerID

			log.Debug("STARTED: %+v", result)

			err = p.Exec(client, []string{"/opt/devstep/bin/hack"})
		} else {
			containerID = containers[0]
			err = p.Exec(client, []string{"bash"})
		}

		if err != nil {
			// Check how many exec instances we have in place and remove the container
			// if it is the last one
			if !client.ContainerHasExecInstancesRunning(containerID) {
				fmt.Printf("Removing container: %+v\n", containerID)
				client.RemoveContainer(containerID)
			} else {
				fmt.Printf("Skipping container removal: %s\n", containerID)
			}
			return err
		}

		// Check how many exec instances we have in place and remove the container
		// if it is the last one
		if !client.ContainerHasExecInstancesRunning(containerID) {
			fmt.Printf("Removing container: %+v\n", containerID)
			client.RemoveContainer(containerID)
		} else {
			fmt.Printf("Skipping container removal: %s\n", containerID)
		}

		return err

	} else {
		opts := p.HackOpts.Merge(cliHackOpts, &DockerRunOpts{
			Cmd: []string{"/opt/devstep/bin/hack"},
		})

		_, err := p.Run(client, opts)

		return err
	}
}

func (p *project) Run(client DockerClient, cliRunOpts *DockerRunOpts) (*DockerRunResult, error) {
	opts := p.Defaults.Merge(cliRunOpts, &DockerRunOpts{
		Image:      p.BaseImage,
		AutoRemove: true,
		Pty:        true,
		Workdir:    p.GuestDir,
		Volumes: []string{
			p.HostDir + ":" + p.GuestDir,
			p.CacheDir + ":/home/devstep/cache",
		},
	})

	fmt.Printf("==> Creating container using '%s'\n", p.BaseImage)

	return client.Run(opts)
}

func (p *project) Exec(client DockerClient, cmd []string) error {
	containers, err := client.ListContainers(p.BaseImage)
	if err != nil {
		return err
	}

	if len(containers) == 0 {
		return errors.New("No containers found to execute the command.")
	}

	cmd = append([]string{"/opt/devstep/bin/exec-entrypoint"}, cmd...)

	log.Debug("==> Executing %v on '%s'\n", cmd, containers[0])
	return client.Execute(&DockerExecOpts{
		ContainerID: containers[0],
		Cmd:         cmd,
		User:        "developer",
	})
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

func (p *project) commit(client DockerClient, containerID, tag string) error {
	fmt.Printf("==> Commiting container to '%s:%s'\n", p.RepositoryName, tag)
	err := client.Commit(&DockerCommitOpts{
		ContainerID:    containerID,
		RepositoryName: p.RepositoryName,
		Tag:            tag,
	})
	if err != nil {
		return errors.New("Error commiting container:\n  " + err.Error())
	}

	return nil
}

func (p *project) startContainer(client DockerClient, cliOpts *DockerRunOpts) (*DockerRunResult, error) {
	executable, err := osext.Executable()
	if err != nil {
		return nil, err
	}

	opts := p.Defaults.Merge(cliOpts, &DockerRunOpts{
		Image:      p.BaseImage,
		Detach:     true,
		AutoRemove: false,
		Pty:        true,
		Cmd:        []string{"--"},
		Workdir:    p.GuestDir,
		Volumes: []string{
			p.HostDir + ":" + p.GuestDir,
			p.CacheDir + ":/home/devstep/cache",
			executable + ":/home/devstep/bin/devstep",
			"/var/run/docker.sock:/var/run/docker.sock",
		},
	})

	result, err := client.Run(opts)
	log.Debug("Docker run result: %+v", result)

	if err != nil {
		if result != nil && result.ContainerID != "" {
			client.RemoveContainer(result.ContainerID)
		}
		return result, err
	}

	return result, nil
}

func (p *project) buildWithCommand(client DockerClient, cliOpts *DockerRunOpts, cmd []string) (*DockerRunResult, error) {
	opts := p.Defaults.Merge(cliOpts, &DockerRunOpts{
		Image:      p.BaseImage,
		AutoRemove: false,
		Pty:        true,
		Cmd:        cmd,
		Workdir:    p.GuestDir,
		Volumes: []string{
			p.HostDir + ":" + p.GuestDir,
			p.CacheDir + ":/home/devstep/cache",
		},
	})

	result, err := client.Run(opts)
	log.Debug("Docker run result: %+v", result)

	if err != nil {
		// TODO: Write test for this behavior
		if result != nil && result.ContainerID != "" {
			client.RemoveContainer(result.ContainerID)
		}
		return result, err
	}

	if result.ExitCode != 0 {
		// TODO: Write test for this behavior
		client.RemoveContainer(result.ContainerID)
		return result, errors.New("Container exited with status != 0, skipping image commit.")
	}

	if changed, err := client.ContainerChanged(result.ContainerID); err != nil {
		return result, err

	} else if changed {
		if err = p.commit(client, result.ContainerID, "latest"); err != nil {
			return result, err
		}

		tag := time.Now().Local().Format("20060102150405")
		if err = p.commit(client, result.ContainerID, tag); err != nil {
			return result, err
		}

	} else {
		// TODO: Write test for this behavior
		fmt.Println("==> Skipping commit (container did not have any file changed)")
	}

	return result, nil
}
