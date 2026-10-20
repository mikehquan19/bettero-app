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
COPY ./repositories ./repositories
COPY ./setup ./setup
COPY ./cmd ./cmd
COPY server.go ./

# Compile the API server and the one-shot cron runner
RUN CGO_ENABLED=0 GOOS=linux go build -o ./bettero-app .
RUN CGO_ENABLED=0 GOOS=linux go build -o ./bettero-cron ./cmd/cron

# Deploy the application to lean image
FROM gcr.io/distroless/base-debian11 AS build-release-stage

WORKDIR /

COPY --from=build-stage ./bettero/bettero-app ./bettero-app
COPY --from=build-stage ./bettero/bettero-cron ./bettero-cron

# API Server
EXPOSE 8080
USER nonroot:nonroot

CMD ["./bettero-app"]
