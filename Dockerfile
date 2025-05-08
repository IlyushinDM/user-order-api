# Stage 1: Build the application
FROM golang:1.24.2 AS builder

ENV GO111MODULE=on \
    CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download && \
    go mod verify && \
    go install github.com/swaggo/swag/cmd/swag@latest

COPY cmd/ cmd/
COPY internal/ internal/
RUN swag init -g cmd/main.go && \
    go build -ldflags="-w -s" -o /user-order-api cmd/main.go

# Stage 2: Final image
FROM alpine:latest
RUN apk --no-cache add ca-certificates tzdata
WORKDIR /app
COPY --from=builder /user-order-api .
ENV TZ=Etc/UTC
RUN ln -snf /usr/share/zoneinfo/$TZ /etc/localtime && echo $TZ > /etc/timezone
EXPOSE 8080
ENTRYPOINT ["/app/user-order-api"]
