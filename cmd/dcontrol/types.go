package main

type Container struct {
	Name       string
	Image      string
	Privileged bool

	Dependencies []string
	Env          []EnvConfig
	Ports        []PortConfig
	Mount        []MountConfig
	MountFrom    []string
}

type PortConfig struct {
	IP            string
	HostPort      uint16
	ContainerPort uint16
}

type EnvConfig struct {
	Key   string
	Value string
}

type MountType int

const (
	MountTypeInvalid MountType = iota
	MountTypeReadOnly
	MountTypeReadWrite
)

func (t MountType) String() string {
	switch t {
	case MountTypeInvalid:
		return "<invalid>"
	case MountTypeReadOnly:
		return "ReadOnly"
	case MountTypeReadWrite:
		return "ReadWrite"
	default:
		return "<unknown>"
	}
}

type MountConfig struct {
	HostDir      string
	ContainerDir string
	Type         MountType
}
