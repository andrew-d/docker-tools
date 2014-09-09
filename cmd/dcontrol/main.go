package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"

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

	log.Println("Started")

	f, err := os.Open(flagConfig)
	if err != nil {
		log.Printf("Error opening config file: %s", err)
		return
	}
	defer f.Close()

	data, err := ioutil.ReadAll(f)
	if err != nil {
		log.Printf("Error reading config file: %s", err)
		return
	}

	var config map[string]interface{}
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		log.Printf("Error parsing config file: %s", err)
		return
	}

	log.Printf("Config: %+v", config)

	// Parse containers.
	var ok bool
	var subConfig map[interface{}]interface{}

	if subConfig, ok = config["containers"].(map[interface{}]interface{}); !ok {
		log.Println("Missing or invalid 'containers' key in config")
		return
	}

	var containers []*Container
	for k, v := range subConfig {
		name := k.(string)

		c, err := parseContainer(name, v)
		if err != nil {
			log.Printf("Error parsing container %s: %s", name, err)
			return
		}

		containers = append(containers, c)
	}

	// Find the topological sorting of our containers.
	toposort, err := TopoSortContainers(containers)
	if err != nil {
		// TODO: good message?
		log.Println(err)
		return
	}
	_ = toposort

	// Figure out what we're doing with our config.
	cmd := strings.ToLower(flag.Arg(0))
	switch cmd {
	case "create":
		CmdCreate()

	default:
		log.Printf("Unknown command: %s", cmd)
		return
	}

	log.Println("Completed successfully")
}
