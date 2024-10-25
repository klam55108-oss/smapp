package main

import (
	"log"
	"net"
	pb "smapp/common/grpc/image"

	"google.golang.org/grpc"
)

type imageServer struct {
	pb.UnimplementedImageServer
}

func main() {
	lis, err := net.Listen("tcp", ":50055")
	if err != nil {
		log.Fatal(err)
	}
	grpcServer := grpc.NewServer()
	pb.RegisterImageServer(grpcServer, &imageServer{})
	log.Fatal(grpcServer.Serve(lis))
}
