# Build stage for Go application
FROM golang:1.25.1-alpine as go-builder
RUN apk update && \
  apk upgrade --no-cache && \
  apk add --no-cache binutils ca-certificates gcc musl-dev sqlite-dev

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download
COPY . .

ENV CGO_ENABLED=1
ENV GOOS=linux
RUN go build -o api

# Production stage
FROM alpine:latest
RUN apk --no-cache add ca-certificates sqlite tzdata
WORKDIR /root/

# Copy the built application
COPY --from=go-builder /app/api .

# Create directory for SQLite database
RUN mkdir -p /data

# Expose port
EXPOSE 8080

# Set environment variable for database path
ENV DB_PATH=/data/dailies.db

# Run the application
CMD ["sh", "-c", "./api -port 8080 -db-path ${DB_PATH}"]
