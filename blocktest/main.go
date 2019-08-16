/*
 *
 * Copyright 2015 gRPC authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

// Package main implements a client for Greeter service.
package main

import (
	"context"
	"log"
	"time"

	pb "gitlab.drn.io/erikreinert/symphony/proto"

	"google.golang.org/grpc"
)

const (
	address = "192.168.159.193:50051"
)

func main() {
	// Set up a connection to the server.
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := pb.NewBlockClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	// CreatePv
	createPv, err := c.CreatePv(ctx, &pb.CreatePvRequest{Device: "/dev/sdb"})
	if err != nil {
		log.Fatalf("could not greet: %v", err)
	}
	log.Printf("CreatePv: %s", createPv.PvName)

	// CreateVg
	createVg, err := c.CreateVg(ctx, &pb.CreateVgRequest{Device: "/dev/sdb", Group: "virtual"})
	if err != nil {
		log.Fatalf("could not greet: %v", err)
	}
	log.Printf("CreateVg: %s", createVg.VgName)

	// CreateLv
	createLv, err := c.CreateLv(ctx, &pb.CreateLvRequest{Group: "virtual", Name: "test", Size: "10g"})
	if err != nil {
		log.Fatalf("could not greet: %v", err)
	}
	log.Printf("CreateLv: %s", createLv.LvName)

	// GetPv
	getPv, err := c.GetPv(ctx, &pb.GetPvRequest{Device: "/dev/sdb"})
	if err != nil {
		log.Fatalf("could not greet: %v", err)
	}
	log.Printf("GetPv: %s", getPv.PvName)

	// GetVg
	getVg, err := c.GetVg(ctx, &pb.GetVgRequest{Group: "virtual"})
	if err != nil {
		log.Fatalf("could not greet: %v", err)
	}
	log.Printf("GetVg: %s", getVg.VgName)

	// GetLv
	getLv, err := c.GetLv(ctx, &pb.GetLvRequest{Group: "virtual", Name: "test"})
	if err != nil {
		log.Fatalf("could not greet: %v", err)
	}
	log.Printf("GetLv: %s", getLv.LvName)

	// RemoveLv
	removeLv, err := c.RemoveLv(ctx, &pb.RemoveLvRequest{Group: "virtual", Name: "test"})
	if err != nil {
		log.Fatalf("could not greet: %v", err)
	}
	log.Printf("RemoveLv: %s", removeLv.Message)

	// RemoveVg
	removeVg, err := c.RemoveVg(ctx, &pb.RemoveVgRequest{Group: "virtual"})
	if err != nil {
		log.Fatalf("could not greet: %v", err)
	}
	log.Printf("RemoveVg: %s", removeVg.Message)

	// RemovePv
	removePv, err := c.RemovePv(ctx, &pb.RemovePvRequest{Device: "/dev/sdb"})
	if err != nil {
		log.Fatalf("could not greet: %v", err)
	}
	log.Printf("RemovePv: %s", removePv.Message)
}
