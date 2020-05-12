FROM alpine:latest

COPY ./bin/go-retwis /usr/bin/go-retwis

ENTRYPOINT ["go-retwis"]