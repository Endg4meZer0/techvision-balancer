package sync

import (
	"context"
	"strings"
	"techvision/balancer/global"
	"time"

	"github.com/docker/docker/api/types/container"
)

var UpdateTicker *time.Ticker = time.NewTicker(time.Second * 5)

func init() {
	go func() {
		for {
			<-UpdateTicker.C
			global.GNodes.M.Lock()

			// Update nodes
			for nodeID := range global.GNodes.N {
				node := global.GNodes.N[nodeID]
				containers, err := node.Client.ContainerList(context.Background(), container.ListOptions{
					All: false,
				})
				if err != nil {
					continue
				}
				for _, cont := range containers {
					for k, cnt := range node.Containers {
						if strings.Contains(k, cont.ID) {
							cnt.Spec = cont
							node.Containers[k] = cnt
						}
					}
				}

				global.GNodes.N[nodeID] = node
			}

			global.GNodes.M.Unlock()
		}
	}()
}
