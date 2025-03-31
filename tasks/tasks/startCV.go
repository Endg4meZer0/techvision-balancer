package tasks

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"techvision/balancer/global"
	"techvision/balancer/tasks/types"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/go-connections/nat"
	"github.com/docker/go-units"
)

var updated = make(chan bool)
var Active = false

type TaskStartCV struct {
	Spec types.TaskSpec
}

var TStartCV = TaskStartCV{
	Spec: types.TaskSpec{
		ContainerConfig: container.Config{
			Image: "docker.local:5000/cvunit:v0.1",
			ExposedPorts: nat.PortSet{
				"9092/tcp": struct{}{},
				"9093/tcp": struct{}{},
				"9094/tcp": struct{}{},
				"8554/tcp": struct{}{},
				"4950/tcp": struct{}{},
			},
			Labels: map[string]string{
				"gpu":          "true",
				"gpu.required": "10",
			},
		},
		ContainerHostConfig: container.HostConfig{
			Resources: container.Resources{
				Ulimits: []*container.Ulimit{
					&units.Ulimit{
						Name: "memlock",
						Hard: -1,
						Soft: -1,
					},
				},
				DeviceRequests: []container.DeviceRequest{
					{
						Driver:       "nvidia",
						Count:        -1,
						Capabilities: [][]string{{"gpu"}},
					},
				},
			},
			PortBindings: nat.PortMap{
				"9092/tcp": []nat.PortBinding{{HostPort: "9092"}},
				"9093/tcp": []nat.PortBinding{{HostPort: "9093"}},
				"9094/tcp": []nat.PortBinding{{HostPort: "9094"}},
				"8554/tcp": []nat.PortBinding{{HostPort: "8554"}},
				"4950/tcp": []nat.PortBinding{{HostPort: "4950"}},
			},
		},
		NetworkingConfig: network.NetworkingConfig{},
	},
}

func (t TaskStartCV) GetSpec() types.TaskSpec {
	return t.Spec
}

const jsonToTransfer string = `{
    "type": "create_process",
    "msg": {
      "parameters": {
        "cvmode": "car",
        "channel": 1,
        "port": 554,
        "ip": "10.40.16.27",
        "login": "admin",
        "password": "bvrn2022",
        "scene_number": 1
      },
      "events": [
        {
            "event_name": "all_frames",
            "event_actions": [
              "box_drawing","line_count", "record", "rtsp_server_stream", "logging"
            ],
            "parameters": {
              "lines": {
                "line0": [[750, 125], [930, 150]]
              },
              "FPS": 30,
              "timer": 600,
              "host_port_rtsp_server": "10.61.36.17:8554"
            }
          }
      ]
    }
  }
`

func (t TaskStartCV) PostAction(host string, containerID string, data interface{}) error {
	jsonTransferred := false
	Active = true
	for !jsonTransferred {
		<-updated
		for _, cnt := range global.GNodes.N[host].Containers {
			if cnt.Spec.State == container.Starting {
				continue
			} else if cnt.Spec.Status == "exited" {
				return errors.New("container exited before post-action")
			}

			var extPort uint16
			for _, p := range cnt.Spec.Ports {
				if p.PrivatePort == 4950 {
					extPort = p.PublicPort
					break
				}
			}

			if data == nil {
				_, err := http.Post(fmt.Sprintf("https://%s:%v/create", host, extPort), "application/json", strings.NewReader(jsonToTransfer))
				if err != nil {
					return err
				}
			} else {
				_, err := http.Post(fmt.Sprintf("https://%s:%v/create", host, extPort), "application/json", strings.NewReader(data.(string)))
				if err != nil {
					return err
				}
			}
			jsonTransferred = true
		}
	}

	Active = false
	return nil
}

func (t TaskStartCV) OnUpdate() {
	if Active {
		updated <- true
	}
}
