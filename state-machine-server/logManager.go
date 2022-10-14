package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type requestLogReq struct {
	Position int `json:"position"`
}

type addLogReq struct {
	Position int    `json:"position"`
	Content  string `json:"content"`
}

type addLogAnswer struct {
	Success      bool   `json:"success"` // true if the queried node replies with the same string
	ErrorMessage string `json:"errorMessage,omitempty"`
}

func (sm *smServer) requestLogHandler(w http.ResponseWriter, r *http.Request) {
	var req requestLogReq
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	sm.logger.Printf("New Log access request, position %d", req.Position)

	log, ok := sm.record[req.Position]
	if !ok {
		http.Error(w, fmt.Sprintf("No entry for position %d", req.Position), http.StatusOK)
		return
	}
	io.WriteString(w, log)
	return
}

func (sm *smServer) addLogHandler(w http.ResponseWriter, r *http.Request) {
	var req addLogReq
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		replyJSON(w, addLogAnswer{
			Success:      false,
			ErrorMessage: err.Error(),
		}, &sm.logger)
		return
	}
	sm.logger.Printf("New add log request, position %d, content %s", req.Position, req.Content)

	rec, ok := sm.record[req.Position]
	if ok {
		// already a value at the requested position
		// "a colored node simply responds with its own color"
		replyJSON(w, addLogAnswer{
			Success: rec == req.Content,
		}, &sm.logger)
		return
	}
	// "an uncolored node adopts the color of the query..."
	sm.record[req.Position] = req.Content

	// "... responds with that color ..."

	replyJSON(w, addLogAnswer{
		Success: true,
	}, &sm.logger)

	// "... and initiates its own query"
	go sm.loopInitiateQuery(req)
	return
}
