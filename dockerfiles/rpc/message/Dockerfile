FROM golang:1.19 AS builder

LABEL stage=gobuilder

ENV CGO_ENABLED 0
ENV GOPROXY https://goproxy.cn,direct

RUN mkdir -p /home/crypt/
COPY rsa_public_key.pem /home/crypt/
COPY rsa_private_key.pem /home/crypt/

WORKDIR /app
ADD messagesrv .
COPY config/ config/
EXPOSE 8083
CMD ["./messagesrv"]