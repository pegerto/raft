package raft

import (
	"strconv"
	"time"

	log "github.com/sirupsen/logrus"
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
	lastVoteTerm    int64
	currentTerm     int64
	clusterNodes    []string
	leader          string
	voteRequestCh   chan RequestVoteRequest
	votesReceivedCh chan RequestVoteResponse
	appendEntriesCh chan AppendEntriesRequest
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

func (r *Raft) quorum() int {
	return len(r.clusterNodes)/2 + 1
}

func (r *Raft) processVoteRequest(req RequestVoteRequest) RequestVoteResponse {
	resp := RequestVoteResponse{}
	log.Debugf("Procesing  vote request: %v", req)

	//  if the request vote term is longer than the current
	//  set the node as follower
	if r.getTerm() < req.Term {
		r.setState(FOLLOWER)
	}

	// voted already in the current term
	if r.lastVoteTerm >= r.getTerm() {
		return resp
	}

	r.lastVoteTerm = r.getTerm()
	resp.Granted = true
	return resp
}

func (r *Raft) startReplication() {
	replication := func(node string, done chan<- bool) {
		log.Infof("Replicating to node %s", node)
		replicationTicker := time.NewTicker(100 * time.Millisecond)
		repl := newReplicator(node)
		for {
			<-replicationTicker.C
			entries := AppendEntriesRequest{
				CurrentTerm: r.getTerm(),
				LeaderID:    r.GetLeader(),
			}
			repl.replicate(entries)
		}
	}

	for _, node := range r.clusterNodes {
		if node == ":"+strconv.Itoa(r.listenTCPPort) {
			continue
		}
		done := make(chan bool)
		go replication(node, done)
	}
}

func (r *Raft) runFollower() {
	log.Info("Node in FOLLOWER mode")
	heartBeatTimeout := randomTimeout(r.electionTimeOut)
	for r.getState() == FOLLOWER {
		select {
		case req := <-r.voteRequestCh:
			req.Response <- r.processVoteRequest(req)

		case <-r.votesReceivedCh:
			log.Panic("Not expecting a vote reponse at this state")

		case entries := <-r.appendEntriesCh:
			heartBeatTimeout = randomTimeout(r.electionTimeOut)
			r.setLeader(entries.LeaderID)
			r.lastEntry = time.Now().UnixNano()
			r.setTerm(entries.CurrentTerm)

		case <-heartBeatTimeout:
			log.Println("Heartbeat timeout")
			r.setState(CANDIDATE)
		}
	}
}

func (r *Raft) runLeader() {
	log.Info("Node in LEADER mode")
	r.startReplication()

	for r.getState() == LEADER {
		select {
		case req := <-r.voteRequestCh:
			req.Response <- r.processVoteRequest(req)

		case <-r.votesReceivedCh:
			log.Debug("Node retrie a vote but already elected as leader")

		}
	}
}

func (r *Raft) runCandidate() {
	var voteRequested = false
	var receivedVotes = 0
	log.Info("Node in CANDIDATE mode")

	electionTimer := time.After(1 * time.Second)
	for r.getState() == CANDIDATE {
		if !voteRequested {
			go r.requestVoteRequest()
			voteRequested = true
		}

		select {
		case req := <-r.voteRequestCh:
			req.Response <- r.processVoteRequest(req)

		case vote := <-r.votesReceivedCh:
			if vote.Granted {
				receivedVotes++
			}
			if receivedVotes >= r.quorum() {
				r.setState(LEADER)
				r.setLeader(":" + strconv.Itoa(r.listenTCPPort))
				log.Infof("Leader elected: %s", r.GetLeader())
			}

		case <-electionTimer:
			log.Debug("Election failed restarting the election")
			return
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
	raftNode := Raft{
		listenTCPPort:   listenPort,
		electionTimeOut: 10 * time.Second,
		lastEntry:       time.Now().UnixNano(),
		currentTerm:     1,
		clusterNodes:    clusterNodes,
		voteRequestCh:   make(chan RequestVoteRequest),
		votesReceivedCh: make(chan RequestVoteResponse),
		appendEntriesCh: make(chan AppendEntriesRequest),
		lastVoteTerm:    -1,
	}

	// start raft node as follower
	raftNode.setState(FOLLOWER)
	go raftNode.listen()
	go raftNode.runFSM()

	return &raftNode
}

// Shutdown terminate a node
func (r *Raft) Shutdown() {
	r.setState(SHUTDOWN)
}
