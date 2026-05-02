FROM --platform=linux/x86_64 ubuntu:24.04@sha256:c4a8d5503dfb2a3eb8ab5f807da5bc69a85730fb49b5cfca2330194ebcc41c7b

RUN apt update && apt install -y curl tar make bash git build-essential

# Install Go 1.26.2
WORKDIR /root

RUN echo "00859d7bd6defe8bf84d9db9e57b9a4467b2887c18cd93ae7460e713db774bc1 go1.25.9.linux-amd64.tar.gz" > SHA256SUMS && curl -LO https://go.dev/dl/go1.25.9.linux-amd64.tar.gz && sha256sum -c SHA256SUMS && tar -C /tmp -xzf go1.25.9.linux-amd64.tar.gz && rm -f /root/go1.25.9.linux-amd64.tar.gz && mv /tmp/go go1.25

WORKDIR /root/go/src/go.googlesource.com/go

RUN git clone https://go.googlesource.com/go goroot && cd goroot && git checkout 9c8bf0e72a6fb3b415b591b124b59fbb7cf92252 && cd src && export GOROOT_BOOTSTRAP=/root/go1.25 && ./all.bash

RUN cp -pr /root/go/src/go.googlesource.com/go/goroot /usr/local/go

RUN /usr/local/go/bin/go version

# Install Goss 0.4.9
WORKDIR /root/go/src/github.com/goss-org

ARG GOSS_COMMIT_HASH="5704120d25902119cb1139e04bca3db7742a9f73"

RUN curl -L https://github.com/goss-org/goss/archive/${GOSS_COMMIT_HASH}.tar.gz | tar -xzvf -

RUN mv "goss-${GOSS_COMMIT_HASH}" goss && cd goss && PATH=$PATH:/usr/local/go/bin TRAVIS_TAG=0.4.9 make build
