FROM alpine:3.7

RUN apk add --no-cache ansible

RUN mkdir -p /opt/lagoon/bin
COPY ./go/installer /opt/lagoon/bin/installer

ENTRYPOINT ["/opt/lagoon/bin/installer"]


