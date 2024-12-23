package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	commondb "smapp/common/db"
	commonenv "smapp/common/env"
	imagePB "smapp/common/grpc/image"
	userPB "smapp/common/grpc/user"
	commonmw "smapp/common/middleware"
	"smapp/common/validation"
	"smapp/post/handlers"
	"smapp/post/repository"
	"smapp/post/service"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
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

func main() {
	mysqlConfig, err := getMysqlConfig()
	if err != nil {
		log.Fatal(err)
	}
	defaultTimeout, err := commonenv.GetEnvDuration("DEFAULT_TIMEOUT")
	if err != nil {
		log.Fatal(err)
	}

	db, err := sql.Open(
		"mysql",
		fmt.Sprintf(
			"%s:%s@tcp(%s)/%s?parseTime=true",
			mysqlConfig.user, mysqlConfig.password, mysqlConfig.host, mysqlConfig.db,
		),
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

	conn, err := grpc.NewClient(
		"user_grpc:50051",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultServiceConfig(`{"loadBalancingConfig": [{"round_robin":{}}]}`),
	)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
	userClient := userPB.NewUserClient(conn)

	conn, err = grpc.NewClient(
		"image_grpc:50051",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultServiceConfig(`{"loadBalancingConfig": [{"round_robin":{}}]}`),
	)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
	imageClient := imagePB.NewImageClient(conn)

	postRepository := repository.NewDefaultPost(db)
	commentRepository := repository.NewComment(db)
	postLikeRepository := repository.NewPostLike(db)
	commentLikeRepository := repository.NewCommentLike(db)

	postService := service.NewDefaultPost(postRepository, commentRepository, postLikeRepository, userClient, imageClient)
	commentService := service.NewComment(commentRepository, postRepository)
	postLikeService := service.NewPostLike(postLikeRepository, postRepository)
	commentLikeService := service.NewCommentLike(commentLikeRepository, commentRepository)

	r := mux.NewRouter()
	r.Handle(
		"/posts",
		commonmw.ParseUserID(handlers.CreatePost(validation.DefaultValidator, postService)),
	).Methods(http.MethodPost)
	r.Handle(
		"/posts/{post_id}",
		handlers.GetPost(postService),
	).Methods(http.MethodGet)
	r.Handle(
		"/posts/{post_id}/comments",
		commonmw.ParseUserID(handlers.CreateComment(commentService)),
	).Methods(http.MethodPost)
	r.Handle(
		"/posts/{post_id}/comments",
		handlers.GetComments(commentService),
	).Methods(http.MethodGet)
	r.Handle(
		"/posts/{entity_id}/likes",
		commonmw.ParseUserID(handlers.CreateLike(postLikeService)),
	).Methods(http.MethodPost)
	r.Handle(
		"/comments/{entity_id}/likes",
		commonmw.ParseUserID(handlers.CreateLike(commentLikeService)),
	).Methods(http.MethodPost)
	r.Handle(
		"/feed",
		commonmw.ParseUserID(handlers.GetFeed(postService)),
	).Methods(http.MethodGet)

	r.Use(commonmw.WithRequestContextTimeout(defaultTimeout))

	srv := &http.Server{
		Addr:        ":8080",
		Handler:     r,
		ReadTimeout: defaultTimeout,
	}
	log.Fatal(srv.ListenAndServe())
}
