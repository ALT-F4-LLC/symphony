package main

import (
	"log"

	"github.com/erkrnt/symphony/internal/block"
	"github.com/erkrnt/symphony/internal/pkg/config"
	"github.com/sirupsen/logrus"
)

func main() {
	flags, err := config.GetFlags()

	if err != nil {
		logrus.Fatal(err)
	}

	node, nodeErr := block.NewNode(flags)

	if nodeErr != nil {
		log.Fatal(nodeErr)
	}

	block.Start(flags, node)
}
