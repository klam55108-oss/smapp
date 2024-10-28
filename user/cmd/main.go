package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"time"

	commondb "smapp/common/db"
	commonenv "smapp/common/env"
	commonhttp "smapp/common/http"
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
	secret         []byte
	expirationTime time.Duration
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
	if jwtConfig.secret, err = commonenv.GetSecret("jwt_secret"); err != nil {
		return nil, err
	}
	if jwtConfig.expirationTime, err = commonenv.GetEnvDuration("JWT_EXPIRATION_TIME"); err != nil {
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
	jwtService := service.NewJWT(jwtConfig.secret, jwtConfig.expirationTime)
	authService := service.NewAuth(userRepository, jwtService)

	r := mux.NewRouter()
	r.Handle("/signup", handlers.Signup(authService)).Methods(http.MethodPost)
	r.Handle("/login", handlers.Login(authService)).Methods(http.MethodPost)
	r.Use(commonhttp.WithRequestContextTimeout(defaultTimeout))
	srv := &http.Server{
		Addr:        ":8081",
		Handler:     r,
		ReadTimeout: defaultTimeout,
	}
	log.Fatal(srv.ListenAndServe())
}
