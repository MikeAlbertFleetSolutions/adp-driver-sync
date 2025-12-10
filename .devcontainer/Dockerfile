FROM ubuntu:20.04

ENV DEBIAN_FRONTEND=noninteractive

# system updates and dependencies
RUN apt-get update &&\
  apt-get -y install \
  jq \
  curl \
  groff \
  unzip \
  vim \
  zip \
  git \
  build-essential \
  software-properties-common \
  libssl-dev \
  libbz2-dev \
  libffi-dev \
  libreadline-dev \
  libsqlite3-dev \
  zlib1g-dev \
  g++ \
  gcc \
  libc6-dev \
  make \
  pkg-config \
  shellcheck \
  net-tools \
  openssh-client \
  libaio1

# golang
COPY --from=golang:1.24 /usr/local/go /usr/local/go
ENV GOPATH /go
ENV GOROOT /usr/local/go
ENV PATH $GOPATH/bin:/usr/local/go/bin:$PATH
RUN mkdir -p "$GOPATH/src" "$GOPATH/bin" && chmod -R 777 "$GOPATH"

# Go language server & tools for visual studio code
RUN go install golang.org/x/tools/gopls@latest
RUN go install github.com/go-delve/delve/cmd/dlv@latest
RUN cp /go/bin/dlv /go/bin/dlv-dap
RUN go install github.com/uudashr/gopkgs/v2/cmd/gopkgs@latest
RUN go install github.com/ramya-rao-a/go-outline@latest
RUN go install golang.org/x/vuln/cmd/govulncheck@latest
RUN curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v1.64.5

# docker binaries
COPY --from=docker:20.10 /usr/local/bin/docker /usr/local/bin/docker

# docker-compose binaries
COPY --from=docker/compose:1.29.2 /usr/local/bin/docker-compose /usr/local/bin/docker-compose

# niceties
RUN chmod -x /usr/bin/dircolors &&\
  echo "syntax off" > /root/.vimrc &&\
  echo "alias ll='ls -alF'" >> /root/.bashrc

# clean up
RUN apt-get autoremove -y \
  && apt-get clean -y \
  && rm -rf /var/lib/apt/lists/*
ENV DEBIAN_FRONTEND=dialog

# set the default shell to bash instead of sh
ENV SHELL /bin/bash
