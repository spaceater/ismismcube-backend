package handler

import (
	"encoding/json"
	"ismismcube-backend/internal/model"
	"net/http"
)

type PageViewResponse struct {
	PageView int `json:"page_view"`
}

func PageViewHandler(w http.ResponseWriter, r *http.Request) {
	pageView, err := model.PageView.GetAndIncrement()
	if err != nil {
		sendResponse(w, PageViewResponse{PageView: -1})
		return
	}
	sendResponse(w, PageViewResponse{PageView: pageView})
}

type ExecutedTaskResponse struct {
	ExecutedTask int `json:"executed_task"`
}

func ExecutedTaskHandler(w http.ResponseWriter, r *http.Request) {
	executedTask, err := model.ExecutedTask.Get()
	if err != nil {
		sendResponse(w, ExecutedTaskResponse{ExecutedTask: -1})
		return
	}
	sendResponse(w, ExecutedTaskResponse{ExecutedTask: executedTask})
}

func sendResponse(w http.ResponseWriter, response interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
