package main

import (
	"context"
	"errors"
	"fmt"
	"slices"

	"techvision/balancer/tasks"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/swarm"
)

func (nodes Nodes) Get() (Nodes, error) {
	return nodes, nil
}

func (nodes Nodes) AddTask(taskId string) (string, error) {
	spec, ok := tasks.Tasks[taskId]
	if !ok {
		return "", errors.New("task spec not found")
	}

	res, err := DockerClient.ServiceCreate(
		context.Background(),
		spec,
		types.ServiceCreateOptions{},
	)
	if err != nil {
		return "", err
	}

	serv, _, err := DockerClient.ServiceInspectWithRaw(
		context.Background(),
		res.ID,
		types.ServiceInspectOptions{},
	)
	if err != nil {
		return "", err
	}
	tskl, err := DockerClient.TaskList(context.Background(), types.TaskListOptions{
		Filters: filters.NewArgs(filters.Arg(
			"service", serv.ID,
		)),
	})
	if err != nil {
		return "", err
	}
	nodeCount := 0
	contCount := 0
	failedCount := 0
	usedNodes := make([]string, 0, len(tskl))
	for _, tsk := range tskl {
		if tsk.Status.State == swarm.TaskStateFailed || tsk.Status.State == swarm.TaskStateRejected {
			failedCount++
			continue
		}

		for nodeID := range nodes {
			if nodeID == tsk.NodeID {
				nodes[nodeID].Containers[tsk.Status.ContainerStatus.ContainerID] = Container{
					TaskID: taskId,
				}
				if !slices.Contains(usedNodes, nodeID) {
					usedNodes = append(usedNodes, nodeID)
					nodeCount++
				}
			}
		}
		contCount++
	}

	return fmt.Sprintf("task started successfully, nodes/containers used: %v/%v; failed containers: %v", nodeCount, contCount, failedCount), err
}

func (nodes Nodes) RemoveTask(taskId string) (string, error) {
	spec, ok := tasks.Tasks[taskId]
	if !ok {
		return "", errors.New("task spec not found")
	}

	servs, err := DockerClient.ServiceList(
		context.Background(),
		types.ServiceListOptions{
			Filters: filters.NewArgs(
				filters.Arg("name", spec.Annotations.Name),
			),
		},
	)
	if err != nil {
		return "", err
	}
	if len(servs) == 0 {
		return "", errors.New("task not started")
	}
	for _, serv := range servs {
		err := DockerClient.ServiceRemove(context.Background(), serv.ID)
		if err != nil {
			return "", err
		}

		for nodeID := range GlobalNodes {
			for contID, cont := range GlobalNodes[nodeID].Containers {
				if cont.TaskID == taskId {
					delete(GlobalNodes[nodeID].Containers, contID)
				}
			}
		}
	}

	return "successfully removed task", err
}
