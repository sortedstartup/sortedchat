package main

import (
	"log"
	"net"

	"google.golang.org/grpc"
	"sortedstartup.com/chatservice/api"

	pb "sortedstartup.com/chatservice/proto"
)

func main() {
	lis, err := net.Listen("tcp", ":8000")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterSortedChatServer(grpcServer, &api.Server{})

	log.Println("gRPC server listening on :8000")
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
