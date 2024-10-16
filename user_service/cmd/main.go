package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"time"

	"smapp/common"
	"smapp/user_service/handlers"
	"smapp/user_service/repository"
	"smapp/user_service/service"
)

type mysqlConfig struct {
	host     string
	user     string
	password []byte
	db       string
}

type queryTimeoutConfig struct {
	insert time.Duration
	get    time.Duration
}

type jwtConfig struct {
	secret         []byte
	expirationTime time.Duration
}

func getMysqlConfig() (*mysqlConfig, error) {
	mysqlConfig := mysqlConfig{}
	var err error
	if mysqlConfig.host, err = common.GetEnv("MYSQL_HOST"); err != nil {
		return nil, err
	}
	if mysqlConfig.user, err = common.GetEnv("MYSQL_USER"); err != nil {
		return nil, err
	}
	if mysqlConfig.password, err = common.GetSecret("mysql_password"); err != nil {
		return nil, err
	}
	if mysqlConfig.db, err = common.GetEnv("MYSQL_DB"); err != nil {
		return nil, err
	}
	return &mysqlConfig, nil
}

func getQueryTimeoutConfig() (*queryTimeoutConfig, error) {
	queryTimeoutConfig := queryTimeoutConfig{}
	var err error
	if queryTimeoutConfig.insert, err = common.GetEnvDuration("INSERT_QUERY_TIMEOUT"); err != nil {
		return nil, err
	}
	if queryTimeoutConfig.get, err = common.GetEnvDuration("GET_QUERY_TIMEOUT"); err != nil {
		return nil, err
	}
	return &queryTimeoutConfig, nil
}

func getJWTConfig() (*jwtConfig, error) {
	jwtConfig := jwtConfig{}
	var err error
	if jwtConfig.secret, err = common.GetSecret("jwt_secret"); err != nil {
		return nil, err
	}
	if jwtConfig.expirationTime, err = common.GetEnvDuration("JWT_EXPIRATION_TIME"); err != nil {
		return nil, err
	}
	return &jwtConfig, nil
}

func main() {
	mysqlConfig, err := getMysqlConfig()
	if err != nil {
		log.Fatal(err)
	}
	queryTimeoutConfig, err := getQueryTimeoutConfig()
	if err != nil {
		log.Fatal(err)
	}
	jwtConfig, err := getJWTConfig()
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
	err = common.WaitForDB(ctx, db)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	userRepository := repository.NewUser(db, queryTimeoutConfig.insert, queryTimeoutConfig.get)
	jwtService := service.NewJWT(jwtConfig.secret, jwtConfig.expirationTime)
	authService := service.NewAuth(userRepository, jwtService)

	http.Handle("/signup", handlers.Signup(authService))
	http.Handle("/login", handlers.Login(authService))
	log.Fatal(http.ListenAndServe(":8081", nil))
}
