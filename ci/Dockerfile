FROM ubuntu:17.04

# Golang
RUN apt-get update && apt-get install -y \
    curl \
    wget

ENV VERSION=1.9
ENV OS=linux
ENV ARCH=amd64
RUN wget https://storage.googleapis.com/golang/go$VERSION.$OS-$ARCH.tar.gz
RUN tar -C /usr/local -xzf go$VERSION.$OS-$ARCH.tar.gz

ENV PATH=$PATH:/usr/local/go/bin
ENV GOPATH=$HOME/workspace

# RethinkDB
ENV DISTRIB_CODENAME=zesty
RUN echo "deb http://download.rethinkdb.com/apt $DISTRIB_CODENAME main" | tee /etc/apt/sources.list.d/rethinkdb.list
RUN wget -qO- https://download.rethinkdb.com/apt/pubkey.gpg | apt-key add -
RUN apt-get update && apt-get install -y rethinkdb

# Node
RUN curl -sL https://deb.nodesource.com/setup_6.x | bash -
RUN apt-get install -y nodejs
RUN npm install -g gobble-cli

# Elasticsearch
ENV ES_JAVA_OPTS="-Xms512m -Xmx512m"
ENV PATH /usr/share/elasticsearch/bin:$PATH
ENV ELASTIC_VERSION="2.4.2"

RUN apt-get update && apt-get install -y default-jre 

RUN wget https://download.elastic.co/elasticsearch/release/org/elasticsearch/distribution/deb/elasticsearch/${ELASTIC_VERSION}/elasticsearch-${ELASTIC_VERSION}.deb
RUN dpkg -i elasticsearch-${ELASTIC_VERSION}.deb
RUN sysctl -w vm.max_map_count=262144

# memcache

RUN apt-get install -y \
    memcached \
    openssh-client
