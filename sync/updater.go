package sync

import (
	"context"
	"techvision/balancer/docker"
	"techvision/balancer/global"
	"techvision/balancer/tasks"
	"time"

	"github.com/docker/docker/api/types/container"
)

var UpdateTicker *time.Ticker = time.NewTicker(time.Second * 10)

func init() {
	go func() {
		for {
			select {
			case <-UpdateTicker.C:
				// Update nodes
				global.GNodes.M.Lock()
				defer global.GNodes.M.Unlock()

				for nodeID := range global.GNodes.N {
					node := global.GNodes.N[nodeID]
					containers, err := node.Client.ContainerList(context.Background(), container.ListOptions{
						All: true,
					})
					if err != nil {
						continue
					}
					for _, cont := range containers {
						if cnt, ok := node.Containers[cont.ID]; ok {
							cnt.Spec = cont
							node.Containers[cont.ID] = cnt
						}
					}

					global.GNodes.N[nodeID] = node
				}
				docker.OnUpdate()
				for _, task := range tasks.Tasks {
					task.OnUpdate()
				}
			}
		}
	}()
}
