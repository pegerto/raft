package raft

import (
	"net"
	"net/http"
	"net/rpc"
	"strconv"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
)

type replicator struct {
	client *rpc.Client
	node   string
}

func newReplicator(node string) replicator {
	for {
		client, err := rpc.DialHTTP("tcp", node)
		if err != nil {
			log.Warn("Connection error: ", err)
			time.Sleep(5 * time.Second)
			continue
		}

		return replicator{client, node}
	}
}
func (r *replicator) replicate(entries AppendEntriesRequest) {
	var resp = AppendEntriesResponse{}
	err := r.client.Call("RPCServer.AppendEntries", entries, &resp)
	if err != nil {
		log.Warnf("Error requesting vote to node %s: %s", r.node, err)
	}
}

func (r *Raft) requestVoteRequest() {
	r.setTerm(r.getTerm() + 1)

	requestVoteRequest := RequestVoteRequest{
		Term:      r.getTerm(),
		Candidate: ":" + strconv.Itoa(r.listenTCPPort),
	}

	sendRequestVote := func(node string) {
		client, err := rpc.DialHTTP("tcp", node)
		if err != nil {
			log.Warnf("Connection error: %s", err)
			return
		}
		var resp = RequestVoteResponse{}
		err = client.Call("RPCServer.RequestVote", requestVoteRequest, &resp)
		if err != nil {
			log.Warnf("Error requesting vote to node %s: %s", node, err)
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
	log.Debug("Request vote received")
	voteRequest.Response = make(chan RequestVoteResponse)
	s.raft.voteRequestCh <- voteRequest
	resp := <-voteRequest.Response
	*response = resp
	log.Debugf("Request vote reponded node: %d, granted %t", s.raft.listenTCPPort, response.Granted)
	return nil
}

// AppendEntries implements append entry RPC node
func (s *RPCServer) AppendEntries(appendRequest AppendEntriesRequest, response *AppendEntriesResponse) error {
	s.raft.appendEntriesCh <- appendRequest
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
	log.Infof("Serving RPC server on port %d", r.listenTCPPort)

	err := http.Serve(listener, mux)
	if err != nil {
		log.Fatal("Error serving: ", err)
	}
}
