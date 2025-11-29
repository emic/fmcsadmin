FROM --platform=linux/x86_64 ubuntu:24.04

RUN apt update && apt install -y curl tar make bash git build-essential

# Install Go
WORKDIR /root

RUN curl -LO https://go.dev/dl/go1.24.10.linux-amd64.tar.gz && tar -C /tmp -xzf go1.24.10.linux-amd64.tar.gz && rm -f /root/go1.24.10.linux-amd64.tar.gz && mv /tmp/go go1.24

WORKDIR /root/go/src/go.googlesource.com/go

RUN git clone https://go.googlesource.com/go goroot && cd goroot && git checkout release-branch.go1.25 && cd src && export GOROOT_BOOTSTRAP=/root/go1.24 && ./all.bash

RUN cp -pr /root/go/src/go.googlesource.com/go/goroot /usr/local/go

RUN /usr/local/go/bin/go version

# Install Goss
WORKDIR /root/go/src/github.com/goss-org

ARG GOSS_VERSION="0.4.9"
ARG GOSS_COMMIT_HASH="5704120d25902119cb1139e04bca3db7742a9f73"

RUN curl -L https://github.com/goss-org/goss/archive/${GOSS_COMMIT_HASH}.tar.gz | tar -xzvf -

RUN mv goss-${GOSS_COMMIT_HASH} goss && cd goss && PATH=$PATH:/usr/local/go/bin TRAVIS_TAG=${GOSS_VERSION} make build
