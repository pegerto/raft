package raft

// RequestVoteRequest requests a
type RequestVoteRequest struct {
	Term      int64
	Candidate string
	Response  chan RequestVoteResponse
}

// RequestVoteResponse does
type RequestVoteResponse struct {
	Granted bool
}

// AppendEntriesRequest request
type AppendEntriesRequest struct {
	currentTerm int64
}

// AppendEntriesResponse response
type AppendEntriesResponse struct {
	term    int64
	success bool
}
