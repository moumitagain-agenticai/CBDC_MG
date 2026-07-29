# Build stage
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Install dependencies
RUN apk add --no-cache git make

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN make build

# Final stage
FROM alpine:3.19

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

# Copy binary and configs
COPY --from=builder /app/bin/server /app/server
COPY --from=builder /app/configs /app/configs

# Create non-root user
RUN adduser -D -u 1001 appuser && chown -R appuser:appuser /app
USER appuser

EXPOSE 8080 9090

ENTRYPOINT ["/app/server"]
CMD ["-config", "/app/configs/config.yaml"]