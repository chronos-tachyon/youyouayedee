#!/usr/bin/make -f

.PHONY: all build test cover

all: build test

build:
	go build ./...

test:
	go test ./...

cover:
	go test -v -cover -covermode atomic -coverprofile cover.out ./...
	go tool cover -html cover.out -o cover.html
