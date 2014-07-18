package main

import (
	"fmt"
	"github.com/fsouza/go-dockerclient"
	"os"
)

func main() {
	endpoint := "unix:///var/run/docker.sock"
	client, _ := docker.NewClient(endpoint)

	// Create container
	container, err := client.CreateContainer(docker.CreateContainerOptions{
		Name: "testing-new-devstep-cli",
		Config: &docker.Config{
			//Image:        "ubuntu:14.04",
			//Cmd:          []string{"/bin/bash"},
			Image:        "fgrehm/devstep:v0.0.1",
			Cmd:          []string{"/.devstep/bin/hack"},
			OpenStdin:    true,
			AttachStdin:  true,
			AttachStdout: true,
			AttachStderr: true,
			Tty:          true,
		},
	})

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// Start container
	pwd, err := os.Getwd()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	binds := []string{
		pwd + ":/workspace",
		"/tmp/devstep/cache:/.devstep/cache",
	}
	err = client.StartContainer(container.ID, &docker.HostConfig{
		Binds: binds,
	})

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// Attach to the container
	err = client.AttachToContainer(docker.AttachToContainerOptions{
		Container:    container.ID,
		InputStream:  os.Stdin,
		OutputStream: os.Stdout,
		ErrorStream:  os.Stderr,
		Stdin:        true,
		Stdout:       true,
		Stderr:       true,
		Stream:       true,
		RawTerminal:  true,
	})

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
