package main

import (
	"log"

	"github.com/erkrnt/symphony/internal/manager"
	"github.com/erkrnt/symphony/internal/pkg/cluster"
)

func main() {
	node, err := cluster.NewNode()

	if err != nil {
		log.Fatal(err)
	}

	go manager.StartRemoteServer(node)

	manager.StartControlServer(node)
}
