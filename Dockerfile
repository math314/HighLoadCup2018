FROM golang:1.11

ADD src /go/src
WORKDIR /go/src/hlc2018

RUN go build -o /go/bin/hlc2018 .

ENV PORT=80

CMD ["/go/bin/hlc2018"]
