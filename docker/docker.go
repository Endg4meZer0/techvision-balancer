package docker

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"techvision/balancer/global"
	"techvision/balancer/tasks"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
)

func Get() (string, error) {
	return global.GNodes.ToJSON(), nil
}

func AddTask(taskID string, taskData global.JSONData) (any, error) {
	tsk, ok := tasks.Tasks[taskID]
	if !ok {
		return "", errors.New("task spec not found")
	}

	global.GNodes.M.Lock()
	defer global.GNodes.M.Unlock()

	var chosenNode global.Node
	var host string
	for nHost := range global.GNodes.N {
		if tsk.GetSpec().ContainerConfig.Labels["gpu"] == "true" {
			gpuMemRequired, err := strconv.Atoi(tsk.GetSpec().ContainerConfig.Labels["gpu.required"])
			if err != nil {
				return "", errors.New("gpu.required label is not a number")
			}
			gpuMemConsumed := 0
			for _, cont := range global.GNodes.N[nHost].Containers {
				if cont.Spec.Labels["gpu"] == "true" {
					contGpuMemConsumed, err := strconv.Atoi(cont.Spec.Labels["gpu.required"])
					if err != nil {
						return "", errors.New("gpu.required label is not a number")
					}
					gpuMemConsumed += contGpuMemConsumed
				}
			}
			gpuMemTotal, err := strconv.Atoi(global.GNodes.N[nHost].Labels["gpu.total"])
			if err != nil {
				return "", errors.New("gpu.total label is not a number")
			}

			// If GPU memory is not enough - skip the node
			if gpuMemConsumed+gpuMemRequired > gpuMemTotal {
				continue
			}
		}

		chosenNode = global.GNodes.N[nHost]
		host = nHost
		break
	}
	if chosenNode.Client == nil {
		return "", errors.New("no free nodes for this task")
	}

	spec := tsk.GetSpec()

	// Some tasty spaghetti to prevent port collisions
	for portBind := range spec.ContainerHostConfig.PortBindings {
		hostBinds := spec.ContainerHostConfig.PortBindings[portBind]
		for i, hostBind := range hostBinds {
			hostPort, err := strconv.ParseUint(hostBind.HostPort, 10, 16)
			if err != nil {
				return "", errors.New("host port is not a number")
			}
			checkForInternalCollisions := func(hostPort uint16) bool {
				for p, hBind := range spec.ContainerHostConfig.PortBindings {
					if p == portBind {
						continue
					}

					for _, binds := range hBind {
						if binds.HostPort == strconv.FormatUint(uint64(hostPort), 10) {
							return true
						}
					}
				}
				return false
			}

			for chosenNode.IsPortTaken(uint16(hostPort)) || checkForInternalCollisions(uint16(hostPort)) {
				hostPort++
			}
			hostBind.HostPort = strconv.FormatUint(hostPort, 10)
			hostBinds[i] = hostBind
		}
		spec.ContainerHostConfig.PortBindings[portBind] = hostBinds
	}

	// Check for name conflicts
	cntSums, err := chosenNode.Client.ContainerList(context.Background(), container.ListOptions{
		All: true,
	})
	if err != nil {
		return "", nil
	}
	newName := fmt.Sprintf("%s.%v", spec.Name, 0)
	for _, v := range cntSums {
		for _, cntName := range v.Names {
			if splits := strings.Split(cntName, "."); strings.Contains(cntName, newName) && len(splits) != 1 {
				idNum, err := strconv.Atoi(splits[1])
				if err != nil {
					continue
				}
				newName = fmt.Sprintf("%s.%v", spec.Name, idNum+1)
			}
		}
	}
	spec.Name = newName

	// Create the container associated with the task
	cnt, err := chosenNode.Client.ContainerCreate(
		context.Background(),
		&spec.ContainerConfig,
		&spec.ContainerHostConfig,
		&spec.NetworkingConfig,
		v1.DescriptorEmptyJSON.Platform,
		spec.Name,
	)
	if err != nil {
		return "", err
	}

	err = chosenNode.Client.ContainerStart(context.Background(), cnt.ID, container.StartOptions{})
	if err != nil {
		return "", err
	}

	cntSums, err = chosenNode.Client.ContainerList(context.Background(), container.ListOptions{
		All: false,
		Filters: filters.NewArgs(
			filters.Arg("id", cnt.ID),
		),
	})
	if err != nil {
		return "", err
	}

	id := fmt.Sprintf("%s_%s", taskID, cnt.ID)

	chosenNode.Containers[id] = global.Container{
		TaskID: taskID,
		Spec:   cntSums[0],
	}

	global.GNodes.N[host] = chosenNode

	go func() {
		t := time.NewTicker(3 * time.Second)
		for {
			<-t.C
			cnt, err := chosenNode.Client.ContainerInspect(context.Background(), cnt.ID)
			if err != nil {
				break
			}
			if cnt.State.Running {
				go tsk.PostAction(host, id, taskData.PostStart)
				break
			}
		}
		t.Stop()
	}()

	return struct {
		Status string `json:"status"`
		ID     string `json:"id"`
	}{fmt.Sprintf("task launched successfully; MAY STILL BE PENDING, you may want to check /get later"), id}, err
}

func RemoveTask(input string) (any, error) {
	global.GNodes.M.Lock()
	defer global.GNodes.M.Unlock()

	for nHost := range global.GNodes.N {
		for id := range global.GNodes.N[nHost].Containers {
			if id == input {
				err := global.GNodes.N[nHost].Client.ContainerRemove(
					context.Background(),
					global.GNodes.N[nHost].Containers[id].Spec.ID,
					container.RemoveOptions{
						Force: true,
					},
				)
				if err != nil {
					return "", err
				}

				delete(global.GNodes.N[nHost].Containers, id)

				return struct {
					Status string `json:"status"`
					ID     string `json:"id"`
				}{"successfully removed task", id}, nil
			}
		}
	}

	return "", errors.New("task not found")
}
