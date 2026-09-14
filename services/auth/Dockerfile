# syntax=docker/dockerfile:1
FROM golang:1.26.1 AS builder

WORKDIR /app

COPY services/auth/go.mod services/auth/go.sum ./services/auth/
COPY go.mod go.sum ./

RUN cd services/auth/ && go mod download

COPY pkg ./pkg

WORKDIR /app/services/auth

COPY services/auth/cmd ./cmd
COPY services/auth/internal ./internal

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /app/app ./cmd/app

#Run
FROM alpine:latest

WORKDIR /root

COPY --from=builder /app/app .

EXPOSE 8080

CMD [ "./app" ]