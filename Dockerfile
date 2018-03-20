FROM alpine:3.7

ENV http_proxy 'http://E518546:Toubin05@internetpsa.inetpsa.com:80'
ENV https_proxy 'http://E518546:Toubin05@internetpsa.inetpsa.com:80'

RUN apk add --no-cache ansible

RUN mkdir -p /opt/lagoon/bin
COPY ./go/installer /opt/lagoon/bin/installer

ENTRYPOINT ["/opt/lagoon/bin/installer"]


