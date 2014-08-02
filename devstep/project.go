package devstep

// The project interface provides access to the configuration, state and
// lifecycle of a Project.
type Project interface {
	Build() error
	Hack()  error
}

// An implementation of a Project.
type project struct { }

// This creates a new project
func NewProject() (Project, error) {
	return &project{}, nil
}

// Build the project and commit it to an image
func (p *project) Build() error {
	println("Will build")
	return nil
}

// Starts a hacking session on the project
func (p *project) Hack() error {
	println("Will hack")
	return nil
}
