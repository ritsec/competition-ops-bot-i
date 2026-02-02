FROM golang:1.25.5-bookworm AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go mod download

RUN go build -C ./cmd -o /competition-ops-bot-i

CMD ["/competition-ops-bot-i"]