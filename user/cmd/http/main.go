package main

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	"database/sql"
	"encoding/pem"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	commondb "smapp/common/db"
	commonenv "smapp/common/env"
	commonmw "smapp/common/middleware"
	"smapp/user/handlers"
	"smapp/user/repository"
	"smapp/user/service"

	"github.com/gorilla/mux"

	_ "github.com/go-sql-driver/mysql"
)

type mysqlConfig struct {
	host     string
	user     string
	password []byte
	db       string
}

type jwtConfig struct {
	privateKey *rsa.PrivateKey
	ttl        time.Duration
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

func getJWTConfig() (*jwtConfig, error) {
	jwtConfig := jwtConfig{}
	var err error
	privateKeyData, err := commonenv.GetSecret("jwt_private_key")
	if err != nil {
		return nil, err
	}
	block, _ := pem.Decode(privateKeyData)
	if block == nil || block.Type != "PRIVATE KEY" {
		return nil, errors.New("failed to decode PEM block containing private key")
	}
	privateKey, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	var ok bool
	if jwtConfig.privateKey, ok = privateKey.(*rsa.PrivateKey); !ok {
		return nil, errors.New("private key is not RSA")
	}

	if jwtConfig.ttl, err = commonenv.GetEnvDuration("JWT_TTL"); err != nil {
		return nil, err
	}
	return &jwtConfig, nil
}

func main() {
	mysqlConfig, err := getMysqlConfig()
	if err != nil {
		log.Fatal(err)
	}
	jwtConfig, err := getJWTConfig()
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

	userRepository := repository.NewUser(db)
	followRepository := repository.NewFollow(db)

	jwtService := service.NewJWT(jwtConfig.privateKey, jwtConfig.ttl)
	userService := service.NewUser(userRepository, jwtService)
	followService := service.NewFollow(followRepository)

	r := mux.NewRouter()
	r.Handle("/signup", handlers.Signup(userService)).Methods(http.MethodPost)
	r.Handle("/login", handlers.Login(userService)).Methods(http.MethodPost)
	r.Handle(
		"/users/{user_id}/follow",
		commonmw.ParseUserID(handlers.Follow(followService)),
	).Methods(http.MethodPost)
	r.Use(commonmw.WithRequestContextTimeout(defaultTimeout))

	srv := &http.Server{
		Addr:        ":8080",
		Handler:     r,
		ReadTimeout: defaultTimeout,
	}
	log.Fatal(srv.ListenAndServe())
}
