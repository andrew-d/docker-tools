package main

import (
	"io/ioutil"
	"log"
	"os"

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

func main() {
	flag.Parse()

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

	for i, c := range containers {
		log.Printf("%d: %#v", i, c)
	}

	log.Println("Completed successfully")
}
