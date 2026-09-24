# syntax=docker/dockerfile:1
FROM golang:1.26.1 AS builder

WORKDIR /app

COPY services/profile/go.mod services/profile/go.sum ./services/profile/
COPY go.mod go.sum ./

RUN cd services/profile/ && go mod download

COPY pkg ./pkg

WORKDIR /app/services/profile

COPY services/profile/cmd ./cmd
COPY services/profile/internal ./internal

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /app/app ./cmd/app

#Run
FROM alpine:latest

WORKDIR /root

COPY --from=builder /app/app .

EXPOSE 8080

CMD [ "./app" ]
