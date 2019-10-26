FROM alpine:3.10 AS wget
RUN apk add --no-cache ca-certificates wget tar

FROM wget AS cuberite
ARG CUBERITE_BUILD=1057
WORKDIR /opt
RUN wget -qO- "https://builds.cuberite.org/job/Cuberite Linux x64 Master/${CUBERITE_BUILD}/artifact/Cuberite.tar.gz" |\
  tar -xzf -

FROM golang:1.12 AS kubecraft
ARG KUBECTL_VERSION=1.16.2
WORKDIR /go/src/kubeproxy
COPY ./go /go
RUN export GO111MODULE=on && go mod init
RUN export GO111MODULE=on && go get k8s.io/client-go@kubernetes-${KUBECTL_VERSION}
RUN export GO111MODULE=on && CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags '-w' -o kubeproxy ./main.go

FROM debian:buster
RUN apt-get update; apt-get install -y ca-certificates
COPY --from=kubecraft /go/src/kubeproxy/kubeproxy /usr/local/bin/goproxy
COPY --from=cuberite /opt /opt

RUN apt-get update -y && apt-get install curl gnupg lsb-release -y && rm -rf /var/lib/apt/lists/*
RUN export CLOUD_SDK_REPO="cloud-sdk-$(lsb_release -c -s)" && \
    echo "deb http://packages.cloud.google.com/apt $CLOUD_SDK_REPO main" | tee -a /etc/apt/sources.list.d/google-cloud-sdk.list && \
    curl https://packages.cloud.google.com/apt/doc/apt-key.gpg | apt-key add - && \
    apt-get update -y && apt-get install google-cloud-sdk kubectl -y && rm -rf /var/lib/apt/lists/*

WORKDIR /opt

COPY ./docs/img/logo64x64.png /opt/Server/favicon.png
COPY ./Docker /opt/Server/Plugins/Docker
COPY ./config /opt/Server

COPY ./docs/img/logo64x64.png logo.png

EXPOSE 25565
ENTRYPOINT ["/opt/Server/start.sh"]
