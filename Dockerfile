# Stage 1: Build the application
FROM golang:1.24.2 AS builder

ENV GO111MODULE=on \
CGO_ENABLED=0 \
GOOS=linux \
GOARCH=amd64

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download
RUN go install github.com/swaggo/swag/cmd/swag@latest

COPY . .

RUN swag init -g cmd/main.go
RUN go build -ldflags="-w -s" -o /user-order-api cmd/main.go

# Stage 2: Create the final lightweight image
FROM alpine:latest

RUN apk update && apk add --no-cache ca-certificates tzdata

WORKDIR /app

COPY --from=builder /user-order-api /app/user-order-api
# COPY .env .env

EXPOSE 8080

ENTRYPOINT ["/app/user-order-api"]
