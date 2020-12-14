package main

import (
	"github.com/erkrnt/symphony/internal/block"
	"github.com/sirupsen/logrus"
)

func main() {
	s, err := block.New()

	if err != nil {
		logrus.Fatal(err)
	}

	go s.Start()

	startErr := <-s.ErrorC

	logrus.Fatal(startErr)
}
