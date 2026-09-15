# syntax=docker/dockerfile:1
FROM golang:1.26.1 AS builder

WORKDIR /app

COPY services/interaction/go.mod services/interaction/go.sum ./services/interaction/
COPY go.mod go.sum ./

RUN cd services/interaction/ && go mod download

COPY pkg ./pkg

WORKDIR /app/services/interaction

COPY services/interaction/cmd ./cmd
COPY services/interaction/internal ./internal

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /app/app ./cmd/app

#Run
FROM alpine:latest

WORKDIR /root

COPY --from=builder /app/app .

EXPOSE 8080

CMD [ "./app" ]
