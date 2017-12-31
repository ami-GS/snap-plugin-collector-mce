FROM golang:latest

ENV GOPATH=/go
ENV HOME=/go
WORKDIR /go
COPY /Users/daminaka/Go/src/github.com/intelsdi-x/snap /go/src/github.com/intelsdi-x/snap
RUN apt-get update \
    && apt-get install -y \
    locales \
    mcelog \
    && locale-gen en_US.UTF-8 \
    && echo "en_US.UTF-8 UTF-8" >> /etc/locale.gen \
    && locale-gen \
    && cd $HOME \
    && rm -rf /var/lib/apt/lists/* \
    && go get github.com/intelsdi-x/gomit \
    && go get github.com/Masterminds/glide \
    && go get github.com/go-swagger/go-swagger/cmd/swagger \
    && go get github.com/intelsdi-x/snap-plugin-collector-cpu \
    && make -C $GOPATH/src/github.com/intelsdi-x/snap-plugin-collector-cpu \
    && go get github.com/ami-GS/snape-plugin-collector-mce \
    && make -C $GOPATH/src/github.com/intelsdi-x/snap-plugin-collector-mce \
    && make stream -C $GOPATH/src/github.com/intelsdi-x/snap-plugin-collector-mce \
    && cp $GOPATH/src/github.com/intelsdi-x/snap-plugin-collector-mce/testlog/mcelog1 /var/log/mcelog \
    && wget http://snap.ci.snap-telemetry.io/plugins/snap-plugin-publisher-file/latest/linux/x86_64/snap-plugin-publisher-file -P $GOPATH/bin/snap-plugin-publisher-file
WORKDIR $GOPATH/src/github.com/intelsdi-x/snap