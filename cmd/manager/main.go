package main

import (
	"log"

	"github.com/erkrnt/symphony/internal/manager"
	"github.com/erkrnt/symphony/internal/pkg/config"
	"github.com/sirupsen/logrus"
)

func main() {
	flags, err := config.GetFlags()

	if err != nil {
		logrus.Fatal(err)
	}

	node, err := manager.NewNode(flags)

	if err != nil {
		log.Fatal(err)
	}

	_, jtsErr := node.State.FindOrCreateJoinTokens()

	if jtsErr != nil {
		log.Fatal(err)
	}

	manager.Start(flags, node)
}
