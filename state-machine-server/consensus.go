package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"sync"
	"time"

	"github.com/mitchellh/mapstructure"
)

type heartBeatRequest struct {
	LeaderID   int    `json:"leaderID"`
	LeaderAddr string `json:"leaderAddr"`
	LeaderTerm int    `json:"leaderTerm"`
}

type heatBeatResponse struct {
	CurrentTerm int  `json:"currentTerm"`
	Success     bool `json:"success"`
}

type voteRequest struct {
	CandidateID int `json:"candidateID"`
	Term        int `json:"term"`
}

type voteResponse struct {
	Term        int  `json:"term"`
	VoteGranted bool `json:"voteGranted"`
}

func (sm *smServer) heartBeatHandler(w http.ResponseWriter, r *http.Request) {
	// new leader or old leader
	var parsed heartBeatRequest
	err := json.NewDecoder(r.Body).Decode(&parsed)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	sm.logger.Printf("New heartbeat from %d, term %d", parsed.LeaderID, parsed.LeaderTerm)
	if parsed.LeaderTerm < sm.currentTerm {
		sm.logger.Printf("Heartbeat rejected: leaderTerm %d is lower than currentTerm %d", parsed.LeaderTerm, sm.currentTerm)
		replyJSON(w, heatBeatResponse{
			CurrentTerm: sm.currentTerm,
			Success:     false,
		}, &sm.logger)
		return
	}
	sm.hbReceived = true
	sm.leaderID = parsed.LeaderID
	sm.leaderAddr = parsed.LeaderAddr
	sm.currentTerm = parsed.LeaderTerm
	sm.status = 2 // switch to follower / stay as follower
	sm.logger.Printf("Heartbeat accepted")
	replyJSON(w, heatBeatResponse{
		CurrentTerm: sm.currentTerm,
		Success:     true,
	}, &sm.logger)
	return
}

func (sm *smServer) voteHandler(w http.ResponseWriter, r *http.Request) {
	// a candidate asks the node for a vote to become the leader
	var parsed voteRequest
	err := json.NewDecoder(r.Body).Decode(&parsed)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	sm.logger.Printf("New Vote request from %d, term %d", parsed.CandidateID, parsed.Term)
	if parsed.Term < sm.currentTerm {
		sm.logger.Printf("Vote request rejected: candidateTerm %d is lower than currentTerm %d", parsed.Term, sm.currentTerm)
		replyJSON(w, voteResponse{
			Term:        sm.currentTerm,
			VoteGranted: false,
		}, &sm.logger)
		return
	}
	sm.logger.Printf("Vote request accepted")
	sm.votedFor = parsed.CandidateID
	sm.currentTerm = parsed.Term
	sm.status = 2 // switch to follower, stay as follower
	replyJSON(w, voteResponse{
		Term:        sm.currentTerm,
		VoteGranted: true,
	}, &sm.logger)
	return

}

func (sm *smServer) launchTicker() {
	ticker := time.NewTicker(5 * time.Second)
	sm.timeout = ticker.C
	for {
		select {
		case <-ticker.C:
			if sm.status == 1 {
				// leader does not expect HB, but sends them
				sm.logger.Printf("Leader sends HB")
				sm.leaderSendsHB()
			} else {
				if sm.hbReceived {
					sm.logger.Printf("Ticker ticked, with heartbeat received")
					sm.hbReceived = false
				} else {
					sm.logger.Printf("Ticker ticked, with heartbeat not received")
					sm.apply()
				}
			}
		}
	}
}

func (sm *smServer) apply() {
	// apply to become the leader, change status immediately to candidate
	numberOfVotes := 1 //the server votes for itself
	sm.status = 3
	sm.currentTerm++
	var wg sync.WaitGroup
	for id, addr := range sm.sys.Addresses {
		if id == sm.ID {
			continue
		}
		wg.Add(1)
		go func(i int, addr string) {
			defer wg.Done()
			if sm.requestVote(addr) {
				numberOfVotes++
			}
		}(id, addr)
	}
	wg.Wait()
	if numberOfVotes <= sm.sys.NumberOfNodes/2 { // no strict majority
		// not a leader
		sm.status = 2
		return
	}
	sm.leaderSendsHB()
}

