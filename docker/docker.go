package docker

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"techvision/balancer/global"
	"techvision/balancer/tasks"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
)

var waitChan = make(chan bool)
var Active = false

func Get() (string, error) {
	return global.GNodes.ToJSON(), nil
}

func AddTask(taskId string) (any, error) {
	tsk, ok := tasks.Tasks[taskId]
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
	time.Sleep(1 * time.Second)
	err = chosenNode.Client.ContainerStart(context.Background(), cnt.ID, container.StartOptions{})
	if err != nil {
		return "", err
	}
	time.Sleep(1 * time.Second)
	cntSums, err := chosenNode.Client.ContainerList(context.Background(), container.ListOptions{
		All: true,
		Filters: filters.NewArgs(
			filters.Arg("id", cnt.ID),
		),
	})
	if err != nil {
		return "", err
	}

	chosenNode.Containers[cnt.ID] = global.Container{
		TaskID: taskId,
		Spec:   cntSums[0],
	}

	global.GNodes.N[host] = chosenNode

	go func() {
		Active = true
		for {
			<-waitChan
			cnt, err := chosenNode.Client.ContainerInspect(context.Background(), cnt.ID)
			if err != nil {
				break
			}
			if cnt.State.Running {
				go tsk.PostAction(host, cnt.ID, nil)
				break
			}
		}
		Active = false
	}()

	return struct {
		Status      string `json:"status"`
		ContainerID string `json:"container_id"`
	}{fmt.Sprintf("task launched successfully; MAY STILL BE PENDING, you may want to check /get later"), cnt.ID}, err
}

func RemoveTask(input string) (any, error) {
	global.GNodes.M.Lock()
	defer global.GNodes.M.Unlock()

	for nHost := range global.GNodes.N {
		for contId, cont := range global.GNodes.N[nHost].Containers {
			if cont.TaskID == input {
				err := global.GNodes.N[nHost].Client.ContainerRemove(context.Background(), contId, container.RemoveOptions{
					Force: true,
				})
				if err != nil {
					return "", err
				}
				delete(global.GNodes.N[nHost].Containers, contId)
			}
		}
	}

	return struct {
		Status string `json:"status"`
		TaskID string `json:"task_id"`
	}{"successfully removed all containers associated with task", input}, nil
}

func RemoveContainer(input string) (any, error) {
	global.GNodes.M.Lock()
	defer global.GNodes.M.Unlock()

	for nHost := range global.GNodes.N {
		for cntId := range global.GNodes.N[nHost].Containers {
			if cntId == input {
				err := global.GNodes.N[nHost].Client.ContainerRemove(
					context.Background(),
					cntId,
					container.RemoveOptions{
						Force: true,
					},
				)
				if err != nil {
					return "", err
				}
				delete(global.GNodes.N[nHost].Containers, cntId)

				return struct {
					Status      string `json:"status"`
					ContainerID string `json:"container_id"`
				}{"successfully removed container", cntId}, nil
			}
		}
	}

	return "", errors.New("container not found")
}

func OnUpdate() {
	if Active {
		waitChan <- true
	}
}
