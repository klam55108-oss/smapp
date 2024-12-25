package image

//go:generate protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative image.proto

//go:generate mockgen -destination mocks/image.go -package mocks . ImageClient
