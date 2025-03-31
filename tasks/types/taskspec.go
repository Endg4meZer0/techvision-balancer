package types

import (
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
)

type TaskSpec struct {
	ContainerConfig     container.Config
	ContainerHostConfig container.HostConfig
	NetworkingConfig    network.NetworkingConfig
	Name                string
}
