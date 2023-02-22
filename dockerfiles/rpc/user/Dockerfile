FROM golang:1.19 AS builder

LABEL stage=gobuilder

ENV CGO_ENABLED 0
ENV GOPROXY https://goproxy.cn,direct


WORKDIR /app
ADD usersrv .
COPY config/ config/
EXPOSE 8085
CMD ["./usersrv"]
