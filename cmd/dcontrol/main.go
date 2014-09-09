package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/andrew-d/docker-tools/log"
	flag "github.com/ogier/pflag"
	"gopkg.in/yaml.v1"
)

var (
	flagConfig string
)

func init() {
	flag.StringVarP(&flagConfig, "config", "c", "./config.yaml",
		"The config file to use")
}

func usage() {
	fmt.Println(strings.TrimSpace(`
Usage: dcontrol <command> [options]

Commands:
    create <cluster>        Builds containers for a given cluster.
    start <cluster>         Start all containers in a given cluster.
    stop <cluster>          Stop all containers in a given cluster.
    restart <cluster>       Restart all containers in a given cluster.
    status <cluster>        Show the status of all the containers in a given
                            cluster.

Options:
`))
	flag.PrintDefaults()
	os.Exit(1)
}

func main() {
	flag.Parse()

	if flag.NArg() < 2 {
		usage()
	}

	log.Infof("Started")

	f, err := os.Open(flagConfig)
	if err != nil {
		log.Errorf("Error opening config file: %s", err)
		return
	}
	defer f.Close()

	data, err := ioutil.ReadAll(f)
	if err != nil {
		log.Errorf("Error reading config file: %s", err)
		return
	}

	var rawConfig map[string]interface{}
	err = yaml.Unmarshal(data, &rawConfig)
	if err != nil {
		log.Errorf("Error parsing config file: %s", err)
		return
	}

	log.Debugf("Config: %+v", rawConfig)

	config := &Config{
		Containers: []*Container{},
	}

	var ok bool
	var subConfig map[interface{}]interface{}

	// Parse containers.
	if subConfig, ok = rawConfig["containers"].(map[interface{}]interface{}); !ok {
		log.Errorf("Missing or invalid 'containers' key in config")
		return
	}

	for k, v := range subConfig {
		name := k.(string)

		c, err := parseContainer(name, v)
		if err != nil {
			log.Errorf("Error parsing container %s: %s", name, err)
			return
		}

		config.Containers = append(config.Containers, c)
	}

	// Find the topological sorting of our containers.
	config.ContainerSort, err = TopoSortContainers(config.Containers)
	if err != nil {
		// TODO: good message?
		log.Errorf("Error topologically sorting: %s", err)
		return
	}

	// Figure out what we're doing with our config.
	cmd := strings.ToLower(flag.Arg(0))
	switch cmd {
	case "create":
		cmdCreate(config)

	default:
		log.Errorf("Unknown command: %s", cmd)
		return
	}

	log.Infof("Completed successfully")
}
