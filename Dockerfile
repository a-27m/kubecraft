FROM alpine:3.6 AS wget
RUN apk add --no-cache ca-certificates wget tar

FROM wget AS docker
ARG DOCKER_VERSION=17.09.0-ce
RUN wget -qO- https://download.docker.com/linux/static/stable/x86_64/docker-${DOCKER_VERSION}.tgz | \
  tar -xvz --strip-components=1 -C /bin

FROM wget AS cuberite
ARG CUBERITE_BUILD=905
WORKDIR /srv
RUN wget -qO- "https://builds.cuberite.org/job/Cuberite Linux x64 Master/${CUBERITE_BUILD}/artifact/Cuberite.tar.gz" |\
  tar -xzf -

FROM wget AS kubectl
ARG KUBECTL_VERSION=v1.2.4
WORKDIR /bin
RUN wget -q "https://storage.googleapis.com/kubernetes-release/release/${KUBECTL_VERSION}/bin/linux/amd64/kubectl"


FROM golang:1.12 AS kubecraft
WORKDIR /go/src/kubeproxy
COPY ./go /go
RUN export GO111MODULE=on && go mod init
RUN export GO111MODULE=on && go get k8s.io/client-go@kubernetes-1.15.3
RUN export GO111MODULE=on && CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags '-w' -o kubeproxy ./main.go

FROM debian:jessie
RUN apt-get update; apt-get install -y ca-certificates
COPY --from=kubecraft /go/src/kubeproxy/kubeproxy /bin/goproxy
COPY --from=cuberite /srv /srv
COPY --from=kubectl /bin/kubectl /bin/kubectl

WORKDIR /srv
RUN chmod +x /bin/kubectl

COPY ./docs/img/logo64x64.png /srv/Server/favicon.png
COPY ./Docker /srv/Server/Plugins/Docker
COPY ./config /srv/Server

COPY ./docs/img/logo64x64.png logo.png

EXPOSE 25565
ENTRYPOINT ["/srv/Server/start.sh"]
