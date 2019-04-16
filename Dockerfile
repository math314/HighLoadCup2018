FROM golang:1.11

# Chosing a working directory
WORKDIR /root

EXPOSE 8080

RUN mkdir go && mkdir go/src && mkdir go/bin && mkdir go/pkg && \
    mkdir go/src/hlc2018

# Setting environment variables for Go
ENV GOPATH=/root/go
ADD src/ go/src
RUN go build go/src/hlc2018/main.go

# Launching our server
CMD go/src/hlc2018/main