# syntax=docker/dockerfile:1

# Base image for Go
FROM golang:1.26-alpine AS build-stage

# Build the work directory where all the codes are located
WORKDIR /bettero

# Copy and download the dependencies to image
COPY go.mod go.sum ./
RUN go mod download

# Copy source code to image
COPY ./controllers ./controllers
COPY ./models ./models
COPY ./routes ./routes
COPY ./services ./services
COPY ./setup ./setup
COPY server.go ./

# Compile the server
RUN CGO_ENABLED=0 GOOS=linux go build -o ./bettero-app .

# Deploy the application to lean image
FROM gcr.io/distroless/base-debian11 AS build-release-stage

WORKDIR /

COPY --from=build-stage ./bettero/bettero-app ./bettero-app

# API Server
EXPOSE 8080
USER nonroot:nonroot

CMD ["./bettero-app"]