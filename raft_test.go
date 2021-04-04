package raft

import (
	"strconv"
	"testing"
	"time"

	"github.com/pegerto/raft/pkg/testutil"
	"github.com/stretchr/testify/assert"
)

func TestSingleNodeTransitionToLeader(t *testing.T) {
	nodePort := testutil.FreePort(t)
	cluster := []string{":" + strconv.Itoa(nodePort)}
	node := NewRaft(nodePort, cluster)
	isLeader := func() bool {
		return node.state == LEADER
	}
	assert.Eventually(t, isLeader, time.Minute*1, time.Second*1)
	assert.Equal(t, node.GetLeader(), ":"+strconv.Itoa(nodePort))
}
