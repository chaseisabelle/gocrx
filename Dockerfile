FROM golang:alpine

WORKDIR /workdir

RUN apk add --update --no-cache chromium

CMD ["chromium", "--help"]