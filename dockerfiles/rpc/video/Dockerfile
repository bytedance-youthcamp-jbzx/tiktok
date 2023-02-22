FROM golang:1.19 AS builder

LABEL stage=gobuilder

ENV CGO_ENABLED 0
ENV GOPROXY https://goproxy.cn,direct


WORKDIR /app
ADD videosrv .
COPY config/ config/
EXPOSE 8086
CMD ["./videosrv"]
