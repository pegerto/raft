package raft

import (
	"fmt"
	"net"
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

type Raft struct {
	state State
}

func (r *Raft) setState(state State) {
	r.state = state
}

func (r *Raft) getState() State {
	return r.state
}

func (r *Raft) listen() {
	handleConnection := func(conn net.Conn) {
		fmt.Println(conn)
	}

	ln, err := net.Listen("tcp", ":4040")
	if err != nil {
		// handle error
	}
	for {
		conn, err := ln.Accept()
		if err != nil {
			// handle error
			println(err)
		}
		go handleConnection(conn)
	}
}

// NewRaft creates a new Raft node
func NewRaft() *Raft {
	raftNode := Raft{}

	// start raft node as follower
	raftNode.setState(FOLLOWER)
	go raftNode.listen()

	return &raftNode
}
