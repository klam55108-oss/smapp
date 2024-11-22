FROM golang:1.22-alpine3.20 AS builder
WORKDIR /app/user
COPY user/go.mod user/go.sum ./
# common/go.sum may not exist
COPY common/go.mod common/go.su[m] ../common/
RUN go mod download
COPY common/. ../common
COPY user/. .
ARG TARGETOS
ARG TARGETARCH
RUN GOOS=$TARGETOS GOARCH=$TARGETARCH go build -o main cmd/grpc/main.go

FROM alpine:3.20
WORKDIR /app/user
COPY --from=builder /app/user/main .
EXPOSE 50051
CMD [ "./main"]