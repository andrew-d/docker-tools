package main

import (
	"os"
	"fmt"

	"github.com/fsouza/go-dockerclient"
)

func getClient() (*docker.Client, error) {
	host := os.Getenv("DOCKER_HOST")
	if len(host) == 0 {
		host = "unix:///var/run/docker.sock"
	}

	client, err := docker.NewClient(host)
	if err != nil {
		return nil, fmt.Errorf("Error creating: %s", err)
	}

	err = client.Ping()
	if err != nil {
		return nil, fmt.Errorf("Error pinging: %s", err)
	}

	return client, nil
}
