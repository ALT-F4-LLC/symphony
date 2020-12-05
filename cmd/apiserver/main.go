package main

import (
	"github.com/erkrnt/symphony/internal/apiserver"
	"github.com/sirupsen/logrus"
)

func main() {
	s, err := apiserver.New()

	if err != nil {
		logrus.Fatal(err)
	}

	go s.Start()

	errorC := <-s.ErrorC

	logrus.Fatal(errorC)
}
