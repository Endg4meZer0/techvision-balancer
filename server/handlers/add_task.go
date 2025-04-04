package handlers

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"techvision/balancer/docker"
	"techvision/balancer/global"
)

type addTaskJSONInput struct {
	TaskID string          `json:"task_id,omitempty"`
	Data   global.JSONData `json:"data"`
}

func AddTask(w http.ResponseWriter, r *http.Request) {
	log.Println("received onAddTask (/addTask)")

	body, err := io.ReadAll(r.Body)
	if err != nil {
		errResp, _ := json.Marshal(global.ErrorResponse{Error: "body is not readable"})
		w.WriteHeader(http.StatusBadRequest)
		w.Write(errResp)
		return
	}
	log.Println("onAddTask body: " + string(body))

	body = global.DoubleEscapeCheck(body)

	var input addTaskJSONInput
	err = json.Unmarshal(body, &input)
	if err != nil || input.TaskID == "" {
		errResp, _ := json.Marshal(global.ErrorResponse{Error: "body is not valid"})
		w.WriteHeader(http.StatusBadRequest)
		w.Write(errResp)
		return
	}

	res, err := docker.AddTask(input.TaskID, input.Data)
	if err != nil {
		errResp, _ := json.Marshal(global.ErrorResponse{Error: "an error occurred: " + err.Error()})
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(errResp)
		return
	}

	resp, _ := json.Marshal(res)
	w.Write(resp)
}
