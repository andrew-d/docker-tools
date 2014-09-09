package main

import (
	"fmt"
	"strconv"
	"strings"
)

func parseContainer(name string, config interface{}) (*Container, error) {

	switch v := config.(type) {
	case string:
		// The value is an image name
		ret := &Container{
			Name:  name,
			Image: v,
		}
		return ret, nil

	case map[interface{}]interface{}:
		// Complex config
		return parseContainerMap(name, v)

	default:
		return nil, fmt.Errorf("Unknown type for 'containers' key: %T", v)
	}
}

func parseContainerMap(name string, config map[interface{}]interface{}) (*Container, error) {
	var key string
	var ok bool

	ret := &Container{Name: name}

	for k, val := range config {
		if key, ok = k.(string); !ok {
			return nil, fmt.Errorf("Unknown key in config for container %s: %+v", name, k)
		}

		var err error
		switch key {
		case "image":
			err = parseContainerMapImage(ret, val)

		case "dependencies":
			err = parseContainerMapDependencies(ret, val)

		case "env":
			err = parseContainerMapEnv(ret, val)

		case "ports":
			err = parseContainerMapPorts(ret, val)

		case "mount":
			err = parseContainerMapMount(ret, val)

		case "mount-from":
			err = parseContainerMapMountFrom(ret, val)

		case "privileged":
			err = parseContainerMapPrivileged(ret, val)

		// TODO: extra runtime arguments

		default:
			return nil, fmt.Errorf("Unknown key in config for container %s: %s", name, key)
		}

		if err != nil {
			return nil, fmt.Errorf("Error parsing key '%s' for container %s: %s", key, name, err)
		}
	}

	return ret, nil
}

func parseContainerMapImage(ret *Container, val interface{}) error {
	var ok bool

	ret.Image, ok = val.(string)
	if !ok {
		return fmt.Errorf("Unknown value type: %T", val)
	}
	return nil
}

func parseContainerMapDependencies(ret *Container, val interface{}) error {
	var ok bool
	var deps []interface{}
	var dep string

	if deps, ok = val.([]interface{}); !ok {
		return fmt.Errorf("Unknown value type: %T", val)
	}

	ret.Dependencies = []string{}
	for _, d := range deps {
		if dep, ok = d.(string); !ok {
			return fmt.Errorf("Unknown value type in array: %T", d)
		}

		ret.Dependencies = append(ret.Dependencies, dep)
	}

	return nil
}

func parseContainerMapEnv(ret *Container, val interface{}) error {
	var ok bool
	var envs []interface{}
	var env string

	if envs, ok = val.([]interface{}); !ok {
		return fmt.Errorf("Unknown value type: %T", val)
	}

	ret.Env = []EnvConfig{}
	for i, e := range envs {
		if env, ok = e.(string); !ok {
			return fmt.Errorf("Unknown value type in array: %T", e)
		}
		parts := strings.SplitN(env, "=", 2)
		if len(parts) != 2 {
			return fmt.Errorf("Env entry %d not in form KEY=VAL", i)
		}

		ret.Env = append(ret.Env, EnvConfig{Key: parts[0], Value: parts[1]})
	}

	return nil
}

func parseContainerMapPorts(ret *Container, val interface{}) error {
	var ok bool
	var ports []interface{}

	if ports, ok = val.([]interface{}); !ok {
		return fmt.Errorf("Unknown value type: %T", val)
	}

	ret.Ports = []PortConfig{}
	for i, p := range ports {
		conf := PortConfig{}

		switch v := p.(type) {
		case int:
			if v <= 0 || v > 65535 {
				return fmt.Errorf("Port %d out of range: %d", i, v)
			}

			conf.IP = "0.0.0.0"
			conf.HostPort = uint16(v)
			conf.ContainerPort = uint16(v)

		case string:
			parts := strings.Split(v, ":")

			var hostPort, containerPort uint64
			var err error

			switch len(parts) {
			case 1:
				conf.IP = "0.0.0.0"
				hostPort, err = strconv.ParseUint(parts[0], 10, 16)
				if err != nil {
					return err
				}
				containerPort = hostPort

			case 2:
				conf.IP = "0.0.0.0"
				hostPort, err = strconv.ParseUint(parts[0], 10, 16)
				if err != nil {
					return err
				}
				containerPort, err = strconv.ParseUint(parts[1], 10, 16)
				if err != nil {
					return err
				}

			case 3:
				conf.IP = parts[0]
				containerPort, err = strconv.ParseUint(parts[2], 10, 16)
				if err != nil {
					return err
				}

				// Note: the middle part may be empty, in which case it is
				// the same as the last.
				if len(parts[1]) > 0 {
					hostPort, err = strconv.ParseUint(parts[1], 10, 16)
					if err != nil {
						return err
					}
				} else {
					hostPort = containerPort
				}

			default:
				return fmt.Errorf("Unknown port format for port %d", i)
			}

			// Note: bounds checking done above.
			conf.HostPort = uint16(hostPort)
			conf.ContainerPort = uint16(containerPort)

		default:
			return fmt.Errorf("Unknown value type in array: %T", v)
		}

		ret.Ports = append(ret.Ports, conf)
	}

	return nil
}

func parseContainerMapMount(ret *Container, val interface{}) error {
	var ok bool
	var mounts []interface{}
	var mount string

	if mounts, ok = val.([]interface{}); !ok {
		return fmt.Errorf("Unknown value type: %T", val)
	}

	ret.Mount = []MountConfig{}
	for i, m := range mounts {
		if mount, ok = m.(string); !ok {
			return fmt.Errorf("Unknown value type in array: %T", m)
		}
		parts := strings.Split(mount, ":")

		var mount MountConfig
		switch len(parts) {
		case 2:
			mount.HostDir = parts[0]
			mount.ContainerDir = parts[1]
			mount.Type = MountTypeReadWrite

		case 3:
			mount.HostDir = parts[0]
			mount.ContainerDir = parts[1]

			switch parts[2] {
			case "rw":
				mount.Type = MountTypeReadWrite
			case "ro":
				mount.Type = MountTypeReadOnly
			default:
				return fmt.Errorf("Mount entry %d has invalid mount type: %s", i, parts[2])
			}
		default:
			return fmt.Errorf("Mount entry %d not in form /host/dir:/container/dir[:type]", i)
		}

		ret.Mount = append(ret.Mount, mount)
	}

	return nil
}

func parseContainerMapMountFrom(ret *Container, val interface{}) error {
	var ok bool
	var mounts []interface{}
	var mount string

	if mounts, ok = val.([]interface{}); !ok {
		return fmt.Errorf("Unknown value type: %T", val)
	}

	ret.MountFrom = []string{}
	for _, d := range mounts {
		if mount, ok = d.(string); !ok {
			return fmt.Errorf("Unknown value type in array: %T", d)
		}

		ret.MountFrom = append(ret.MountFrom, mount)
	}

	return nil
}

func parseContainerMapPrivileged(ret *Container, val interface{}) error {
	var ok bool

	ret.Privileged, ok = val.(bool)
	if !ok {
		return fmt.Errorf("Unknown value type: %T", val)
	}
	return nil
}
