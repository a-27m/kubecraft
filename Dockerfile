FROM alpine AS wget
RUN apk add --no-cache ca-certificates wget tar

FROM wget AS cuberite
ARG CUBERITE_BUILD=1057
WORKDIR /opt
# TODO: replace with a specific version, use CUBERITE_BUILD
RUN wget -qO- "https://download.cuberite.org/linux-x86_64/Cuberite.tar.gz" | \
  tar -xzf -

FROM golang:1.24 AS kubecraft
ARG KUBECTL_VERSION=1.32.3
WORKDIR /go/src/kubeproxy
COPY ./go /go
RUN go get k8s.io/client-go@kubernetes-${KUBECTL_VERSION}
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags '-w' -o kubeproxy ./main.go

FROM debian:bookworm-slim
COPY --from=kubecraft /go/src/kubeproxy/kubeproxy /usr/local/bin/kubeproxy
COPY --from=cuberite /opt /opt/Server

RUN apt-get update -y \
    && apt-get install -y \
      ca-certificates \
      curl \
      gnupg \
      lsb-release \
    && rm -rf /var/lib/apt/lists/*
RUN export CLOUD_SDK_REPO="cloud-sdk-$(lsb_release -c -s)" && \
    echo "deb http://packages.cloud.google.com/apt $CLOUD_SDK_REPO main" | tee -a /etc/apt/sources.list.d/google-cloud-sdk.list && \
    curl https://packages.cloud.google.com/apt/doc/apt-key.gpg | apt-key add - && \
    apt-get update -y && apt-get install kubectl -y && rm -rf /var/lib/apt/lists/*

WORKDIR /opt

COPY ./docs/img/logo64x64.png /opt/Server/favicon.png
COPY ./Docker /opt/Server/Plugins/Docker
COPY ./config /opt/Server

COPY ./docs/img/logo64x64.png logo.png

EXPOSE 25565
ENTRYPOINT ["/opt/Server/start.sh"]
