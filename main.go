package main

import (
	"log"
	"personal_schedule_service/cmd"
)

func main() {
	log.Println("gRPC servers are running...\n")
	cmd.RunGRPCServer()
	// run the grpc server
}
