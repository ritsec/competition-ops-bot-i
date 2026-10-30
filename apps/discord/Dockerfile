FROM golang:1.25.5-bookworm AS builder

WORKDIR /app

# Download dependencies first so they're cached between code changes
COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Static binary so it can run on a minimal image
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /competition-ops-bot-i ./cmd

# Includes CA certificates for Discord's HTTPS API and runs as a non-root user
FROM gcr.io/distroless/static-debian12:nonroot

COPY --from=builder /competition-ops-bot-i /competition-ops-bot-i

# The bot shuts down cleanly on SIGINT, not Docker's default SIGTERM
STOPSIGNAL SIGINT

ENTRYPOINT ["/competition-ops-bot-i"]
