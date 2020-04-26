package main

import (
	"log"

	"github.com/erkrnt/symphony/internal/block"
)

func main() {
	b, err := block.New()

	if err != nil {
		log.Fatal(err)
	}

	go block.RemoteServer(b)

	block.ControlServer(b)
}
