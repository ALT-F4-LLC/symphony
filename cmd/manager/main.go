package main

import (
	"log"

	"github.com/erkrnt/symphony/internal/manager"
)

func main() {
	m, err := manager.New()

	if err != nil {
		log.Fatal(err)
	}

	go manager.StartRemoteServer(m)

	manager.StartControlServer(m)
}
