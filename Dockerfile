FROM golang:alpine AS go

WORKDIR /workdir

RUN apk add --update --no-cache chromium

CMD ["chromium", "--help"]