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
	CurrentTerm int64
	LeaderId    string
}

// AppendEntriesResponse response
type AppendEntriesResponse struct {
	Term    int64
	Success bool
}
