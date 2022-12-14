package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"

	"github.com/mitchellh/mapstructure"
)

type requestLogReq struct {
	Position int `json:"position"`
}

type addLogReq struct {
	Position int    `json:"position"`
	Content  string `json:"content"`
	Internal bool   `json:"internal"`
}

type addLogAnswer struct {
	Success bool `json:"success"`
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
		http.Error(w, fmt.Sprintf("No entry for position %d", req.Position), http.StatusNotFound)
		return
	}
	io.WriteString(w, log)
	return
}

func (sm *smServer) addLogHandler(w http.ResponseWriter, r *http.Request) {
	var req addLogReq
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	sm.logger.Printf("New add log request, position %d", req.Position)

	_, ok := sm.record[req.Position]
	if ok {
		sm.logger.Printf("Map already contains a log at position %d", req.Position)
		replyJSON(w, addLogAnswer{
			Success: false,
		}, &sm.logger)
		return
	}
	if sm.status == 1 {
		// node is leader
		sm.logger.Printf("Node is leader and the new record was saved at position %d", req.Position)
		sm.record[req.Position] = req.Content
		sm.transferNewLog(req)
		replyJSON(w, addLogAnswer{
			Success: true,
		}, &sm.logger)
		return
	}
	// node is not leader
	if req.Internal {
		sm.logger.Printf("Node is follower, receiving a request from the leader and the new record was saved at position %d", req.Position)
		// request comes from leader, just update node state
		sm.record[req.Position] = req.Content
		replyJSON(w, addLogAnswer{
			Success: true,
		}, &sm.logger)
		return
	}
	// request comes from client, transfer to leader
	sm.logger.Printf("Node is not leader and the request was transmitted")
	success, err := sm.transferNewLogRequest(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if !success {
		err = replyJSON(w, addLogAnswer{
			Success: false,
		}, &sm.logger)
	} else {
		err = replyJSON(w, addLogAnswer{Success: true}, &sm.logger)
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	return
}

func (sm *smServer) transferNewLogRequest(req addLogReq) (bool, error) {
	// for followers to transfer the new log request to leader
	req.Internal = true
	sm.counterRequestsClient++
	resp, err := postJSON(sm.leaderAddr+addLogEndpoint, req, &sm.logger, true)
	if err != nil {
		return false, err
	}
	ans, err := decodeJSONResponse(resp, &sm.logger)
	if err != nil {
		return false, err
	}
	var answer addLogAnswer
	err = mapstructure.Decode(ans, &answer)
	if err != nil {
		return false, err
	}
	return answer.Success, nil
}

func (sm *smServer) transferNewLog(req addLogReq) {
	// for leader to transfer the new log entry to update followers
	var wg sync.WaitGroup
	var numPosts int
	var numErrors int
	req.Internal = true
	for _, addr := range sm.sys.Addresses {
		if addr == sm.addr {
			continue
		}
		wg.Add(1)
		go func(ad string, numPosts *int, numErrors *int) {
			defer wg.Done()
			sm.counterRequestsClient++
			_, err := postJSON(ad+addLogEndpoint, req, &sm.logger, false)
			if err != nil {
				*numErrors++
			}
		}(addr, &numPosts, &numErrors)
	}
	wg.Wait()
	sm.logger.Printf("Transfered new log to %d followers, including %d errors", numPosts, numErrors)
}
