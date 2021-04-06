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
	// SHUTDOWN state
	SHUTDOWN
)

// Raft node
type Raft struct {
	state               State
	listenAddress       string
	listenTCPPort       int
	electionTimeOut     time.Duration
	lastEntry           int64
	lastVoteTerm        uint64
	currentTerm         uint64
	clusterNodes        []string
	leader              string
	voteRequestCh       chan RequestVoteRequest
	votesReceivedCh     chan RequestVoteResponse
	appendEntriesCh     chan AppendEntriesRequest
	stopReplicationTask []chan bool
}

func (r *Raft) getNodeID() string {
	return r.listenAddress + ":" + strconv.Itoa(r.listenTCPPort)
}

func (r *Raft) setState(state State) {
	r.state = state
}

func (r *Raft) getState() State {
	return r.state
}

func (r *Raft) setTerm(term uint64) {
	r.currentTerm = term
}

func (r *Raft) getTerm() uint64 {
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
	// TODO: safty requires to persist the node before, grant a vote, to avoid
	// a node that has restarted to vote twice under the same term.
	if r.lastVoteTerm >= r.getTerm() {
		return resp
	}

	r.lastVoteTerm = r.getTerm()

	resp.Granted = true
	return resp
}

func (r *Raft) startReplication() {
	replication := func(node string, kill <-chan bool) {
		log.Infof("Replicating to node %s", node)
		replicationTicker := time.NewTicker(100 * time.Millisecond)
		repl := newReplicator(node)
		entries := AppendEntriesRequest{
			CurrentTerm: r.getTerm(),
			LeaderID:    r.GetLeader(),
		}
		for {
			select {
			case <-replicationTicker.C:
				repl.replicate(entries)
			case <-kill:
				log.Println("Stop replication")
				return
			}
		}
	}

	i := 0
	r.stopReplicationTask = make([]chan bool, len(r.clusterNodes)-1)
	for _, node := range r.clusterNodes {
		if node == r.getNodeID() {
			continue
		}
		r.stopReplicationTask[i] = make(chan bool)
		go replication(node, r.stopReplicationTask[i])
		i++
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
			heartBeatTimeout = randomTimeout(r.electionTimeOut)
			r.setState(CANDIDATE)
		}
	}
}

func (r *Raft) runLeader() {
	log.Info("Node in LEADER mode")
	r.startReplication()
	// terminate replication if the node does hold leadership
	defer func() {
		for _, task := range r.stopReplicationTask {
			task <- true
		}
	}()

	leaderTicker := time.NewTicker(100 * time.Millisecond)

	for r.getState() == LEADER {
		select {
		case req := <-r.voteRequestCh:
			req.Response <- r.processVoteRequest(req)

		case <-r.votesReceivedCh:
			log.Debug("Node retreived a vote but already elected as leader")

		case <-leaderTicker.C:
			continue
		}
	}

}

func (r *Raft) runCandidate() {
	var voteRequested = false
	var receivedVotes = 0
	log.Info("Node in CANDIDATE mode")
	electionTimer := time.After(1 * time.Second)

	//increase term for a new election
	r.setTerm(r.getTerm() + 1)

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
				r.setLeader(r.getNodeID())
				log.Infof("Leader elected: %s", r.GetLeader())
			}

		case <-electionTimer:
			log.Debug("Election failed restarting the election")
			return
		}

	}
}

func (r *Raft) runShutdown() {

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
		case SHUTDOWN:
			r.runShutdown()
			goto DONE
		}
	}
DONE:
	log.Println("Raft FSM stopped")
}

// NewRaft creates a new Raft node
func NewRaft(listenAddress string, listenPort int, clusterNodes []string) *Raft {
	raftNode := Raft{
		listenTCPPort:   listenPort,
		listenAddress:   listenAddress,
		electionTimeOut: 10 * time.Second,
		lastEntry:       time.Now().UnixNano(),
		currentTerm:     1,
		clusterNodes:    clusterNodes,
		voteRequestCh:   make(chan RequestVoteRequest),
		votesReceivedCh: make(chan RequestVoteResponse),
		appendEntriesCh: make(chan AppendEntriesRequest),
		lastVoteTerm:    0,
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
