FROM golang:1.22-alpine3.20 AS builder
WORKDIR /app/user_service
COPY user_service/go.mod user_service/go.sum ./
# common/go.sum may not exist
COPY common/go.mod common/go.su[m] ../common/
RUN go mod download
COPY common/. ../common
COPY user_service/. .
RUN go build -o main cmd/main.go

FROM alpine:3.20
WORKDIR /app/user_service
COPY --from=builder /app/user_service/main .
EXPOSE 8081
CMD [ "./main"]