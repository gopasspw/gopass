FIRST_GOPATH              := $(firstword $(subst :, ,$(GOPATH)))
PKGS                      := $(shell go list ./... | grep -v /tests)
GOFILES_NOVENDOR          := $(shell find . -type f -name '*.go' -not -path "./vendor/*")
GOFILES_NOTEST            := $(shell find . -type f -name '*.go' -not -path "./vendor/*" -not -name "*_test.go")
GOPASS_VERSION            ?= $(shell cat VERSION)
GOPASS_OUTPUT             ?= gopass
GOPASS_REVISION           := $(shell cat COMMIT 2>/dev/null || git rev-parse --short=8 HEAD)
BASH_COMPLETION_OUTPUT    := bash.completion
FISH_COMPLETION_OUTPUT    := fish.completion
ZSH_COMPLETION_OUTPUT     := zsh.completion
# Support reproducible builds by embedding date according to SOURCE_DATE_EPOCH if present
DATE                      := $(shell date -u -d "@$(SOURCE_DATE_EPOCH)" '+%FT%T%z' 2>/dev/null || date -u '+%FT%T%z')
BUILDFLAGS                := -ldflags="-s -w -X main.version=$(GOPASS_VERSION) -X main.commit=$(GOPASS_REVISION) -X main.date=$(DATE) -extldflags '-static'" -gcflags="-trimpath=$(GOPATH)" -asmflags="-trimpath=$(GOPATH)"
TESTFLAGS                 ?=
PWD                       := $(shell pwd)
PREFIX                    ?= $(GOPATH)
BINDIR                    ?= $(PREFIX)/bin
GO                        := CGO_ENABLED=0 go
GOOS                      ?= $(shell go version | cut -d' ' -f4 | cut -d'/' -f1)
GOARCH                    ?= $(shell go version | cut -d' ' -f4 | cut -d'/' -f2)
TAGS                      ?= netgo

all: sysinfo crosscompile build install test codequality completion

.PHONY: clean build man

define ok
	@tput setaf 6 2>/dev/null || echo -n ""
	@echo " [OK]"
	@tput sgr0 2>/dev/null || echo -n ""
endef

sysinfo:
	@echo ">> SYSTEM INFORMATION"
	@echo -n "     PLATFORM: $(shell uname -a)"
	@$(call ok)
	@echo -n "     PWD:    : $(shell pwd)"
	@$(call ok)
	@echo -n "     GO      : $(shell go version)"
	@$(call ok)
	@echo -n "     BUILDFLAGS: $(BUILDFLAGS)"
	@$(call ok)
	@echo -n "     GIT     : $(shell git version)"
	@$(call ok)
	@echo -n "     GPG1    : $(shell gpg --version | head -1)"
	@$(call ok)
	@echo -n "     GPG2    : $(shell gpg2 --version | head -1)"
	@$(call ok)

clean:
	@echo -n ">> CLEAN"
	@$(GO) clean -i ./...
	@rm -f ./coverage-all.html
	@rm -f ./coverage-all.out
	@rm -f ./coverage.out
	@find . -type f -name "coverage.out" -delete
	@rm -f gopass_*.deb
	@rm -f gopass-*.pkg.tar.xz
	@rm -f gopass-*.rpm
	@rm -f gopass-*.tar.bz2
	@rm -f gopass-*.tar.gz
	@rm -f gopass-*-*
	@rm -f tests/tests
	@$(call ok)

build:
	@echo -n ">> BUILD, version = $(GOPASS_VERSION)/$(GOPASS_REVISION), output = $(GOPASS_OUTPUT)"
	@$(GO) build -o $(GOPASS_OUTPUT) $(BUILDFLAGS)
	@$(call ok)

install: build
	@echo -n ">> INSTALL, version = $(GOPASS_VERSION)"
	@install -m 0755 -d $(DESTDIR)$(BINDIR)
	@install -m 0755 $(GOPASS_OUTPUT) $(DESTDIR)$(BINDIR)/gopass
	@$(call ok)

fulltest: build
	@echo ">> TEST, \"full-mode\": race detector on"
	@echo "mode: atomic" > coverage-all.out
	@$(foreach pkg, $(PKGS),\
	    echo -n "     ";\
		go test -run '(Test|Example)' $(BUILDFLAGS) $(TESTFLAGS) -race -coverprofile=coverage.out -covermode=atomic $(pkg) || exit 1;\
		tail -n +2 coverage.out >> coverage-all.out;)
	@$(GO) tool cover -html=coverage-all.out -o coverage-all.html

test: build
	@echo ">> TEST, \"fast-mode\": race detector off"
	@echo "mode: count" > coverage-all.out
	@$(foreach pkg, $(PKGS),\
	    echo -n "     ";\
		$(GO) test  -run '(Test|Example)' $(BUILDFLAGS) $(TESTFLAGS) -coverprofile=coverage.out -covermode=count $(pkg) || exit 1;\
		tail -n +2 coverage.out >> coverage-all.out;)
	@$(GO) tool cover -html=coverage-all.out -o coverage-all.html

