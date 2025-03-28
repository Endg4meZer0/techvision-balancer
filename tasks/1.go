package tasks

import "github.com/docker/docker/api/types/swarm"

// Service specification
var task1 = swarm.ServiceSpec{
	Annotations: swarm.Annotations{
		Name: "ping-service",
		Labels: map[string]string{
			"com.example.description": "Service that pings docker.com",
		},
	},
	TaskTemplate: swarm.TaskSpec{
		ContainerSpec: &swarm.ContainerSpec{
			Image: "alpine",
			Command: []string{
				"ping",
				"docker.com",
			},
		},
		RestartPolicy: &swarm.RestartPolicy{
			Condition: swarm.RestartPolicyConditionAny,
		},
	},
	Mode: swarm.ServiceMode{
		Replicated: &swarm.ReplicatedService{
			Replicas: uintPtr(1),
		},
	},
}