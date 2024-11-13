package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net"
	commondb "smapp/common/db"
	commonenv "smapp/common/env"
	pb "smapp/common/grpc/user"
	"smapp/user/repository"
	"smapp/user/service"
	"time"

	"github.com/google/uuid"
	"google.golang.org/grpc"
)

type mysqlConfig struct {
	host     string
	user     string
	password []byte
	db       string
}

func getMysqlConfig() (*mysqlConfig, error) {
	mysqlConfig := mysqlConfig{}
	var err error
	if mysqlConfig.host, err = commonenv.GetEnv("MYSQL_HOST"); err != nil {
		return nil, err
	}
	if mysqlConfig.user, err = commonenv.GetEnv("MYSQL_USER"); err != nil {
		return nil, err
	}
	if mysqlConfig.password, err = commonenv.GetSecret("mysql_password"); err != nil {
		return nil, err
	}
	if mysqlConfig.db, err = commonenv.GetEnv("MYSQL_DB"); err != nil {
		return nil, err
	}
	return &mysqlConfig, nil
}

type userServer struct {
	pb.UnimplementedUserServer
	followService *service.Follow
}

func (s *userServer) GetFollowed(ctx context.Context, req *pb.GetFollowedRequest) (*pb.GetFollowedResponse, error) {
	userID, err := uuid.FromBytes(req.UserId)
	if err != nil {
		return nil, err
	}
	followed, err := s.followService.GetFollowed(ctx, userID)
	if err != nil {
		return nil, err
	}
	followedBytes := make([][]byte, len(followed))
	for i, id := range followed {
		followedBytes[i] = id[:]
	}
	return &pb.GetFollowedResponse{UserIds: followedBytes}, nil
}

func main() {
	mysqlConfig, err := getMysqlConfig()
	if err != nil {
		log.Fatal(err)
	}

	db, err := sql.Open(
		"mysql",
		fmt.Sprintf("%s:%s@tcp(%s)/%s", mysqlConfig.user, mysqlConfig.password, mysqlConfig.host, mysqlConfig.db),
	)
	if err != nil {
		log.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	err = commondb.WaitForDB(ctx, db)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	followService := service.NewFollow(repository.NewFollow(db))

	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatal(err)
	}
	s := grpc.NewServer()
	pb.RegisterUserServer(s, &userServer{followService: followService})
	log.Fatal(s.Serve(lis))
}
