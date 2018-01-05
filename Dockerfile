FROM golang:latest
MAINTAINER Daiki Aminaka <1991.daiki@gmail.com>

ENV GOPATH=/go
ENV HOME=/go
WORKDIR /go
RUN apt-get update \
    && apt-get install -y \
    locales \
    mcelog \
    tmux \
    && locale-gen en_US.UTF-8 \
    && echo "en_US.UTF-8 UTF-8" >> /etc/locale.gen \
    && locale-gen \
    && cd $HOME \
    && rm -rf /var/lib/apt/lists/* \
    && go get -d github.com/intelsdi-x/snap \
    && make -C $GOPATH/src/github.com/intelsdi-x/snap \
    && make install -C $GOPATH/src/github.com/intelsdi-x/snap \
    && go get github.com/ami-GS/snap-plugin-collector-mce \
    && go get -v ./src/github.com/ami-GS/snap-plugin-collector-mce/... \
    && make -C $GOPATH/src/github.com/ami-GS/snap-plugin-collector-mce \
    && cp $GOPATH/src/github.com/ami-GS/snap-plugin-collector-mce/build/Linux/x86_64/snap-plugin-collector-mce $GOPATH/bin \
    && make stream -C $GOPATH/src/github.com/ami-GS/snap-plugin-collector-mce \
    && cp $GOPATH/src/github.com/ami-GS/snap-plugin-collector-mce/build/Linux/x86_64/snap-plugin-collector-mce-stream $GOPATH/bin \
    && cp $GOPATH/src/github.com/ami-GS/snap-plugin-collector-mce/testlog/mcelog1 /var/log/mcelog \
    && wget http://snap.ci.snap-telemetry.io/plugins/snap-plugin-publisher-file/latest/linux/x86_64/snap-plugin-publisher-file -P $GOPATH/bin/
WORKDIR $GOPATH/bin/
CMD tmux -2
