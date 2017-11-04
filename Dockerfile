FROM debian:jessie

LABEL maintainer="tim@townsourced.com"

# Golang
RUN apt-get update && apt-get install -y \
    curl \
    git \
    wget

ENV VERSION=1.9
ENV OS=linux
ENV ARCH=amd64
RUN wget https://storage.googleapis.com/golang/go$VERSION.$OS-$ARCH.tar.gz && \
    tar -C /usr/local -xzf go$VERSION.$OS-$ARCH.tar.gz

ENV PATH=$PATH:/usr/local/go/bin
ENV GOPATH=/go

# Node
ENV HOME=.
RUN curl -sL https://deb.nodesource.com/setup_6.x | bash -
RUN apt-get install -y nodejs
RUN npm install -g gobble-cli


# build townsourced
RUN go get github.com/timshannon/townsourced && \
    cd $GOPATH/src/github.com/timshannon/townsourced && \
    go build -o townsourced && \
    cd web  && \
    npm install && \
    gobble build static -f && \
    cd ../.. && \
    echo '{"web":{"address": "http://localhost:8080"}}' > settings.json

ENTRYPOINT [ "/go/src/github.com/timshannon/townsourced/townsourced" ]