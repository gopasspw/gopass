FROM docker.io/library/golang:1.24-alpine@sha256:7772cb5322baa875edd74705556d08f0eeca7b9c4b5367754ce3f2f00041ccee AS build-env

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

FROM docker.io/library/alpine@sha256:a8560b36e8b8210634f77d9f7f9efd7ffa463e380b75e2e74aff4511df3ef88c
RUN apk add --no-cache ca-certificates git gnupg
COPY --from=build-env /home/runner/work/gopass/gopass/gopass /usr/local/bin/
COPY --from=build-env /home/runner/work/gopass/gopass-jsonapi/gopass-jsonapi /usr/local/bin/
COPY --from=build-env /home/runner/work/gopass/gopass-hibp/gopass-hibp /usr/local/bin/
COPY --from=build-env /home/runner/work/gopass/gopass-summon-provider/gopass-summon-provider /usr/local/bin/
COPY --from=build-env /home/runner/work/gopass/git-credential-gopass/git-credential-gopass /usr/local/bin/
