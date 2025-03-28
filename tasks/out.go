package tasks

import "github.com/docker/docker/api/types/swarm"

var Tasks map[string]swarm.ServiceSpec = map[string]swarm.ServiceSpec{
	"1": task1,
}
