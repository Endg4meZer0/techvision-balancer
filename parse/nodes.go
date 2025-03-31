package parse

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"techvision/balancer/global"

	"github.com/docker/docker/client"
	"github.com/goccy/go-yaml"
)

func ParseNodes() (nodes global.Nodes) {
	nodes.M.Lock()
	defer nodes.M.Unlock()

	if _, err := os.Stat("nodes.yaml"); os.IsNotExist(err) {
		log.Fatalln("can't find nodes.yaml, terminating")
		return
	}
	f, err := os.ReadFile("nodes.yaml")
	if err != nil {
		log.Fatalln(err)
		return
	}
	var dto struct {
		Nodes []struct {
			IP     string            `yaml:"ip"`
			Port   uint16            `yaml:"port"`
			Labels map[string]string `yaml:"labels"`
		} `yaml:"nodes"`
	}

	err = yaml.Unmarshal(f, &dto)
	if err != nil {
		log.Fatalln(err)
		return
	}

	nodes.N = make(map[string]global.Node)

	for _, n := range dto.Nodes {
		var node global.Node
		node.Client, err = client.NewClientWithOpts(
			client.WithHost(fmt.Sprintf("tcp://%s:%v", n.IP, n.Port)),
			client.WithAPIVersionNegotiation(),
		)
		if err != nil {
			log.Fatalf("couldn't connect to node %s:%v, exiting", n.IP, n.Port)
			return
		}
		deadline, cFunc := context.WithDeadline(context.Background(), time.Now().Add(5*time.Second))
		_, err := node.Client.Ping(deadline)
		if err != nil {
			log.Fatalf("couldn't ping node %s:%v, exiting", n.IP, n.Port)
			return
		}
		cFunc()

		node.Labels = n.Labels
		node.Containers = make(map[string]global.Container)
		nodes.N[fmt.Sprintf("%s:%v", n.IP, n.Port)] = node
	}

	return
}
