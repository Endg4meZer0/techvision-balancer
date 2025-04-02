package main

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"techvision/balancer/docker"
)

type ErrorResponse struct {
	Error string `json:"error"`
}

func SetupServer() {
	http.HandleFunc("/get", onGet)
	http.HandleFunc("/addTask", onAddTask)
	http.HandleFunc("/removeTask", onRemoveTask)
	http.ListenAndServe("localhost:8081", nil)
}

func onGet(w http.ResponseWriter, r *http.Request) {
	log.Println("received onGet (/get)")
	res, err := docker.Get()
	if err != nil {
		errResp, _ := json.Marshal(ErrorResponse{"an error occurred: " + err.Error()})
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(errResp)
		return
	}

	w.Write([]byte(res))
}

func onAddTask(w http.ResponseWriter, r *http.Request) {
	log.Println("received onAddTask (/addTask)")

	body, err := io.ReadAll(r.Body)
	if err != nil {
		errResp, _ := json.Marshal(ErrorResponse{"body is not readable"})
		w.WriteHeader(http.StatusBadRequest)
		w.Write(errResp)
		return
	}
	log.Println("onAddTask body: " + string(body))

	var input JSONInput
	err = json.Unmarshal(body, &input)
	if err != nil || input.TaskID == "" {
		errResp, _ := json.Marshal(ErrorResponse{"body is not valid"})
		w.WriteHeader(http.StatusBadRequest)
		w.Write(errResp)
		return
	}

	res, err := docker.AddTask(input.TaskID)
	if err != nil {
		errResp, _ := json.Marshal(ErrorResponse{"an error occurred: " + err.Error()})
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(errResp)
		return
	}

	resp, _ := json.Marshal(res)
	w.Write(resp)
}

func onRemoveTask(w http.ResponseWriter, r *http.Request) {
	log.Println("received onRemoveTask (/removeTask)")

	body, err := io.ReadAll(r.Body)
	if err != nil {
		errResp, _ := json.Marshal(ErrorResponse{"body is not readable"})
		w.WriteHeader(http.StatusBadRequest)
		w.Write(errResp)
		return
	}
	log.Println("onRemoveTask body: " + string(body))

	var input JSONInput
	err = json.Unmarshal(body, &input)
	if err != nil {
		errResp, _ := json.Marshal(ErrorResponse{"body is not valid"})
		w.WriteHeader(http.StatusBadRequest)
		w.Write(errResp)
		return
	}

	var res any
	if input.TaskID != "" {
		res, err = docker.RemoveTask(input.TaskID)
		if err != nil {
			errResp, _ := json.Marshal(ErrorResponse{"an error occurred: " + err.Error()})
			w.WriteHeader(http.StatusInternalServerError)
			w.Write(errResp)
			return
		}
	} else if input.ContainerID != "" {
		res, err = docker.RemoveContainer(input.ContainerID)
		if err != nil {
			errResp, _ := json.Marshal(ErrorResponse{"an error occurred: " + err.Error()})
			w.WriteHeader(http.StatusInternalServerError)
			w.Write(errResp)
			return
		}
	}

	resp, _ := json.Marshal(res)
	w.Write(resp)
}
