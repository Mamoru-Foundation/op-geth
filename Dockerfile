# Support setting various labels on the final image
ARG COMMIT=""
ARG VERSION=""
ARG BUILDNUM=""

# Build Geth in a stock Go builder container
FROM golang:1.21 as builder

RUN apt-get update  \
    && apt-get install -y gcc musl-dev git curl tar libc6-dev

# Get dependencies - will also be cached if we won't change go.mod/go.sum
COPY go.mod /go-ethereum/
COPY go.sum /go-ethereum/
RUN cd /go-ethereum && go mod download

ADD . /go-ethereum
RUN cd /go-ethereum && GO111MODULE=on go run build/ci.go install ./cmd/geth

# Pull Geth into a second stage deploy alpine container
FROM debian:12.0-slim

COPY docker/cron/cron.conf /etc/cron.d/cron.conf
COPY docker/cron/prune.sh /prune.sh
COPY docker/supervisord/supervisor.conf /etc/supervisor/conf.d/supervisord.conf
COPY docker/docker-entrypoint.sh /docker-entrypoint.sh
RUN chmod +x /docker-entrypoint.sh

# Install Supervisor and create the Unix socket
RUN touch /var/run/supervisor.sock

RUN apt-get update  \
    && apt-get install -y ca-certificates jq unzip bash grep curl sed htop procps cron supervisor \
    && apt-get clean \
    && crontab /etc/cron.d/cron.conf

COPY --from=builder /go-ethereum/build/bin/geth /usr/local/bin/

EXPOSE 8545 8546 30303 30303/udp
ENTRYPOINT ["/docker-entrypoint.sh"]

# Add some metadata labels to help programatic image consumption
ARG COMMIT=""
ARG VERSION=""
ARG BUILDNUM=""

LABEL commit="$COMMIT" version="$VERSION" buildnum="$BUILDNUM"
