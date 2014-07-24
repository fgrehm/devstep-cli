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
		Config: &docker.Config{
			Image:        "fgrehm/devstep:v0.0.1",
			Cmd:          []string{"/.devstep/bin/build-project", "/workspace"},
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
			ID:    container.ID,
			Force: true,
		})

		//client.RemoveImage("new-devstep/testing-build:latest")
	}()

	if err := runBuild(client, container); err != nil {
		fmt.Println("Error during build!")
		fmt.Println(err)
		os.Exit(1)
	}
	if err := commit(client, container.ID); err != nil {
		fmt.Println("Error during commit!")
		fmt.Println(err)
		os.Exit(1)
	}
}

func runBuild(client *docker.Client, container *docker.Container) (err error) {
	// Start container
	// pwd, err := os.Getwd()
	// if err != nil {
	// 	fmt.Println(err)
	// 	os.Exit(1)
	// }

	exitedChannel := make(chan struct{})

	// This goroutine will listen to Docker events and will signal that is has
	// stopped at the exitedChannel
	go listenForContainerExit(client, container.ID, exitedChannel)

	// Attach to the container on a separate goroutine
	go attachToContainer(client, container.ID)

	binds := []string{
		"/home/fabio/projects/oss/devstep-cli:/workspace",
		"/tmp/devstep/cache:/.devstep/cache",
	}

	err = client.StartContainer(container.ID, &docker.HostConfig{
		Binds: binds,
	})

	<-exitedChannel

	// TODO: Check for exit status

	return err
}

func listenForContainerExit(client *docker.Client, containerID string, exitedChannel chan struct{}) error {
	listenerChannel := make(chan *docker.APIEvents)
	client.AddEventListener(listenerChannel)

	for {
		event := <-listenerChannel
		if event.ID == containerID && event.Status == "die" {
			exitedChannel <- struct{}{}
		}
	}
}

func attachToContainer(client *docker.Client, containerID string) {
	client.AttachToContainer(docker.AttachToContainerOptions{
		Container:    containerID,
		OutputStream: os.Stdout,
		ErrorStream:  os.Stderr,
		Stdout:       true,
		Stderr:       true,
		Stream:       true,
		RawTerminal:  true,
	})
}

func commit(client *docker.Client, containerID string) error {
	_, err := client.CommitContainer(docker.CommitContainerOptions{
		Container:  containerID,
		Repository: "new-devstep/testing-build",
		Tag:        "latest",
		Run: &docker.Config{
			Cmd: []string{"/.devstep/bin/hack"},
		},
	})
	return err
}
