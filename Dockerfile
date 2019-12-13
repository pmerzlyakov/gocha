FROM golang:1.13

WORKDIR /go/src/github.com/pmerzlyakov/gocha
COPY chat ./chat
COPY public ./public
COPY config.json main.go ./

RUN go get -d -v ./...
RUN go install -v ./...

CMD ["gocha"]