package main

import "net/http"

type dataResponse struct {
	NumberOfRequest int `json:"numberOfRequests"`
}

func (sm *smServer) dataHandler(w http.ResponseWriter, r *http.Request) {
	err := replyJSON(w, dataResponse{
		NumberOfRequest: sm.counterNumberOfRequests,
	}, &sm.logger)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
	}
}
