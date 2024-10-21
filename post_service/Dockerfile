FROM golang:1.22-alpine3.20 AS builder
WORKDIR /app/post_service
COPY post_service/go.mod post_service/go.sum ./
# common/go.sum may not exist
COPY common/go.mod common/go.su[m] ../common/
RUN go mod download
COPY common/. ../common
COPY post_service/. .
RUN go build -o main cmd/main.go

FROM alpine:3.20
WORKDIR /app/post_service
COPY --from=builder /app/post_service/main .
EXPOSE 8082
CMD [ "./main"]