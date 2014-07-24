package main

import (
	"fmt"
	"github.com/fgrehm/go-dockerpty"
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
			Image:        "fgrehm/devstep:v0.0.1",
			Cmd:          []string{"/.devstep/bin/hack"},
			OpenStdin:    true,
			StdinOnce:    true,
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

	defer func() {
		client.RemoveContainer(docker.RemoveContainerOptions{
			ID: container.ID,
			Force: true,
		})
	}()

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
	err = dockerpty.Start(client, container, &docker.HostConfig{
		Binds: binds,
	})

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
