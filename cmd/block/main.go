package main

import (
	"log"

	"github.com/erkrnt/symphony/internal/block"
	"github.com/erkrnt/symphony/internal/pkg/cluster"
)

func main() {
	node, err := cluster.NewNode()

	if err != nil {
		log.Fatal(err)
	}

	go block.StartRemoteServer(node)

	block.StartControlServer(node)
}
