#!/usr/bin/env bash

go mod tidy
gofmt -w .
go test -v ./...
go vet ./...

