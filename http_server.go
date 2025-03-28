package main

import (
	"encoding/json"
	"io"
	"net/http"
)

type ErrorResponse struct {
	Error string `json:"error"`
}

type JsonInput struct {
	TaskId string `json:"taskId"`
}

func SetupServer() {
	http.HandleFunc("/get", onGet)
	http.HandleFunc("/addTask", onAddTask)
	http.HandleFunc("/removeTask", onRemoveTask)
	http.ListenAndServe("localhost:8081", nil)
}

func onGet(w http.ResponseWriter, r *http.Request) {
	res, err := GlobalNodes.Get()
	if err != nil {
		errResp, _ := json.Marshal(ErrorResponse{"an error occurred: " + err.Error()})
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(errResp)
		return
	}

	resp, _ := json.Marshal(res)
	w.Write(resp)
}

func onAddTask(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		errResp, _ := json.Marshal(ErrorResponse{"body is not readable"})
		w.WriteHeader(http.StatusBadRequest)
		w.Write(errResp)
		return
	}

	var input JsonInput
	err = json.Unmarshal(body, &input)
	if err != nil {
		errResp, _ := json.Marshal(ErrorResponse{"body is not valid"})
		w.WriteHeader(http.StatusBadRequest)
		w.Write(errResp)
		return
	}

	res, err := GlobalNodes.AddTask(input.TaskId)
	if err != nil {
		errResp, _ := json.Marshal(ErrorResponse{"an error occurred: " + err.Error()})
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(errResp)
		return
	}

	resp, _ := json.Marshal(struct {
		Res string `json:"res"`
	}{res})
	w.Write(resp)
}

func onRemoveTask(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		errResp, _ := json.Marshal(ErrorResponse{"body is not readable"})
		w.WriteHeader(http.StatusBadRequest)
		w.Write(errResp)
		return
	}

	var input JsonInput
	err = json.Unmarshal(body, &input)
	if err != nil {
		errResp, _ := json.Marshal(ErrorResponse{"body is not valid"})
		w.WriteHeader(http.StatusBadRequest)
		w.Write(errResp)
		return
	}

	res, err := GlobalNodes.RemoveTask(input.TaskId)
	if err != nil {
		errResp, _ := json.Marshal(ErrorResponse{"an error occurred: " + err.Error()})
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(errResp)
		return
	}

	resp, _ := json.Marshal(struct {
		Res string `json:"res"`
	}{res})
	w.Write(resp)
}
