FROM golang:1.22

RUN apt update && apt install curl tar make bash git

WORKDIR /go/src/github.com/goss-org

ARG GOSS_VERSION="0.4.8"
ARG GOSS_COMMIT_HASH="aed56336c3e8ff683e9540065b502f423dd6760d"

RUN curl -L https://github.com/goss-org/goss/archive/${GOSS_COMMIT_HASH}.tar.gz | tar -xzvf -

RUN mv goss-${GOSS_COMMIT_HASH} goss && cd goss && TRAVIS_TAG=${GOSS_VERSION} make build
