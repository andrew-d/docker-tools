package main

import (
	"fmt"

	"github.com/andrew-d/docker-tools/log"
	"github.com/fsouza/go-dockerclient"
)

func cmdCreate(config *Config) {
	client, err := getClient()
	if err != nil {
		log.Errorf("Error getting client: %s", err)
		return
	}

	for _, idx := range config.ContainerSort {
		container := config.Containers[idx]

		// Check if the container exists.
		inspect, err := client.InspectContainer(container.Name)
		if err != nil {
			if _, ok := err.(*docker.NoSuchContainer); ok {
				log.Infof("Container %s not found, creating...", container.Name)
			} else {
				log.Errorf("Error inspecting container %s: %s", container.Name, err)
				return
			}
		}

		// Sanity check: ensure the container is using our image.
		if inspect.Image != container.Image {
			log.Errorf("Container %s exists, but is not using the correct image (using: %s, expected: %s)",
				container.Name, inspect.Image, container.Image)
			return
		}

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
				log.Warnf(`Currently only support one 'volumes-from'.  The last entry will be used.`)
			}
			opts.Config.VolumesFrom = mfrom
		}

		// Create the container.
		_, err = client.CreateContainer(opts)
		if err != nil {
			log.Errorf("Error creating container %s: %s", container.Name, err)
			return
		}
	}
}
