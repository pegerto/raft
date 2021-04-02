package raft

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStartAsFollower(t *testing.T) {
	node := NewRaft()
	assert.Equal(t, node.getState(), FOLLOWER)
}
