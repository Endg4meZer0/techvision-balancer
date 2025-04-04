package handlers

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"techvision/balancer/docker"
	"techvision/balancer/global"
)

type removeTaskJSONInput struct {
	ID string `json:"id"`
}

func RemoveTask(w http.ResponseWriter, r *http.Request) {
	log.Println("received onRemoveTask (/removeTask)")

	body, err := io.ReadAll(r.Body)
	if err != nil {
		errResp, _ := json.Marshal(global.ErrorResponse{Error: "body is not readable"})
		w.WriteHeader(http.StatusBadRequest)
		w.Write(errResp)
		return
	}
	log.Println("onRemoveTask body: " + string(body))

	var input removeTaskJSONInput
	err = json.Unmarshal(body, &input)
	if err != nil {
		errResp, _ := json.Marshal(global.ErrorResponse{Error: "body is not valid (unmarshal error)"})
		w.WriteHeader(http.StatusBadRequest)
		w.Write(errResp)
		return
	}

	if input.ID == "" {
		errResp, _ := json.Marshal(global.ErrorResponse{Error: "body is not valid (id is empty)"})
		w.WriteHeader(http.StatusBadRequest)
		w.Write(errResp)
		return
	}

	res, err := docker.RemoveTask(input.ID)
	if err != nil {
		errResp, _ := json.Marshal(global.ErrorResponse{Error: "an error occurred: " + err.Error()})
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(errResp)
		return
	}

	resp, _ := json.Marshal(res)
	w.Write(resp)
}
