package main

import (
	"context"
	"errors"
	"log"
	"net"
	commonenv "smapp/common/env"
	pb "smapp/common/grpc/image"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/status"
)

type imageServer struct {
	pb.UnimplementedImageServer
	client *s3.Client
}

func (s *imageServer) CheckObjectExists(ctx context.Context, req *pb.ObjectExistsRequest) (*pb.ObjectExistsResponse, error) {
	_, err := s.client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(req.Bucket),
		Key:    aws.String(req.Key),
	})

	var code codes.Code
	if err != nil {
		if errors.As(err, new(*types.NoSuchKey)) || errors.As(err, new(*types.NotFound)) {
			code = codes.NotFound
		} else {
			log.Println(err)
			code = codes.Unknown
		}
		return &pb.ObjectExistsResponse{}, status.Error(code, err.Error())
	}

	return &pb.ObjectExistsResponse{}, nil
}

func main() {
	defaultTimeout, err := commonenv.GetEnvDuration("DEFAULT_TIMEOUT")
	if err != nil {
		log.Fatal(err)
	}
	region, err := commonenv.GetEnv("S3_REGION")
	if err != nil {
		log.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(region))
	if err != nil {
		log.Fatal(err)
	}
	client := s3.NewFromConfig(cfg)

	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatal(err)
	}
	// Set maximum connection age to periodically trigger DNS lookups in case replicas were added/removed.
	grpcServer := grpc.NewServer(
		grpc.KeepaliveParams(keepalive.ServerParameters{
			MaxConnectionAge:      15 * time.Second,
			MaxConnectionAgeGrace: 5 * time.Second,
		}),
	)
	pb.RegisterImageServer(grpcServer, &imageServer{client: client})
	log.Fatal(grpcServer.Serve(lis))
}
