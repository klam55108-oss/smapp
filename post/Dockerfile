FROM golang:1.22-alpine3.20 AS builder
WORKDIR /app/post
COPY post/go.mod post/go.sum ./
# common/go.sum may not exist
COPY common/go.mod common/go.su[m] ../common/
RUN go mod download
COPY common/. ../common
COPY post/. .
ARG TARGETOS
ARG TARGETARCH
RUN GOOS=$TARGETOS GOARCH=$TARGETARCH go build -o main cmd/main.go

FROM alpine:3.20
WORKDIR /app/post
COPY --from=builder /app/post/main .
EXPOSE 8080
CMD [ "./main"]