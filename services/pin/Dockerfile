# syntax=docker/dockerfile:1
FROM golang:1.26.1 AS builder

WORKDIR /app

COPY services/pin/go.mod services/pin/go.sum ./services/pin/
COPY go.mod go.sum ./

RUN cd services/pin/ && go mod download

COPY pkg ./pkg

WORKDIR /app/services/pin

COPY services/pin/cmd ./cmd
COPY services/pin/internal ./internal

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /app/app ./cmd/app

#Run
FROM alpine:latest

WORKDIR /root

COPY --from=builder /app/app .

EXPOSE 8080

CMD [ "./app" ]
