FROM golang:1.15-alpine AS builder

MAINTAINER Karthik Pothineni

LABEL service=go-worker

RUN apk add -U --no-cache ca-certificates curl

RUN apk add --no-cache bash

# Install golang linting
RUN wget -O- -nv https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v1.32.0

WORKDIR /etc/go-worker

ENV CGO_ENABLED=0

COPY . .

RUN go mod download

# Run linter
RUN golangci-lint run -v -c golangci.yml

# Run tests
RUN go test -v -cover ./...

# Build application binary
RUN GOOS=linux GOARCH=amd64 go build -a -o worker-svc -v -ldflags '-w' main.go

# Build a small image
FROM scratch

COPY --from=builder /etc/go-worker/worker-svc /

COPY --from=builder /etc/go-worker/config/config.toml /etc/go-worker/config/

ENTRYPOINT ["/worker-svc"]
