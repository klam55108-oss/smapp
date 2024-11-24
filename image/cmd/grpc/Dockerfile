FROM golang:1.22-alpine3.20 AS builder
WORKDIR /app/image
COPY image/go.mod image/go.sum ./
# common/go.sum may not exist
COPY common/go.mod common/go.su[m] ../common/
RUN go mod download
COPY common/. ../common
COPY image/. .
ARG TARGETOS
ARG TARGETARCH
RUN GOOS=$TARGETOS GOARCH=$TARGETARCH go build -o main cmd/grpc/main.go

FROM alpine:3.20
WORKDIR /app/image
COPY --from=builder /app/image/main .
EXPOSE 50055
CMD [ "./main"]