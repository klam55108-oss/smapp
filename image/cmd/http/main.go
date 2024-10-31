package main

import (
	"context"
	"log"
	"net/http"

	commonenv "smapp/common/env"
	"smapp/image/handlers"
	"smapp/image/service"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/gorilla/mux"
)

func main() {
	defaultTimeout, err := commonenv.GetEnvDuration("DEFAULT_TIMEOUT")
	if err != nil {
		log.Fatal(err)
	}
	profileImgLimit, err := commonenv.GetEnvInt64("PROFILE_IMG_LIMIT")
	if err != nil {
		log.Fatal(err)
	}
	postImgLimit, err := commonenv.GetEnvInt64("POST_IMG_LIMIT")
	if err != nil {
		log.Fatal(err)
	}
	policyTTL, err := commonenv.GetEnvDuration("POLICY_TTL")
	if err != nil {
		log.Fatal(err)
	}
	bucket, err := commonenv.GetEnv("S3_BUCKET")
	if err != nil {
		log.Fatal(err)
	}
	region, err := commonenv.GetEnv("S3_REGION")
	if err != nil {
		log.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		log.Fatal(err)
	}
	creds, err := cfg.Credentials.Retrieve(ctx)
	if err != nil {
		log.Fatal(err)
	}
	if creds.CanExpire {
		log.Fatal("credentials must not expire")
	}

	generateUploadFormService := service.NewGenerateUploadForm(&creds, policyTTL, bucket, region)

	r := mux.NewRouter()
	r.Handle(
		"/upload-form/profile",
		handlers.GenerateUploadForm(generateUploadFormService, profileImgLimit),
	).Methods(http.MethodPost)
	r.Handle(
		"/upload-form/post",
		handlers.GenerateUploadForm(generateUploadFormService, postImgLimit),
	).Methods(http.MethodPost)
	srv := &http.Server{
		Addr:        ":8085",
		Handler:     r,
		ReadTimeout: defaultTimeout,
	}
	log.Fatal(srv.ListenAndServe())
}
