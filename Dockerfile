FROM golang:alpine AS go

WORKDIR /workdir

RUN go install github.com/go-delve/delve/cmd/dlv@latest

RUN apk add --update --no-cache chromium

CMD ["chromium", "--help"]