package raft

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"net"
	"strconv"
)

func (r *Raft) requestVoteRequest() {
	r.setTerm(r.getTerm() + 1)
	header := RPCHeader{
		1,
		VoteRequest,
	}
	requestVoteRequest := RequestVoteRequest{
		RPCHeader: header,
		Candidate: ":" + strconv.Itoa(r.listenTCPPort),
	}

	sendRequestVote := func(node string) {
		conn, err := net.Dial("tcp", node)
		if err != nil {
			fmt.Printf("Error open a connection to node: %s \n", err)
			return
		}
		defer conn.Close()
		enc := gob.NewEncoder(conn)
		enc.Encode(requestVoteRequest)

		var resp = RequestVoteResponse{}
		dec := gob.NewDecoder(conn)
		dec.Decode(&resp)
		fmt.Println(resp)
	}

	for _, node := range r.clusterNodes {
		fmt.Printf("Request vote to candidate: %s \n", node)
		go sendRequestVote(node)
	}

}

func (r *Raft) listen() {

	handleConnection := func(conn net.Conn) {
		fmt.Println("new connection")
		defer conn.Close()
		enc := gob.NewEncoder(conn)
		request := RawRequest{}

		//connBytes, err := ioutil.ReadAll(conn)
		// can contine here as connection is not close.

		// if err != nil {
		// 	fmt.Printf("Error reading request: %s\n", err)
		// }

		dec := gob.NewDecoder(bytes.NewReader(connBytes))
		err = dec.Decode(&request)
		if err != nil {
			fmt.Printf("Error decoding the request: %s\n", err)
		}

		switch request.RPCHeader.Type {
		case VoteRequest:
			fmt.Println("New vote request")
			voteRequest := RequestVoteRequest{}
			dec = gob.NewDecoder(bytes.NewReader(connBytes))
			dec.Decode(&voteRequest)
			voteRequest.Response = make(chan RequestVoteResponse)

			r.voteRequestCh <- voteRequest
			enc.Encode(<-voteRequest.Response)
		}
	}

	ln, err := net.Listen("tcp", ":"+strconv.Itoa(r.listenTCPPort))
	if err != nil {
		// handle error
		fmt.Println(err)
	}
	for {
		conn, err := ln.Accept()
		if err != nil {
			// handle error
			fmt.Println(err)
		}
		go handleConnection(conn)
	}
}
