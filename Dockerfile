# Use an official Golang image as the builder
FROM golang:1.26-alpine AS builder

# Set the working directory inside the container
WORKDIR /app

# Copy the go.mod and go.sum files to the working directory
COPY go.mod go.sum ./

RUN go mod download

# Copy the rest of the application source code to the working directory
COPY cmd/ ./cmd/
COPY internal/ ./internal/

RUN CGO_ENABLED=0 GOOS=linux go build -o realtime-engine ./cmd/server/main.go

FROM alpine:latest

WORKDIR /root/

COPY --from=builder /app/realtime-engine .

EXPOSE 8080

CMD ["./realtime-engine"]