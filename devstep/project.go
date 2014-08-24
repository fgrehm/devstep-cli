package devstep

// The project interface provides access to the configuration, state and
// lifecycle of a Project.
type Project interface {
	Build(DockerClient) error
	Hack(DockerClient) error
}

// Project specific configuration, usually parsed from an yaml file
type ProjectConfig struct {
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
	baseImage      string
	repositoryName string
	hostDir        string
	guestDir       string
	cacheDir       string
	defaults       *DockerRunOpts
	hackOpts       *DockerRunOpts
}

// This creates a new project
func NewProject(config *ProjectConfig) (Project, error) {
	// TODO: Set defaults if not provided
	project := &project{
		baseImage:      config.BaseImage,
		repositoryName: config.RepositoryName,
		hostDir:        config.HostDir,
		guestDir:       config.GuestDir,
		cacheDir:       config.CacheDir,
		defaults:       config.Defaults,
		hackOpts:       config.HackOpts,
	}
	return project, nil
}

// Build the project and commit it to an image
func (p *project) Build(client DockerClient) error {
	println("Will build")
	return nil
}

// Starts a hacking session on the project
func (p *project) Hack(client DockerClient) error {
	volumes := []string{
		p.hostDir + ":" + p.guestDir,
		"/tmp/devstep/cache:/.devstep/cache",
	}

	client.Run(&DockerRunOpts{
		Image:      p.baseImage,
		AutoRemove: true,
		Pty:        true,
		Cmd:        []string{"/.devstep/bin/hack"},
		Volumes:    volumes,
		Workdir:    p.guestDir,
	})
	return nil
}
