FROM golang:1.10-alpine AS builder

RUN apk add -U make gcc musl-dev ncurses git

ADD .   /go/src/github.com/gopasspw/gopass
WORKDIR /go/src/github.com/gopasspw/gopass

RUN TERM=vt100 make install

FROM alpine:3.7
RUN apk add -U git gnupg
COPY --from=0 /go/src/github.com/gopasspw/gopass /usr/bin/

RUN chown -Rh 1000:1000 -- /root
ENV HOME /root
USER 1000:1000
ENTRYPOINT [ "/usr/bin/gopass" ]
