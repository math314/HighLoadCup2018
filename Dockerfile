FROM golang:1.11 AS build-env

# Chosing a working directory
WORKDIR /root

#build goapp
RUN mkdir go && mkdir go/src && mkdir go/bin && mkdir go/pkg && \
    mkdir go/src/hlc2018

ENV GOPATH=/root/go
ADD src go/src
RUN go build -o go/bin/hlc go/src/hlc2018/main.go

# prepare mysql
FROM mysql:5.7
ENV MYSQL_ALLOW_EMPTY_PASSWORD=yes MYSQL_DATABASE=hlc2018 MYSQL_USER=hlc MYSQL_PASSWORD=hlc

ADD docker /root/docker
ENTRYPOINT /root/docker/entrypoint.sh

COPY --from=build-env /root/go/bin/hlc /root/hlc

EXPOSE 3306 8080
