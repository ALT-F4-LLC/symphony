package main

import (
	"github.com/erkrnt/symphony/internal/manager"
	"github.com/sirupsen/logrus"
)

func main() {
	m, err := manager.New()

	if err != nil {
		logrus.Fatal(err)
	}

	m.Start()
}
