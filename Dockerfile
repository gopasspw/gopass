FROM golang:1.9-alpine AS builder

RUN apk add -U make git gnupg

ADD .   /go/src/github.com/justwatchcom/gopass
WORKDIR /go/src/github.com/justwatchcom/gopass

RUN make install

CMD [ "/go/src/github.com/justwatchcom/gopass/gopass" ]
