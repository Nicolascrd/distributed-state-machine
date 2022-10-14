package main

import (
	"math/rand"
	"sync"
	"time"

	"github.com/mitchellh/mapstructure"
)

func (sm *smServer) initiateQuery(req addLogReq) (bool, error) {
	// pick a sample of the network

	s := sampleNetwork(sm.ID, sm.sys.NumberOfNodes)
	sm.logger.Printf("Initiating query with sample %v", s)

	var wg sync.WaitGroup
	var nbQueries int
	var nbErrors int
	var nbSuccess int
	for i := range s {
		wg.Add(1)
		go func(add string, nbErr *int, nbSucc *int, nbQuer *int) {
			defer wg.Done()
			*nbQuer++
			success, err := sm.queryNode(add, req)
			if err != nil {
				*nbErr++
			} else if success {
				*nbSucc++
			}
		}(sm.sys.Addresses[s[i]], &nbErrors, &nbSuccess, &nbQueries)
	}
	wg.Wait()

	sm.logger.Printf("Queried %d nodes with %d errors and %d successes", nbQueries, nbErrors, nbSuccess)
	return nbSuccess >= globalConfig.MajorityThreshold, nil
}

func (sm *smServer) queryNode(addr string, req addLogReq) (bool, error) {
	// return true, nil if reply with the same value
	// return false, nil if reply with any other value
	resp, err := postJSON(addr+addLogEndpoint, req, &sm.logger, false)
	if err != nil {
		sm.logger.Printf("Error posting JSON at %s : %s", addr, err.Error())
		return false, err
	}
	ans, err := decodeJSONResponse(resp, &sm.logger)
	if err != nil {
		sm.logger.Printf("Cannot decode JSON response : %s", err.Error())
		return false, err
	}
	var addLogAns addLogAnswer
	err = mapstructure.Decode(ans, &addLogAns)
	if err != nil {
		sm.logger.Printf("Cannot decode JSON response to addLogAnswer format : %s", err.Error())
		return false, err
	}
	return addLogAns.Success, nil
}

func sampleNetwork(excludedID int, numberOfNodes int) []int {
	s := make([]int, numberOfNodes-1)
	var index int
	for id := 1; id <= numberOfNodes; id++ { // nodes number start at 1
		if id == excludedID {
			continue
		}
		s[index] = id
		index++
	}
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(s), func(i, j int) { s[i], s[j] = s[j], s[i] })
	s = s[:globalConfig.SampleSize]
	return s
}
