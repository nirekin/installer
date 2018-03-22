#!/bin/sh

docker run --rm -v "$PWD/go":/go/src/installer -w /go/src/installer iron/go:dev go build -o installer

echo http proxy \$1:  $1
echo http2 proxy \$2:  $2

if [ "$1" = "" ]
then
    echo "   \$1 : http_proxy setting is required!"
else
	if [ "$2" = "" ]
	then
		echo "   using only http_proxy..."
		docker build --build-arg http_proxy="$1" --build-arg https_proxy="$1" -t lagoon-platform/installer .     
	else
		echo "   using http_proxy and https_proxy..."
		docker build --build-arg http_proxy="$1" --build-arg https_proxy="$2" -t lagoon-platform/installer .     
	fi		
fi