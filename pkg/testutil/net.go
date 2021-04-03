package testutil

import (
	"net"
	"testing"
)

func FreePort(t *testing.T) int {
	addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
	if err != nil {
		t.Fatal("Error finding a free port", err)
	}

	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		t.Fatal("Error finding a free port", err)
	}
	defer l.Close()
	return l.Addr().(*net.TCPAddr).Port
}
