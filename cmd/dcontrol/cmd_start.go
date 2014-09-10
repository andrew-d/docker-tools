package main

import (
	"fmt"

	"github.com/andrew-d/docker-tools/log"
	"github.com/fsouza/go-dockerclient"
)

func cmdStart(config *Config) {
	client, err := getClient()
	if err != nil {
		log.Errorf("Error getting client: %s", err)
		return
	}

	started := 0
	skipped := 0

	for _, idx := range config.ContainerSort {
		container := config.Containers[idx]

		// Check if the container exists.
		exists, err := checkContainerExists(client, container)
		if err != nil {
			log.Errorf("%s: %s", container.Name, err)
			return
		} else if exists {
			log.Infof("%s: Container exists", container.Name)
		} else {
			log.Errorf("%s: Container not found, did you run `dcontrol create`?",
				container.Name)
			return
		}

		// Check if the container is started.
		inspect, err := client.InspectContainer(container.Name)
		if err != nil {
			log.Errorf("%s: Error inspecting container: %s", container.Name, err)
			return
		}
		if inspect.State.Running {
			log.Infof("%s: Container is already running, skipping...", container.Name)
			skipped++
			continue
		}

		// Build options
		opts := &docker.HostConfig{
			Privileged: container.Privileged,
		}

		for _, port := range container.Ports {
			dport := docker.Port(fmt.Sprintf("%d/tcp", port.ContainerPort))
			opts.PortBindings[dport] = append(opts.PortBindings[dport], docker.PortBinding{
				HostIp:   port.IP,
				HostPort: fmt.Sprintf("%d", port.HostPort),
			})
		}
		for _, mount := range container.Mount {
			bind := mount.HostDir + ":" + mount.ContainerDir

			// TODO: read/write?

			opts.Binds = append(opts.Binds, bind)
		}
		for _, mfrom := range container.MountFrom {
			opts.VolumesFrom = append(opts.VolumesFrom, mfrom)
		}
		for _, dep := range container.Dependencies {
			opts.Links = append(opts.Links, fmt.Sprintf("%s:%s", dep.Name, dep.Alias))
		}

		err = client.StartContainer(container.Name, opts)
		if err != nil {
			log.Errorf("%s: Error starting: %s", container.Name, err)
			return
		}

		log.Infof("%s: Started container", container.Name)
		started++
	}

	log.Infof("Finished starting containers")
	log.Infof("Total: %d (%d started / %d skipped)",
		len(config.Containers), started, skipped)
}
