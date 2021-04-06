package raft

// RequestVoteRequest requests a
type RequestVoteRequest struct {
	Term      uint64
	Candidate string
	Response  chan RequestVoteResponse
}

// RequestVoteResponse does
type RequestVoteResponse struct {
	Granted bool
}

// AppendEntriesRequest request
type AppendEntriesRequest struct {
	CurrentTerm uint64
	LeaderID    string
}

// AppendEntriesResponse response
type AppendEntriesResponse struct {
	Term    uint64
	Success bool
}
