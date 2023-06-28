# syntax=docker/dockerfile:1

# use golang alpine image
FROM golang:1.18-alpine AS builder

# Move to working directory.
WORKDIR /build

# Copy and download dependency using go mod. @TODO: should only copy files needed to build binary
COPY . .        
RUN go mod download

# Build binary
RUN go build ./go-twitch-web/api/main.go


# Run main on start
ENTRYPOINT ["./main"]
