FROM golang:1.9-alpine AS builder

RUN apk add -U make git gnupg

ADD .   /go/src/github.com/justwatchcom/gopass
WORKDIR /go/src/github.com/justwatchcom/gopass

RUN make install

RUN chown -R 1000:1000 /go/src/github.com/justwatchcom/gopass
ENV HOME /go/src/github.com/justwatchcom/gopass
USER 1000:1000

CMD [ "/go/src/github.com/justwatchcom/gopass/gopass" ]
