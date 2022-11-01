package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
)

type system struct {
	NumberOfNodes int            `json:"numberOfNodes"` // number of nodes in the whole system
	Addresses     map[int]string `json:"addresses"`     // ports of all nodes in order (including this one)
}

type smServer struct {
	logger                  log.Logger     // associated logger
	addr                    string         // URL in container eg centra-calcu-1:8000
	ID                      int            // server number e.g. 1
	record                  map[int]string // the distributed record
	cnt                     map[int]int    // the counter (one per position)
	counterNumberOfRequests int            // count number of requests to track efficiency
	sys                     system         // each node knows the system
}

type config struct {
	MajorityThreshold int `json:"majorityThreshold"`
	SampleSize        int `json:"sampleSize"`
	CounterThreshold  int `json:"counterThreshold"`
}

var globalConfig config

func main() {
	fmt.Println("Hello State Machine")
	args := os.Args[1:]
	if len(args) != 2 {
		fmt.Println("Wrong number of arguments in command line, expecting only 2 numbers between 0 and 99")
		return
	}

	ind, err := strconv.Atoi(args[0])
	if err != nil {
		fmt.Println("First argument provided should be an int but \n " + err.Error())
		return
	}
	if ind < 0 || ind > 99 {
		fmt.Println("First Number given is out of bounds ([0,99])")
		return
	}
	tot, err := strconv.Atoi(args[1])
	if err != nil {
		fmt.Println("Second argument provided should be an int but \n" + err.Error())
		return
	}
	if tot < 0 || tot > 99 {
		fmt.Println("Second Number given is out of bounds ([0,99])")
		return
	}
	configFile, err := os.Open("config.json")

	if err != nil {
		fmt.Println("Could not open config json : " + err.Error())
		return
	}
	jsonParser := json.NewDecoder(configFile)
	jsonParser.Decode(&globalConfig)
	if err != nil {
		fmt.Println("Could not decode config json : " + err.Error())
		return
	}
	configFile.Close()
	fmt.Println("config : ", globalConfig)
	sm := newStateMachineServer(ind, tot)

	sm.launchStateMachineServer()
}

func newStateMachineServer(num int, tot int) *smServer {
	// num : id of this server
	// tot : total number of servers

	l := log.New(log.Writer(), "SMServer - "+fmt.Sprint(num)+"  ", log.Ltime)

	addresses := make(map[int]string)
	for i := 1; i <= tot; i++ {
		addresses[i] = buildAddress(i)
	}
	sys := system{
		NumberOfNodes: tot,
		Addresses:     addresses,
	}

	return &smServer{
		logger: *l,
		ID:     num,
		addr:   buildAddress(num),
		sys:    sys,
		record: make(map[int]string),
		cnt:    make(map[int]int),
	}
}

func buildAddress(num int) string {
	return "sm-server-" + fmt.Sprint(num) + ":8000"
}
