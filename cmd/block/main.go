package main

import (
	"github.com/erkrnt/symphony/internal/block"
	"github.com/sirupsen/logrus"
)

func main() {
	b, err := block.New()

	if err != nil {
		logrus.Fatal(err)
	}

	go block.RemoteServer(b)

	block.ControlServer(b)
}
