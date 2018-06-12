FROM golang:1.10 AS builder

RUN apt-get update && \
  apt-get install --yes --force-yes \
  make \
  git \
  gnupg

ADD .   /go/src/github.com/gopasspw/gopass
WORKDIR /go/src/github.com/gopasspw/gopass

RUN make install

RUN chown -R 1000:1000 /go/src/github.com/gopasspw/gopass
ENV HOME /go/src/github.com/gopasspw/gopass
USER 1000:1000

CMD [ "/go/src/github.com/gopasspw/gopass/gopass" ]
