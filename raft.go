package raft

import (
	"log"
	"strconv"
	"time"
)

// State represent the current state of a raft node
type State uint32

const (
	// FOLLOWER state
	FOLLOWER State = iota
	// CANDIDATE state
	CANDIDATE
	// LEADER state
	LEADER
	// SHUTDOWN satate
	SHUTDOWN
)

// Raft node
type Raft struct {
	state           State
	listenTCPPort   int
	electionTimeOut time.Duration
	lastEntry       int64
	currentTerm     int64
	clusterNodes    []string
	leader          string
	voteRequestCh   chan RequestVoteRequest
	votesReceivedCh chan RequestVoteResponse
}

func (r *Raft) setState(state State) {
	r.state = state
}

func (r *Raft) getState() State {
	return r.state
}

func (r *Raft) setTerm(term int64) {
	r.currentTerm = term
}

func (r *Raft) getTerm() int64 {
	return r.currentTerm
}

func (r *Raft) setLeader(leader string) {
	r.leader = leader
}

// GetLeader return leader address
func (r *Raft) GetLeader() string {
	return r.leader
}

func (r *Raft) processVoteRequest(req RequestVoteRequest) RequestVoteResponse {
	resp := RequestVoteResponse{
		Granted: true,
	}
	return resp
}

func (r *Raft) runFollower() {
	for r.getState() == FOLLOWER {
		if r.lastEntry+r.electionTimeOut.Nanoseconds() < time.Now().UnixNano() {
			r.setState(CANDIDATE)
		}
	}
}

func (r *Raft) runLeader() {
	for r.getState() == LEADER {
		time.Sleep(100)
	}
}

func (r *Raft) runCandidate() {
	var voteRequested = false
	var receivedVotes = 0

	majority := len(r.clusterNodes)/2 + 1

	for r.getState() == CANDIDATE {
		if !voteRequested {
			r.requestVoteRequest()
			voteRequested = true
		}

		select {
		case req := <-r.voteRequestCh:
			req.Response <- r.processVoteRequest(req)
		case vote := <-r.votesReceivedCh:
			if vote.Granted {
				receivedVotes++
			}
			if receivedVotes >= majority {
				log.Printf("Leader elected")
				r.setState(LEADER)
				r.setLeader(":" + strconv.Itoa(r.listenTCPPort))
			}
		}
	}
}

func (r *Raft) runFSM() {
	for {
		switch r.getState() {
		case FOLLOWER:
			r.runFollower()
		case CANDIDATE:
			r.runCandidate()
		case LEADER:
			r.runLeader()
		}
	}
}

// NewRaft creates a new Raft node
func NewRaft(listenPort int, clusterNodes []string) *Raft {
	raftNode := Raft{}
	raftNode.listenTCPPort = listenPort
	raftNode.electionTimeOut = 10 * time.Second
	raftNode.lastEntry = time.Now().UnixNano()
	raftNode.currentTerm = 1
	raftNode.clusterNodes = clusterNodes
	raftNode.voteRequestCh = make(chan RequestVoteRequest)
	raftNode.votesReceivedCh = make(chan RequestVoteResponse)
	// start raft node as follower
	raftNode.setState(FOLLOWER)
	go raftNode.listen()
	go raftNode.runFSM()

	return &raftNode
}
