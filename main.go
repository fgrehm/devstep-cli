package main

import (
	"fmt"
	"os"
	"github.com/fsouza/go-dockerclient"
)

func main() {
	endpoint := "unix:///var/run/docker.sock"
	client, _ := docker.NewClient(endpoint)

	container, err := client.CreateContainer(docker.CreateContainerOptions{
		Name:  "container-from-api",
		Config: &docker.Config{
			Image: "busybox",
		},
	})

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	fmt.Println(container)

	// imgs, err := client.ListImages(true)
	// if err != nil {
	// 	fmt.Println(err)
	// 	os.Exit(1)
	// }

	// for _, img := range imgs {
	// 	fmt.Println("ID: ", img.ID)
	// 	fmt.Println("RepoTags: ", img.RepoTags)
	// 	fmt.Println("Created: ", img.Created)
	// 	fmt.Println("Size: ", img.Size)
	// 	fmt.Println("VirtualSize: ", img.VirtualSize)
	// 	fmt.Println("ParentId: ", img.ParentId)
	// 	fmt.Println("Repository: ", img.Repository)
	// }
}
