package tasks

import (
	"techvision/balancer/tasks/types"

	"github.com/docker/docker/api/types/container"
)

type Task1 struct {
	Spec types.TaskSpec
}

var T1 = Task1{
	Spec: types.TaskSpec{
		ContainerConfig: container.Config{
			Image: "alpine",
			Cmd:   []string{"ping", "docker.com"},
			Labels: map[string]string{
				"gpu":          "false",
				"gpu.required": "0",
			},
		},
		ContainerHostConfig: container.HostConfig{
			RestartPolicy: container.RestartPolicy{
				Name:              "always",
				MaximumRetryCount: 0,
			},
		},
		Name: "ping-task",
	},
}

func (t Task1) GetSpec() types.TaskSpec {
	return t.Spec
}

func (t Task1) PostAction(host string, containerID string, data interface{}) error {
	return nil
}

func (t Task1) OnUpdate() {}
