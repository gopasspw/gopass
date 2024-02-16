FROM golang:1.22-alpine@sha256:8e96e6cff6a388c2f70f5f662b64120941fcd7d4b89d62fec87520323a316bd9 AS build-env

ENV CGO_ENABLED 0

RUN apk add --no-cache make git ncurses

# Build gopass
WORKDIR /home/runner/work/gopass/gopass

COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . .

ARG golags_arg=""
ENV GOFLAGS=$goflags_arg

RUN make clean
RUN make gopass

# Build gopass-jsonapi
WORKDIR /home/runner/work/gopass

RUN git clone https://github.com/gopasspw/gopass-jsonapi.git

WORKDIR /home/runner/work/gopass/gopass-jsonapi
RUN go mod download
RUN make clean
RUN make gopass-jsonapi

# Build gopass-hibp
WORKDIR /home/runner/work/gopass

RUN git clone https://github.com/gopasspw/gopass-hibp.git

WORKDIR /home/runner/work/gopass/gopass-hibp
RUN go mod download
RUN make clean
RUN make gopass-hibp

# Build gopass-summon-provider
WORKDIR /home/runner/work/gopass

RUN git clone https://github.com/gopasspw/gopass-summon-provider.git

WORKDIR /home/runner/work/gopass/gopass-summon-provider
RUN go mod download
RUN make clean
RUN make gopass-summon-provider

# Build git-credential-gopass
WORKDIR /home/runner/work/gopass

RUN git clone https://github.com/gopasspw/git-credential-gopass.git

WORKDIR /home/runner/work/gopass/git-credential-gopass
RUN go mod download
RUN make clean
RUN make git-credential-gopass

FROM alpine@sha256:c5b1261d6d3e43071626931fc004f70149baeba2c8ec672bd4f27761f8e1ad6b
RUN apk add --no-cache ca-certificates git gnupg
COPY --from=build-env /home/runner/work/gopass/gopass/gopass /usr/local/bin/
COPY --from=build-env /home/runner/work/gopass/gopass-jsonapi/gopass-jsonapi /usr/local/bin/
COPY --from=build-env /home/runner/work/gopass/gopass-hibp/gopass-hibp /usr/local/bin/
COPY --from=build-env /home/runner/work/gopass/gopass-summon-provider/gopass-summon-provider /usr/local/bin/
COPY --from=build-env /home/runner/work/gopass/git-credential-gopass/git-credential-gopass /usr/local/bin/
