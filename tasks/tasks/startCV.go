package tasks

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"techvision/balancer/global"
	"techvision/balancer/tasks/types"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/go-connections/nat"
	"github.com/docker/go-units"
)

type TaskStartCV struct {
	Spec types.TaskSpec
}

var TStartCV = TaskStartCV{
	Spec: types.TaskSpec{
		ContainerConfig: container.Config{
			Image: "cvunit-work:1.3",
			ExposedPorts: nat.PortSet{
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
				"4950/tcp": []nat.PortBinding{{HostPort: "4950"}},
			},
		},
		NetworkingConfig: network.NetworkingConfig{},
		Name:             "cvunit-work",
	},
}

func (t TaskStartCV) GetSpec() types.TaskSpec {
	return t.Spec
}

func (t TaskStartCV) PostAction(host string, id string, data any) error {
	node := global.GNodes.N[host]
	jsonTransferred := false
	ticker := time.NewTicker(3 * time.Second)
	for !jsonTransferred {
		<-ticker.C
		cnt := node.Containers[id]

		if cnt.Spec.State == container.Starting {
			continue
		} else if cnt.Spec.Status == "exited" {
			return errors.New("container exited before post-action")
		}

		// logsR, err := node.Client.ContainerLogs(context.Background(), cnt.Spec.ID, container.LogsOptions{
		// 	ShowStdout: true,
		// })
		// if err != nil {
		// 	break
		// }
		// logs, err := io.ReadAll(logsR)
		// if err != nil {
		// 	break
		// }
		// if !strings.Contains(string(logs), "Press CTRL+C to quit") {
		// 	continue
		// }

		var extPort uint16
		for _, p := range cnt.Spec.Ports {
			if p.PrivatePort == 4950 && !strings.ContainsRune(p.IP, ':') {
				extPort = p.PublicPort
				break
			}
		}
		actualHost := strings.Split(host, ":")[0]

		_, err := http.Post(fmt.Sprintf("http://%s:%v/create", actualHost, extPort), "application/json", bytes.NewReader((data.(json.RawMessage))))
		if err != nil {
			return err
		}
		jsonTransferred = true
	}

	ticker.Stop()
	return nil
}
