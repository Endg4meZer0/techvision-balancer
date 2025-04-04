package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"techvision/balancer/docker"
	"techvision/balancer/global"
)

func Get(w http.ResponseWriter, r *http.Request) {
	log.Println("received onGet (/get)")
	res, err := docker.Get()
	if err != nil {
		errResp, _ := json.Marshal(global.ErrorResponse{Error: "an error occurred: " + err.Error()})
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(errResp)
		return
	}

	w.Write([]byte(res))
}
