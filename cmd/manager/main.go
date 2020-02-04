package main

import (
	"log"

	"github.com/erkrnt/symphony/internal/manager"
)

func main() {
	m, err := manager.NewNode()

	if err != nil {
		log.Fatal(err)
	}

	go m.StartRemoteServer()

	m.StartControlServer()
}
