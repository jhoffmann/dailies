# Build stage for Go application
FROM golang:1.24.6-alpine as go-builder
RUN apk update && \
  apk upgrade --no-cache && \
  apk add --no-cache binutils ca-certificates gcc musl-dev sqlite-dev

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download
COPY . .

ENV CGO_ENABLED=1
ENV GOOS=linux
RUN go build -o server cmd/server/main.go

# Production stage
FROM alpine:latest
RUN apk --no-cache add ca-certificates sqlite
WORKDIR /root/

# Copy the built application
COPY --from=go-builder /app/server .

# Copy templates and static files
COPY web/static/js/bundle.js ./web/static/js/
COPY web/static/css ./web/static/css/
COPY web/templates ./web/templates/

# Create directory for SQLite database
RUN mkdir -p /data

# Expose port
EXPOSE 8080

# Set environment variable for database path
ENV DB_PATH=/data/dailies.db

# Run the application
CMD ["sh", "-c", "./server --address :8080 --db ${DB_PATH}"]
