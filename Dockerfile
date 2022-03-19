FROM docker.io/library/golang:1.18-alpine AS builder

WORKDIR /src

ENV CGO_ENABLED=0

RUN apk add --no-cache git gcc musl-dev

COPY . /src/

ARG gmnhg_version=v0+HEAD
RUN go build \
    -v \
    -trimpath \
    -ldflags="-linkmode=external -X main.version=${gmnhg_version}" \
    -o /tmp/ \
    ./cmd/gmnhg \
    ./cmd/md2gmn

FROM docker.io/library/alpine:3.15 AS runner

LABEL maintainer "Timur Demin <me@tdem.in>"

COPY --from=builder /tmp/gmnhg /tmp/md2gmn /bin/

CMD ["/bin/sh"]
