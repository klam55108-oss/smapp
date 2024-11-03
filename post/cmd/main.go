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
	commonhttp "smapp/common/http"
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

	conn, err := grpc.NewClient("image_grpc:50055", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
	imageClient := imagePB.NewImageClient(conn)
	postRepository := repository.NewPost(db)
	postService := service.NewPost(postRepository, imageClient)
	postHandler := handlers.CreatePost(postService)

	commentRepository := repository.NewComment(db)
	commentService := service.NewComment(commentRepository)
	commentHandler := handlers.CreateComment(commentService)

	r := mux.NewRouter()
	r.Handle("/posts", postHandler).Methods(http.MethodPost)
	r.Handle("/posts/{post_id}/comments", commentHandler).Methods(http.MethodPost)
	r.Use(commonhttp.WithRequestContextTimeout(defaultTimeout))
	srv := &http.Server{
		Addr:        ":8082",
		Handler:     r,
		ReadTimeout: defaultTimeout,
	}
	log.Fatal(srv.ListenAndServe())
}
