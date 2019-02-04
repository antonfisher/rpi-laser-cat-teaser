# rpi-laser-cat-teaser

NAME=rpi-laser-cat-teaser

VERSION ?= $(shell git rev-parse --abbrev-ref HEAD | sed -e "s/.*\\///")
COMMIT ?= $(shell git rev-parse HEAD | cut -c 1-7)
LDFLAGS ?= \
	-X github.com/antonfisher/rpi-laser-cat-teaser/pkg/params.Name=${NAME} \
	-X github.com/antonfisher/rpi-laser-cat-teaser/pkg/params.Version=${VERSION} \
	-X github.com/antonfisher/rpi-laser-cat-teaser/pkg/params.Commit=${COMMIT}

.PHONY: all
all: build

.PHONY: build
build: build-rpi

.PHONY: build-rpi
build-rpi:
	env CGO_ENABLED=0 GOOS=linux GOARCH=arm go build -o bin/$(NAME) -ldflags "$(LDFLAGS)" ./cmd

.PHONY: build-amd64
build-amd64:
	env CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o bin/$(NAME) -ldflags "$(LDFLAGS)" ./cmd
