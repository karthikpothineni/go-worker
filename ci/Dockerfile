FROM golang:1.15-alpine AS builder

USER root

LABEL service=go-worker

RUN apk add -U --no-cache ca-certificates curl

RUN apk add --no-cache bash

WORKDIR /etc/go-worker

COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o worker-svc -v -ldflags '-w' main.go

RUN chmod 755 /etc/go-worker/ci/entrypoint.sh

# Build a small image
FROM scratch

COPY --from=builder /etc/go-worker/worker-svc /

COPY --from=builder /etc/go-worker/config/config.toml /etc/go-worker/config/

ENTRYPOINT ["/worker-svc"]
