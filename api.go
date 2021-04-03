package raft

type RequestType int32

const (
	VoteRequest RequestType = iota
	VoteResponse
)

type RPCHeader struct {
	Version int16
	Type    RequestType
}

type RawRequest struct {
	RPCHeader
}

// RequestVoteRequest requests a
type RequestVoteRequest struct {
	RPCHeader
	Term      uint64
	Candidate string
	Response  chan RequestVoteResponse
}

// RequestVoteResponse
type RequestVoteResponse struct {
	RPCHeader
	Granted bool
}
