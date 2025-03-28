package main

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"time"

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
	contCount := 0
	failedCount := 0
	for _, tsk := range tskl {
		if tsk.Status.State == swarm.TaskStateFailed || tsk.Status.State == swarm.TaskStateRejected {
			failedCount++
			continue
		}
		contCount++
	}

	go func() {
		for len(tskl) > 0 {
			time.Sleep(5 * time.Second)
			newTskl, err := DockerClient.TaskList(context.Background(), types.TaskListOptions{
				Filters: filters.NewArgs(filters.Arg(
					"service", serv.ID,
				)),
			})
			if err != nil {
				continue
			}
			for _, tsk := range newTskl {
				if tsk.Status.State != swarm.TaskStateRunning {
					continue
				}

				for nodeID := range nodes {
					if nodeID == tsk.NodeID {
						nodes[nodeID].Containers[tsk.Status.ContainerStatus.ContainerID] = Container{
							TaskID: taskId,
						}
					}
				}
				tskl = slices.DeleteFunc(tskl, func(tsk1 swarm.Task) bool { return tsk1.ID == tsk.ID })
			}
		}
	}()

	return fmt.Sprintf("task launched successfully, containers total/already failed: %v/%v; MAY STILL BE PENDING, check /get later", contCount, failedCount), err
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
