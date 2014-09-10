package main

import (
	"fmt"

	"github.com/andrew-d/docker-tools/log"
	"github.com/fsouza/go-dockerclient"
)

func checkContainerExists(client *docker.Client, container *Container) (bool, error) {
	inspect, err := client.InspectContainer(container.Name)
	if err != nil {
		if _, ok := err.(*docker.NoSuchContainer); ok {
			return false, nil
		} else {
			return false, fmt.Errorf("Error inspecting container: %s", err)
		}
	}
	// Sanity check: ensure the container is using our image.
	imageInfo, err := client.InspectImage(container.Image)
	if err != nil {
		if err == docker.ErrNoSuchImage {
			return false, fmt.Errorf("Error creating container: no such image (%s)",
				container.Image)
		}

		return false, fmt.Errorf("Error creating container: %s", container.Image)
	}

	if inspect.Image != imageInfo.ID {
		return false, fmt.Errorf("Container exists, but is not using the correct image (using: %s, expected: %s)",
			inspect.Image, imageInfo.ID)
	}

	return true, nil
}

func createContainer(client *docker.Client, container *Container) error {
	// Set the options used when creating our container.
	opts := docker.CreateContainerOptions{
		Name: container.Name,
		Config: &docker.Config{
			Image: container.Image,
		},
	}

	for _, env := range container.Env {
		opts.Config.Env = append(opts.Config.Env, env.Key+"="+env.Value)
	}
	for _, port := range container.Ports {
		exposed := docker.Port(fmt.Sprintf("%d/tcp", port.ContainerPort))
		opts.Config.ExposedPorts[exposed] = struct{}{}
	}
	for _, mount := range container.Mount {
		opts.Config.Volumes[mount.ContainerDir] = struct{}{}
	}
	for i, mfrom := range container.MountFrom {
		if i > 0 {
			log.Warnf("%s: Currently only support one 'volumes-from'.  The last entry will be used.",
				container.Name)
		}
		opts.Config.VolumesFrom = mfrom
	}

	// Create the container.
	_, err := client.CreateContainer(opts)
	return err
}

func cmdCreate(config *Config) {
	client, err := getClient()
	if err != nil {
		log.Errorf("Error getting client: %s", err)
		return
	}

	created := 0
	skipped := 0

	for _, idx := range config.ContainerSort {
		container := config.Containers[idx]

		// Check if the container exists.
		exists, err := checkContainerExists(client, container)
		if err != nil {
			log.Errorf("%s: %s", container.Name, err)
			return
		} else if exists {
			log.Infof("%s: Container exists, skipping...", container.Name)
			skipped++
			continue
		} else {
			log.Infof("%s: Container not found, creating...", container.Name)
		}

		err = createContainer(client, container)
		if err != nil {
			log.Errorf("%s: Error creating: %s", container.Name, err)
			return
		}

		log.Infof("%s: Created container", container.Name)
		created++
	}

	log.Infof("Finished creating containers")
	log.Infof("Total: %d (%d created / %d skipped)",
		len(config.Containers), created, skipped)
}
