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
