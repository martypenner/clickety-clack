# Define build arguments
ARG GO_VERSION=1.23.1
ARG BINARY_NAME=clickety-clack

# Base builder
FROM golang:${GO_VERSION}-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
COPY . .

# Build Linux binary
FROM --platform=linux/amd64 golang:${GO_VERSION}-alpine AS build-linux
RUN apk add --no-cache gcc musl-dev libx11-dev libxtst-dev libxkbcommon-dev
WORKDIR /app
COPY --from=builder /app .
ENV GOOS=linux
ENV GOARCH=amd64
ENV CGO_ENABLED=1

# Build macOS binary
FROM golang:${GO_VERSION}-alpine AS build-macos
RUN apk add --no-cache clang
WORKDIR /app
COPY --from=builder /app .
ENV GOOS=darwin
ENV GOARCH=amd64
ENV CGO_ENABLED=1

# Build Windows binary
FROM golang:${GO_VERSION} AS build-windows
RUN apt-get update && apt-get install -y mingw-w64 && \
  apt-get clean && rm -rf /var/lib/apt/lists/*
WORKDIR /app
COPY go.mod go.sum ./
COPY . .
ENV GOOS=windows
ENV GOARCH=amd64
ENV CGO_ENABLED=1
ENV CC=x86_64-w64-mingw32-gcc
ENV CXX=x86_64-w64-mingw32-g++
