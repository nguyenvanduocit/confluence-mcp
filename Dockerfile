FROM golang:1.23-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o confluence-mcp .

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/confluence-mcp .

# Expose port for HTTP server (optional)
EXPOSE 8080

ENTRYPOINT ["/app/confluence-mcp"]
CMD [] 