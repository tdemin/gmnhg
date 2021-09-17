FROM golang:1.17-alpine AS builder

WORKDIR /src

ENV CGO_ENABLED=0
ENV GO111MODULE=on
ENV GOFLAGS="-trimpath -ldflags=-linkmode=external"

RUN apk add --no-cache git gcc musl-dev

COPY . /src/

RUN \
    go build -v -o /tmp/gmnhg ./cmd/gmnhg && \
    go build -v -o /tmp/md2gmn ./cmd/md2gmn

FROM alpine:3.14 AS runner

LABEL maintainer "Timur Demin <me@tdem.in>"

COPY --from=builder /tmp/gmnhg /tmp/md2gmn /bin/

CMD ["/bin/sh"]
