FROM --platform=linux/x86_64 ubuntu:22.04

RUN apt update && apt install -y curl tar make bash git build-essential

# Install Go
WORKDIR /root

RUN curl -LO https://go.dev/dl/go1.22.6.linux-amd64.tar.gz && tar -C /tmp -xzf go1.22.6.linux-amd64.tar.gz && rm -f /root/go1.22.6.linux-amd64.tar.gz && mv /tmp/go go1.22

WORKDIR /root/go/src/go.googlesource.com/go

RUN git clone https://go.googlesource.com/go goroot && cd goroot && git checkout release-branch.go1.23 && cd src && export GOROOT_BOOTSTRAP=/root/go1.22 && ./all.bash

RUN cp -pr /root/go1.22 /usr/local/go

# Install Goss
WORKDIR /root/go/src/github.com/goss-org

ARG GOSS_VERSION="0.4.8"
ARG GOSS_COMMIT_HASH="aed56336c3e8ff683e9540065b502f423dd6760d"

RUN curl -L https://github.com/goss-org/goss/archive/${GOSS_COMMIT_HASH}.tar.gz | tar -xzvf -

RUN mv goss-${GOSS_COMMIT_HASH} goss && cd goss && PATH=$PATH:/usr/local/go/bin TRAVIS_TAG=${GOSS_VERSION} make build