test-integration: build
	cd tests && GOPASS_BINARY=$(PWD)/$(GOPASS_OUTPUT) GOPASS_TEST_DIR=$(PWD)/tests go test -v

crosscompile:
	@echo -n ">> CROSSCOMPILE linux/amd64"
	@GOOS=linux GOARCH=amd64 $(GO) build -o $(GOPASS_OUTPUT)-linux-amd64
	@$(call ok)
	@echo -n ">> CROSSCOMPILE darwin/amd64"
	@GOOS=darwin GOARCH=amd64 $(GO) build -o $(GOPASS_OUTPUT)-darwin-amd64
	@$(call ok)
	@echo -n ">> CROSSCOMPILE windows/amd64"
	@GOOS=windows GOARCH=amd64 $(GO) build -o $(GOPASS_OUTPUT)-windows-amd64
	@$(call ok)

completion: $(BASH_COMPLETION_OUTPUT) $(FISH_COMPLETION_OUTPUT) $(ZSH_COMPLETION_OUTPUT)

$(BASH_COMPLETION_OUTPUT): build
	@echo -n ">> BASH COMPLETION, output = $(BASH_COMPLETION_OUTPUT)"
	@./gopass completion bash > $(BASH_COMPLETION_OUTPUT)
	@$(call ok)

$(FISH_COMPLETION_OUTPUT): build
	@echo -n ">> FISH COMPLETION, output = $(FISH_COMPLETION_OUTPUT)"
	@./gopass completion fish > $(FISH_COMPLETION_OUTPUT)
	@$(call ok)
	
$(ZSH_COMPLETION_OUTPUT): build
	@echo -n ">> ZSH COMPLETION, output = $(ZSH_COMPLETION_OUTPUT)"
	@./gopass completion zsh > $(ZSH_COMPLETION_OUTPUT)
	@$(call ok)

codequality:
	@echo ">> CODE QUALITY"
	@echo -n "     FMT       "
	@$(foreach gofile, $(GOFILES_NOVENDOR),\
			out=$$(gofmt -s -l -d -e $(gofile) | tee /dev/stderr); if [ -n "$$out" ]; then exit 1; fi;)
	@$(call ok)

	@echo -n "     VET       "
	@$(GO) vet ./...
	@$(call ok)

	@echo -n "     CYCLO     "
	@which gocyclo > /dev/null; if [ $$? -ne 0 ]; then \
		$(GO) get -u github.com/fzipp/gocyclo; \
	fi
	@$(foreach gofile, $(GOFILES_NOVENDOR),\
			gocyclo -over 15 $(gofile) || exit 1;)
	@$(call ok)

	@echo -n "     LINT      "
	@which golint > /dev/null; if [ $$? -ne 0 ]; then \
		$(GO) get -u github.com/golang/lint/golint; \
	fi
	@$(foreach pkg, $(PKGS),\
			golint -set_exit_status $(pkg) || exit 1;)
	@$(call ok)

	@echo -n "     INEFF     "
	@which ineffassign > /dev/null; if [ $$? -ne 0 ]; then \
		$(GO) get -u github.com/gordonklaus/ineffassign; \
	fi
	@ineffassign . || exit 1
	@$(call ok)

	@echo -n "     SPELL     "
	@which misspell > /dev/null; if [ $$? -ne 0 ]; then \
		$(GO) get -u github.com/client9/misspell/cmd/misspell; \
	fi
	@$(foreach gofile, $(GOFILES_NOVENDOR),\
			misspell --error $(gofile) || exit 1;)
	@$(call ok)

	@echo -n "     MEGACHECK "
	@which megacheck > /dev/null; if [ $$? -ne 0  ]; then \
		$(GO) get -u honnef.co/go/tools/cmd/megacheck; \
	fi
	@megacheck $(PKGS) || exit 1
	@$(call ok)

	@echo -n "     ERRCHECK  "
	@which errcheck > /dev/null; if [ $$? -ne 0  ]; then \
		$(GO) get -u github.com/kisielk/errcheck; \
	fi
	@errcheck $(PKGS) || exit 1
	@$(call ok)

	@echo -n "     UNCONVERT "
	@which unconvert > /dev/null; if [ $$? -ne 0  ]; then \
		$(GO) get -u github.com/mdempsky/unconvert; \
	fi
	@unconvert -v $(PKGS) || exit 1
	@$(call ok)

fuzz-gpg:
	mkdir -p workdir/gpg-cli/corpus
	go-fuzz-build github.com/justwatchcom/gopass/backend/gpg/cli
	go-fuzz -bin=cli-fuzz.zip -workdir=workdir/gpg-cli

fuzz-jsonapi:
	mkdir -p workdir/jsonapi/corpus
	go-fuzz-build github.com/justwatchcom/gopass/utils/jsonapi
	go-fuzz -bin=jsonapi-fuzz.zip -workdir=workdir/jsonapi
