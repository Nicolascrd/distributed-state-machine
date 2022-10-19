package main

import "net/http"

type dataResponse struct {
	ClientRelatedRequests    int `json:"clientRelatedRequests"`
	ConsensusRelatedRequests int `json:"consensusRelatedRequests"`
}

func (sm *smServer) requestDataHandler(w http.ResponseWriter, r *http.Request) {
	err := replyJSON(w, dataResponse{
		ClientRelatedRequests:    sm.counterRequestsClient,
		ConsensusRelatedRequests: sm.counterRequestsConsensus,
	}, &sm.logger)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
