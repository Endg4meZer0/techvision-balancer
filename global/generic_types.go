package global

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
)

type Container struct {
	TaskID string
	Spec   container.Summary
}

type Node struct {
	Client     *client.Client
	Labels     map[string]string
	Containers map[string]Container
}

type Nodes struct {
	M sync.Mutex
	N map[string]Node
}

// IsWorking returns true (working) if len(Containers) >= 1
func (n *Node) IsWorking() bool { return len(n.Containers) >= 1 }
func (n *Node) HasTask(taskID string) bool {
	for _, v := range n.Containers {
		if v.TaskID == taskID {
			return true
		}
	}
	return false
}
func (n *Node) IsPortTaken(port uint16) bool {
	containers, err := n.Client.ContainerList(context.Background(), container.ListOptions{
		All: true,
	})
	if err != nil {
		return false
	}
	for _, cont := range containers {
		for _, p := range cont.Ports {
			if p.PublicPort == port {
				return true
			}
		}
	}
	return false
}
func (n *Node) GetContainerByID(id string) (Container, error) {
	for key, cont := range n.Containers {
		if id == key {
			return cont, nil
		}
	}
	return Container{}, errors.New("container not found")
}

type ContainerSpecDTO struct {
	Created  time.Time `json:"created"`
	State    string    `json:"state"`
	Status   string    `json:"status"`
	OutPorts []uint16  `json:"out_ports"`
}

type ContainerDTO struct {
	TaskID string           `json:"task_id"`
	Spec   ContainerSpecDTO `json:"spec"`
}

type NodeDTO struct {
	Labels     map[string]string       `json:"labels"`
	Containers map[string]ContainerDTO `json:"containers"`
}

func (n *Nodes) ToJSON() string {
	var nodesDTO = make(map[string]NodeDTO)
	for host, node := range n.N {
		var nodeDTO NodeDTO
		nodeDTO.Labels = node.Labels
		nodeDTO.Containers = make(map[string]ContainerDTO)
		for cntID, cnt := range node.Containers {
			var contDTO ContainerDTO
			contDTO.TaskID = cnt.TaskID
			contDTO.Spec.State = cnt.Spec.State
			contDTO.Spec.Status = cnt.Spec.Status
			contDTO.Spec.Created = time.Unix(cnt.Spec.Created, 0)
			contDTO.Spec.OutPorts = make([]uint16, 0, len(cnt.Spec.Ports))
			for _, p := range cnt.Spec.Ports {
				contDTO.Spec.OutPorts = append(contDTO.Spec.OutPorts, p.PublicPort)
			}

			nodeDTO.Containers[cntID] = contDTO
		}

		nodesDTO[host] = nodeDTO
	}

	b, _ := json.Marshal(nodesDTO)
	return string(b)
}
