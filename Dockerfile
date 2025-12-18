FROM golang:1.25-alpine AS builder

ENV GOTOOLCHAIN=auto

# Set environment
ENV CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64

# Set working directory
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy seluruh source code
COPY . .

# Build binary
RUN go build -o dist/main ./cmd/main/main.go

# ---------- Stage 2: Runtime ----------
FROM alpine:latest

WORKDIR /app

# Copy binary dari builder
COPY --from=builder /app/dist ./dist

# Copy file .env kalau mau bawa ke container
# COPY .env .env

EXPOSE 8080

# Jalankan aplikasi
CMD ["./dist/main"]
