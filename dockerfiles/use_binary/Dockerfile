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
    && rm -rf /var/lib/apt/lists/* \
    && wget https://github.com/intelsdi-x/snap/releases/download/2.0.0/snap-2.0.0-linux-amd64.tar.gz -P $GOPATH/bin/ \
    && tar -xvzf $GOPATH/bin/snap-2.0.0-linux-amd64.tar.gz -C $GOPATH/bin/ && rm $GOPATH/bin/snap-2.0.0-linux-amd64.tar.gz\
    && wget https://github.com/ami-GS/snap-plugin-collector-mce/releases/download/v_1.0.0/snap-plugin-collector-mce -P $GOPATH/bin/ \
    && wget https://github.com/ami-GS/snap-plugin-collector-mce/releases/download/v_1.0.0/snap-plugin-collector-mce-stream -P $GOPATH/bin/ \
    && wget http://snap.ci.snap-telemetry.io/plugins/snap-plugin-publisher-file/latest/linux/x86_64/snap-plugin-publisher-file -P $GOPATH/bin/
WORKDIR $GOPATH/bin/
CMD tmux -2
