package raft

import (
	"math/rand"
	"time"
)

func randomTimeout(minVal time.Duration) <-chan time.Time {
	epsilon := (time.Duration(rand.Int63()) % minVal)
	return time.After(minVal + epsilon)
}
