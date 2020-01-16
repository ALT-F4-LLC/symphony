package main

import (
	"log"

	"github.com/erkrnt/symphony/internal/block"
	"github.com/erkrnt/symphony/internal/pkg/config"
	"github.com/sirupsen/logrus"
)

func main() {
	f, err := config.GetFlags()

	if err != nil {
		logrus.Fatal(err)
	}

	_, nodeErr := block.NewNode(f)

	if nodeErr != nil {
		log.Fatal(err)
	}

	// block.Start(f, m)
}
