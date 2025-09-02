# Build stage for frontend assets
FROM oven/bun:1-alpine as frontend-builder
WORKDIR /app
COPY package.json bun.lockb webpack.config.js ./
RUN bun install
COPY web/static/js/main.js ./web/static/js/
RUN bun run build

# Build stage for Go application
FROM golang:1.21-alpine as go-builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o main cmd/server/main.go

# Production stage
FROM alpine:latest
RUN apk --no-cache add ca-certificates sqlite
WORKDIR /root/

# Copy the built application
COPY --from=go-builder /app/main .

# Copy templates and static files
COPY --from=frontend-builder /app/web/static/js/bundle.js ./web/static/js/
COPY web/static/css ./web/static/css/
COPY web/templates ./web/templates/

# Create directory for SQLite database
RUN mkdir -p /data

# Expose port
EXPOSE 8080

# Set environment variable for database path
ENV DB_PATH=/data/dailies.db

# Run the application
CMD ["./main", "serve", "--address", ":8080"]