func (sm *smServer) leaderSendsHB() {
	// new leader
	sm.status = 1
	sm.leaderAddr = sm.addr
	sm.leaderID = sm.ID
	var wg sync.WaitGroup
	numOfValidations := 0
	doFollow := make([]int, 0)
	for id, addr := range sm.sys.Addresses {
		if id == sm.ID {
			continue
		}
		wg.Add(1)
		go func(i int, addr string, df *[]int) {
			defer wg.Done()
			if sm.sendHB(addr) {
				numOfValidations++
				*df = append(*df, i)
			}
		}(id, addr, &doFollow)
	}
	wg.Wait()
	if numOfValidations < sm.sys.NumberOfNodes-1 {
		sm.logger.Printf("Leader send HB process terminated, %d nodes do not follow : %v", sm.sys.NumberOfNodes-1-numOfValidations, doFollow)
		if globalConfig.UpdateSystem {
			sm.newSys(doFollow)
		}
	}
	sm.logger.Printf("Leader send HB process terminated, %d nodes follow", numOfValidations)
}

func (sm *smServer) requestVote(addr string) bool {
	var response voteResponse
	sm.logger.Printf("Request vote sent to %s", addr)
	sm.counterRequestsConsensus++
	resp, err := postJSON(addr+voteEndpoint, voteRequest{CandidateID: sm.ID, Term: sm.currentTerm}, &sm.logger, false)
	if err != nil {
		sm.logger.Printf("Error requesting vote at %s : %s", addr, err.Error())
		return false
	}

	res, err := decodeJSONResponse(resp, &sm.logger)

	err = mapstructure.Decode(res, &response)
	if err != nil {
		sm.logger.Printf("Error parsing vote from %s : %s", addr, err.Error())
		return false
	}

	return response.VoteGranted
}

func (sm *smServer) sendHB(addr string) bool {
	var response heatBeatResponse
	sm.counterRequestsConsensus++
	resp, err := postJSON(addr+heartbeatEndpoint, heartBeatRequest{LeaderID: sm.ID, LeaderAddr: sm.addr, LeaderTerm: sm.currentTerm}, &sm.logger, false)

	if err != nil {
		sm.logger.Printf("ERROR SENDING HB at %s : %s", addr, err.Error())
		return false
	}

	res, err := decodeJSONResponse(resp, &sm.logger)

	if err != nil {
		sm.logger.Printf("Error decoding JSON HB response at %s : %s", addr, err.Error())
		return false
	}

	err = mapstructure.Decode(res, &response)
	if err != nil {
		sm.logger.Printf("Error parsing vote from %s : %s", addr, err.Error())
		return false
	}
	return response.Success
}

func (sm *smServer) newSys(doFollow []int) {
	fmt.Printf("new sys called with doFollow %v and addresses %v", doFollow, sm.sys.Addresses)
	sort.Slice(doFollow, func(i, j int) bool {
		return i < j
	})
	// doFollow contains the ids of the following nodes, sorted

	// self-update
	nbFollowers := len(doFollow)
	sm.sys.NumberOfNodes = nbFollowers + 1
	newAddresses := make(map[int]string)
	newAddresses[sm.ID] = sm.addr
	for i := 0; i < nbFollowers; i++ {
		// sm.sys.addresses[doFollow[i]] is the address of the node (or addr for the leader)
		newAddresses[doFollow[i]] = sm.sys.Addresses[doFollow[i]]
	}
	sm.sys.Addresses = newAddresses
	sm.logger.Printf("New addresses list in system : %v", newAddresses)

	// update all followers
	var wg sync.WaitGroup
	for id, addr := range sm.sys.Addresses {
		if id == sm.ID {
			continue
		}
		wg.Add(1)
		go func(addr string) {
			defer wg.Done()
			sm.counterRequestsConsensus++
			postJSON(addr+updateSysEndpoint, sm.sys, &sm.logger, false)
		}(addr)
	}
	wg.Wait()
}

func (sm *smServer) updateSysHandler(w http.ResponseWriter, r *http.Request) {
	var parsed system
	err := json.NewDecoder(r.Body).Decode(&parsed)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	sm.logger.Printf("New system : %v", parsed)

	sm.sys = parsed

	return
}
