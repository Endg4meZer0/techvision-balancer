package tasks

import (
	"techvision/balancer/tasks/tasks"
	"techvision/balancer/tasks/types"
)

type Task interface {
	GetSpec() types.TaskSpec
	PostAction(host string, containerID string, data interface{}) error
	OnUpdate()
}

var Tasks map[string]Task = map[string]Task{
	"1":       tasks.T1,
	"startCV": tasks.TStartCV,
}
