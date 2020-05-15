FROM golang:latest

RUN mkdir /app

ADD . /app/

WORKDIR /app

RUN go build -o go-retwis .

ENTRYPOINT ["/app/go-retwis"]

EXPOSE 80