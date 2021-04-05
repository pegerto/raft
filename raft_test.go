package raft

import (
	"strconv"
	"testing"
	"time"

	"github.com/pegerto/raft/pkg/testutil"
	"github.com/stretchr/testify/assert"
)

func oneLeaderInCluster(nodes []*Raft) func() bool {
	return func() bool {

		var leaderCount = 0

		for _, node := range nodes {
			if node.state == LEADER {
				leaderCount++
			}
		}
		return leaderCount == 1
	}
}

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

func TestLeaderElectionOnCluster(t *testing.T) {
	node1Port, node2Port, node3Port := testutil.ClusterPorts(t)
	cluster := []string{":" + strconv.Itoa(node1Port), ":" + strconv.Itoa(node2Port), ":" + strconv.Itoa(node3Port)}
	node1 := NewRaft(node1Port, cluster)
	node2 := NewRaft(node2Port, cluster)
	node3 := NewRaft(node3Port, cluster)
	nodes := []*Raft{node1, node2, node3}

	allNodesKnownTheLeader := func() bool {
		leader := nodes[0].GetLeader()
		for _, node := range nodes {
			if node.GetLeader() != leader {
				return false
			}
		}
		return leader != ""
	}

	assert.Eventually(t, oneLeaderInCluster(nodes), time.Minute*1, time.Second*1)
	assert.Eventually(t, allNodesKnownTheLeader, time.Minute*1, time.Second*1)
}

func TestShutdownOfLeaderNode(t *testing.T) {

	node1Port, node2Port, node3Port := testutil.ClusterPorts(t)
	cluster := []string{":" + strconv.Itoa(node1Port), ":" + strconv.Itoa(node2Port), ":" + strconv.Itoa(node3Port)}
	node1 := NewRaft(node1Port, cluster)
	node2 := NewRaft(node2Port, cluster)
	node3 := NewRaft(node3Port, cluster)
	nodes := []*Raft{node1, node2, node3}

	assert.Eventually(t, oneLeaderInCluster(nodes), time.Minute*1, time.Second*1)
	// Set node in shutdown
	for _, node := range nodes {
		if node.getState() == LEADER {
			node.Shutdown()
		}
	}
	assert.Eventually(t, oneLeaderInCluster(nodes), time.Minute*1, time.Second*1)
}
