package main

import (
	"net"

	"github.com/jinzhu/gorm"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"

	"github.com/erkrnt/symphony/protobuff"
)

const (
	port = ":36837"
)

type managerServer struct {
	db *gorm.DB
}

func main() {
	flags := getFlags()

	if flags.Verbose {
		logrus.SetLevel(logrus.DebugLevel)
	}

	db, err := getDatabase(flags)

	if err != nil {
		panic(err)
	}

	defer db.Close()

	lis, err := net.Listen("tcp", port)

	if err != nil {
		logrus.Fatal("Failed to listen: %v", err)
	}

	s := grpc.NewServer()

	protobuff.RegisterManagerServer(s, &managerServer{db: db})

	logrus.Info("Started manager service.")

	if err := s.Serve(lis); err != nil {
		logrus.Fatal("Failed to serve: %v", err)
	}
}
