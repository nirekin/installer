#!/bin/sh

docker run --rm -v "$PWD/go":/go/src/installer -w /go/src/installer iron/go:dev go build -o installer


echo Proxy  $1

docker build --build-arg http_proxy=$1 --build-arg https_proxy=$1 -t lagoon-platform/installer . 
