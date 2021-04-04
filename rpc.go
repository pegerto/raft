package raft

import (
	"log"
	"net"
	"net/http"
	"net/rpc"
	"strconv"
	"sync"
)

func (r *Raft) requestVoteRequest() {
	r.setTerm(r.getTerm() + 1)

	requestVoteRequest := RequestVoteRequest{
		Candidate: ":" + strconv.Itoa(r.listenTCPPort),
	}

	sendRequestVote := func(node string) {
		client, err := rpc.DialHTTP("tcp", node)
		if err != nil {
			log.Fatal("Connection error: ", err)
		}
		var resp = RequestVoteResponse{}
		err = client.Call("RPCServer.RequestVote", requestVoteRequest, &resp)
		if err != nil {
			log.Printf("Error requesting vote to node %s: %s", node, err)
		}
		r.votesReceivedCh <- resp
	}

	for _, node := range r.clusterNodes {
		log.Printf("Request vote to candidate: %s \n", node)
		go sendRequestVote(node)
	}

}

// RPCServer provide a network interface for raft nodes
type RPCServer struct {
	raft *Raft
}

// RequestVote request
func (s *RPCServer) RequestVote(voteRequest RequestVoteRequest, response *RequestVoteResponse) error {
	log.Printf("Request vote received")
	voteRequest.Response = make(chan RequestVoteResponse)
	s.raft.voteRequestCh <- voteRequest
	resp := <-voteRequest.Response
	*response = resp
	log.Printf("Request vote reponded node: %d, granted %t", s.raft.listenTCPPort, response.Granted)
	return nil
}

func (r *Raft) listen() {
	var mutex sync.Mutex
	mutex.Lock()

	serv := rpc.NewServer()
	var server = &RPCServer{
		raft: r,
	}
	serv.Register(server)

	// RPC by itself does not allow multiple RPC servers under
	// differnt ports, this is a workaround:
	// https://github.com/golang/go/issues/13395
	oldMux := http.DefaultServeMux
	mux := http.NewServeMux()
	http.DefaultServeMux = mux
	serv.HandleHTTP(rpc.DefaultRPCPath, rpc.DefaultDebugPath)
	http.DefaultServeMux = oldMux
	mutex.Unlock()

	listener, e := net.Listen("tcp", ":"+strconv.Itoa(r.listenTCPPort))
	if e != nil {
		log.Fatal("Listen error: ", e)
	}
	log.Printf("Serving RPC server on port %d", r.listenTCPPort)

	err := http.Serve(listener, mux)
	if err != nil {
		log.Fatal("Error serving: ", err)
	}
}
