FROM golang:1.25.5-bookworm

WORKDIR /app

COPY . .

RUN go mod download

RUN go build -C ./cmd -o /competition-ops-bot-i

CMD ["/competition-ops-bot-i"]