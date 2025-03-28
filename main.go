package main

import (
	"context"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/swarm"
	"github.com/docker/docker/client"
	_ "github.com/joho/godotenv/autoload"
)

func main() {
	var err error
	DockerClient, err = client.NewClientWithOpts(
		client.FromEnv,
		client.WithAPIVersionNegotiation(),
	)
	if err != nil {
		panic(err)
	}

	nl, err := DockerClient.NodeList(context.Background(), types.NodeListOptions{})
	if err != nil {
		panic(err)
	}
	for _, n := range nl {
		node := Node{}
		node.IsManager = n.Spec.Role == swarm.NodeRoleManager
		node.Containers = make(map[string]Container)
		GlobalNodes[n.ID] = node
	}

	SetupServer()
}
