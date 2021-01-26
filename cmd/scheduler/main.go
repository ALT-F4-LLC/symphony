package main

import (
	"github.com/erkrnt/symphony/internal/scheduler"
	"github.com/sirupsen/logrus"
)

func main() {
	s, err := scheduler.New()

	if err != nil {
		logrus.Fatal(err)
	}

	go s.Start()

	startErr := <-s.Service.ErrorC

	logrus.Fatal(startErr)
}
