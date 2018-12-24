# rpi-laser-cat-teaser

NAME=rpi-laser-cat-teaser

.PHONY: all
all: build

.PHONY: build
build:
	env CGO_ENABLED=0 GOOS=linux GOARCH=arm go build -o bin/$(NAME) ./cmd
