package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/pegerto/raft"
	"github.com/pkg/profile"
	"github.com/urfave/cli"
)

func main() {
	var port int
	var cluster string
	var address string

	defer profile.Start().Stop()
	app := cli.NewApp()
	app.Name = "raft"
	app.Usage = "run a raft node"
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:        "ip",
			Usage:       "listen api address",
			Value:       "127.0.0.1",
			Destination: &address,
		},
		cli.IntFlag{
			Name:        "port",
			Value:       4001,
			Usage:       "listening port",
			Destination: &port,
		},
		cli.StringFlag{
			Name:        "cluster",
			Usage:       "coma separate nodes",
			Destination: &cluster,
		},
	}

	app.Action = func(c *cli.Context) error {
		nodes := strings.Split(cluster, ",")
		raftNode := raft.NewRaft(address, port, nodes)
		for {
			fmt.Printf("Leader: %s\n", raftNode.GetLeader())
			time.Sleep(10 * time.Second)
		}
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
