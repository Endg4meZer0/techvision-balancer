package server

import (
	"net/http"

	"techvision/balancer/server/handlers"
)

func SetupServer() {
	http.HandleFunc("/get", handlers.Get)
	http.HandleFunc("/addTask", handlers.AddTask)
	http.HandleFunc("/removeTask", handlers.RemoveTask)
	http.ListenAndServe("0.0.0.0:8081", nil)
}
