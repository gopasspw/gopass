DIST := dist
BIN := bin

EXECUTABLE := gopass

PWD := $(shell pwd)
VERSION := $(shell cat VERSION)
SHA := $(shell cat COMMIT 2>/dev/null || git rev-parse --short=8 HEAD)
DATE := $(shell date -u '+%FT%T%z')

GOLDFLAGS += -X "main.Version=$(VERSION)"
GOLDFLAGS += -X "main.BuildTime=$(DATE)"
GOLDFLAGS += -X "main.Commit=$(SHA)"
GOLDFLAGS += -extldflags '-static'

PREFIX ?= /usr
BINDIR ?= $(PREFIX)/bin

GO := CGO_ENABLED=0 go

GOOS ?= $(shell go version | cut -d' ' -f4 | cut -d'/' -f1)
GOARCH ?= $(shell go version | cut -d' ' -f4 | cut -d'/' -f2)

PACKAGES ?= $(shell go list ./... | grep -v /vendor/ | grep -v /tests)

TAGS ?= netgo

.PHONY: all
all: clean test build

.PHONY: clean
clean:
	$(GO) clean -i ./...
	find . -type f -name "coverage.out" -delete
	rm -f gopass_*.deb
	rm -f gopass-*.pkg.tar.xz
	rm -f gopass-*.rpm
	rm -f gopass-*.tar.bz2
	rm -f gopass-*.tar.gz
	rm -f gopass-*-*
	rm -f tests/tests

.PHONY: fmt
fmt:
	$(GO) fmt $(PACKAGES)

.PHONY: tests
tests: test vet lint errcheck

.PHONY: vet
vet:
	$(GO) vet $(PACKAGES)

.PHONY: lint
lint:
	@which golint > /dev/null; if [ $$? -ne 0 ]; then \
		$(GO) get -u github.com/golang/lint/golint; \
	fi
	for PKG in $(PACKAGES); do golint -set_exit_status $$PKG || exit 1; done;

.PHONY: errcheck
errcheck:
	@which errcheck > /dev/null; if [ $$? -ne 0 ]; then \
		$(GO) get -u github.com/kisielk/errcheck; \
	fi
	for PKG in $(PACKAGES); do errcheck $$PKG || exit 1; done;

.PHONY: test
test:
	for PKG in $(PACKAGES); do go test -cover -coverprofile $$GOPATH/src/$$PKG/coverage.out $$PKG || exit 1; done;

.PHONY: test-integration
test-integration: build
	cd tests && GOPASS_BINARY=$(PWD)/$(EXECUTABLE)-$(GOOS)-$(GOARCH) GOPASS_TEST_DIR=$(PWD)/tests go test -v

.PHONY: install
install: build
	install -m 0755 -d $(DESTDIR)$(BINDIR)
	install -m 0755 $(EXECUTABLE)-$(GOOS)-$(GOARCH) $(DESTDIR)$(BINDIR)/gopass

.PHONY: build
build: $(EXECUTABLE)-$(GOOS)-$(GOARCH)

$(EXECUTABLE)-$(GOOS)-$(GOARCH): $(wildcard *.go)
	$(GO) build -tags '$(TAGS)' -ldflags '-s -w $(GOLDFLAGS)' -o gopass-$(GOOS)-$(GOARCH)

.PHONY: release
release: clean
	dist/release.sh

bash.completion: $(EXECUTABLE)-$(GOOS)-$(GOARCH)
	./$(EXECUTABLE)-$(GOOS)-$(GOARCH) completion bash >bash.completion

zsh.completion: $(EXECUTABLE)-$(GOOS)-$(GOARCH)
	./$(EXECUTABLE)-$(GOOS)-$(GOARCH) completion zsh >zsh.completion

.PHONY: completion
completion: bash.completion zsh.completion
