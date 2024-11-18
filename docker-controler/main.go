package main

import (
	"context" // For controlling cancellation and deadlines
	"fmt"     // For printing output

	// Docker API types
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
)

// Main function to control Docker containers
func main() {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		panic(err) // Handle any errors while creating Docker client 😅
	}

	// Define container ID to work with
	containerID := "455358e7674337ed8e3a5312f3ab68c79d2ca4fe0cfa97e5e4f664f2bcd577c6" // Replace with your container's ID 🌟

	// Start container
	// if err := startContainer(cli, containerID); err != nil {
	// 	fmt.Println("Failed to start container: ", err)
	// }

	// Stop container
	// if err := stopContainer(cli, containerID); err != nil {
	// 	fmt.Println("Failed to stop container: ", err)
	// }

	// // Duplicate container
	// if err := duplicateContainer(cli, containerID); err != nil {
	// 	fmt.Println("Failed to duplicate container: ", err)
	// }

	// Duplicate container with port mapping
	if err := duplicateContainerWithPort(cli, containerID, "8081:80"); err != nil {
		fmt.Println("Failed to duplicate container with port: ", err)
	}
}

// Function to start a container
func startContainer(cli *client.Client, containerID string) error {
	ctx := context.Background() // Create a context for cancellation control 🌼
	return cli.ContainerStart(ctx, containerID, container.StartOptions{})
}

// Function to start a container with port mapping 🚀✨
func startContainerWithPort(cli *client.Client, containerID string, portMapping string) error {
	ctx := context.Background() // Create a context for cancellation control 🌼

	// Create port bindings
	ports, err := nat.ParsePortSpec(portMapping)
	if err != nil {
		return err
	}

	// Start container with port binding
	hostConfig := &container.HostConfig{
		PortBindings: nat.PortMap{},
	}
	for _, p := range ports {
		hostConfig.PortBindings[p.Port] = []nat.PortBinding{p.Binding}
	}
	return cli.ContainerStart(ctx, containerID, container.StartOptions{})
}

// Function to stop a container
func stopContainer(cli *client.Client, containerID string) error {
	ctx := context.Background() // Context again! 🌼
	return cli.ContainerStop(ctx, containerID, container.StopOptions{})
}

// Function to duplicate (create new container from existing)
func duplicateContainer(cli *client.Client, containerID string) error {
	ctx := context.Background() // Keep the context for communication 💭

	// Get container configuration to use for the new container 📝
	containerJson, err := cli.ContainerInspect(ctx, containerID)
	if err != nil {
		return err
	}

	// Create new container using the old configuration 🎁
	resp, err := cli.ContainerCreate(ctx, &container.Config{
		Image: containerJson.Config.Image,
	}, nil, nil, nil, "")
	if err != nil {
		return err
	}

	fmt.Printf("New container ID: %s\n", resp.ID) // 🎉
	return nil
}


// Function to duplicate container with port mapping 🐣✨
func duplicateContainerWithPort(cli *client.Client, containerID string, portMapping string) error {
	ctx := context.Background() // Keep the context for communication 💭

	// Get container configuration to use for the new container 📝
	containerJson, err := cli.ContainerInspect(ctx, containerID)
	if err != nil {
		return err
	}

	// Create port bindings
	ports, err := nat.ParsePortSpec(portMapping)
	if err != nil {
		return err
	}

	// Create new container using the old configuration 🎁
	resp, err := cli.ContainerCreate(ctx, &container.Config{
		Image: containerJson.Config.Image,
	}, &container.HostConfig{
		PortBindings: nat.PortMap{
			ports[0].Port: []nat.PortBinding{
				{
					HostIP:   "",
					HostPort: ports[0].Binding.HostPort,
				},
			},
		},
	}, &network.NetworkingConfig{}, nil, "")
	if err != nil {
		return err
	}

	fmt.Printf("New container ID with port binding: %s\n", resp.ID) // 🎉
	return nil
}