DIST := dist
BIN := bin

EXECUTABLE := gopass

PWD := $(shell pwd)
VERSION := $(shell cat VERSION)
SHA := $(shell cat COMMIT 2>/dev/null || git rev-parse --short=8 HEAD)
DATE := $(shell date -u '+%FT%T%z')

GOLDFLAGS += -X "main.version=$(VERSION)"
GOLDFLAGS += -X "main.date=$(DATE)"
GOLDFLAGS += -X "main.commit=$(SHA)"
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
tests: test-cross test vet lint errcheck megacheck

.PHONY: vet
vet:
	$(GO) vet $(PACKAGES)

.PHONY: lint
lint:
	@which golint > /dev/null; if [ $$? -ne 0 ]; then \
		$(GO) get -u github.com/golang/lint/golint; \
	fi
	STATUS=0; for PKG in $(PACKAGES); do golint -set_exit_status $$PKG || STATUS=1; done; exit $$STATUS

.PHONY: errcheck
errcheck:
	@which errcheck > /dev/null; if [ $$? -ne 0 ]; then \
		$(GO) get -u github.com/kisielk/errcheck; \
	fi
	STATUS=0; for PKG in $(PACKAGES); do errcheck $$PKG || STATUS=1; done; exit $$STATUS

.PHONY: megacheck
megacheck:
	@which megacheck > /dev/null; if [ $$? -ne 0  ]; then \
		$(GO) get -u honnef.co/go/tools/cmd/megacheck; \
	fi
	STATUS=0; for PKG in $(PACKAGES); do megacheck $$PKG || STATUS=1; done; exit $$STATUS

.PHONY: test
test:
	STATUS=0; for PKG in $(PACKAGES); do go test -cover -coverprofile $$GOPATH/src/$$PKG/coverage.out $$PKG || STATUS=1; done; exit $$STATUS

.PHONY: test-cross
test-cross:
	GOOS=linux GOARCH=amd64 $(GO) build
	GOOS=darwin GOARCH=amd64 $(GO) build
	GOOS=windows GOARCH=amd64 $(GO) build

.PHONY: test-integration
test-integration: clean build
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
