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

	go manager.RemoteServer(m)

	manager.ControlServer(m)
}
