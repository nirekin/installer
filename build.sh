#!/bin/sh

docker run --rm -v "$PWD/go":/go/src/installer -w /go/src/installer iron/go:dev go build -o installer

docker build -t lagoon-platform/installer . 
