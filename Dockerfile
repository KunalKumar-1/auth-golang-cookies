# ---------- Build stage ----------
FROM golang:1.25.5-alpine3.23 AS builder

WORKDIR /app

# Install git (needed for some Go modules)
RUN apk add --no-cache git

# Cache dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build static binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -o app cmd/main.go


# ---------- Runtime stage ----------
FROM alpine:3.23

WORKDIR /app

# TLS certificates for HTTPS, DB, APIs
RUN apk add --no-cache ca-certificates

# Copy binary from builder
COPY --from=builder /app/app .

# Expose app port
EXPOSE 4000

# Run the app
CMD ["./app"]