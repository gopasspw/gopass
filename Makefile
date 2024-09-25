FIRST_GOPATH              := $(firstword $(subst :, ,$(GOPATH)))
PKGS                      := $(shell go list ./... | grep -v /tests | grep -v /xcpb | grep -v /gpb)
GOFILES_NOVENDOR          := $(shell find . -name vendor -prune -o -type f -name '*.go' -not -name '*.pb.go' -print)
GOFILES_BUILD             := $(shell find . -type f -name '*.go' -not -name '*_test.go')
GOPASS_VERSION            ?= $(shell cat VERSION)
GOPASS_OUTPUT             ?= gopass
GOPASS_REVISION           := $(shell cat COMMIT 2>/dev/null || git rev-parse --short=8 HEAD)
BASH_COMPLETION_OUTPUT    := bash.completion
FISH_COMPLETION_OUTPUT    := fish.completion
ZSH_COMPLETION_OUTPUT     := zsh.completion
CLIPHELPERS               ?= ""
# Support reproducible builds by embedding date according to SOURCE_DATE_EPOCH if present
DATE                      := $(shell date -u -d "@$(SOURCE_DATE_EPOCH)" '+%FT%T%z' 2>/dev/null || date -u '+%FT%T%z')
BUILDFLAGS_NOPIE          := -tags=netgo -trimpath -ldflags="-s -w -X main.version=$(GOPASS_VERSION) -X main.commit=$(GOPASS_REVISION) -X main.date=$(DATE) $(CLIPHELPERS)" -gcflags="-trimpath=$(GOPATH)" -asmflags="-trimpath=$(GOPATH)"
BUILDFLAGS                ?= $(BUILDFLAGS_NOPIE) -buildmode=pie
TESTFLAGS                 ?=
PWD                       := $(shell pwd)
PREFIX                    ?= $(GOPATH)
BINDIR                    ?= $(PREFIX)/bin
GO                        ?= GO111MODULE=on CGO_ENABLED=0 go
GOOS                      ?= $(shell $(GO) version | cut -d' ' -f4 | cut -d'/' -f1)
GOARCH                    ?= $(shell $(GO) version | cut -d' ' -f4 | cut -d'/' -f2)
TAGS                      ?= netgo
export GO111MODULE=on

OK := $(shell tput setaf 6; echo ' [OK]'; tput sgr0;)

all: sysinfo build
build: $(GOPASS_OUTPUT)
completion: $(BASH_COMPLETION_OUTPUT) $(FISH_COMPLETION_OUTPUT) $(ZSH_COMPLETION_OUTPUT)
travis: sysinfo crosscompile build fulltest completion codequality
travis-osx: sysinfo build test completion
travis-windows: sysinfo build test-win completion

sysinfo:
	@echo ">> SYSTEM INFORMATION"
	@echo -n "     PLATFORM   : $(shell uname -a)"
	@printf '%s\n' '$(OK)'
	@echo -n "     PWD:       : $(shell pwd)"
	@printf '%s\n' '$(OK)'
	@echo -n "     GO         : $(shell $(GO) version)"
	@printf '%s\n' '$(OK)'
	@echo -n "     BUILDFLAGS : $(BUILDFLAGS)"
	@printf '%s\n' '$(OK)'
	@echo -n "     GIT        : $(shell git version)"
	@printf '%s\n' '$(OK)'
	@echo -n "     GPG        : $(shell which gpg) $(shell gpg --version | head -1)"
	@printf '%s\n' '$(OK)'
	@echo -n "     GPGAgent   : $(shell which gpg-agent) $(shell gpg-agent --version | head -1)"
	@printf '%s\n' '$(OK)'

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
	@rm -f *.test
	@rm -rf dist/*
	@printf '%s\n' '$(OK)'

$(GOPASS_OUTPUT): $(GOFILES_BUILD)
	@echo -n ">> BUILD, version = $(GOPASS_VERSION)/$(GOPASS_REVISION), output = $@"
	@$(GO) build -o $@ $(BUILDFLAGS)
	@printf '%s\n' '$(OK)'

install: all install-completion install-man
	@echo -n ">> INSTALL, version = $(GOPASS_VERSION)"
	@install -m 0755 -d $(DESTDIR)$(BINDIR)
	@install -m 0755 $(GOPASS_OUTPUT) $(DESTDIR)$(BINDIR)/gopass
	@printf '%s\n' '$(OK)'

install-completion:
	@install -d $(DESTDIR)$(PREFIX)/share/zsh/site-functions $(DESTDIR)$(PREFIX)/share/bash-completion/completions $(DESTDIR)$(PREFIX)/share/fish/vendor_completions.d
	@install -m 0644 $(ZSH_COMPLETION_OUTPUT) $(DESTDIR)$(PREFIX)/share/zsh/site-functions/_gopass
	@install -m 0644 $(BASH_COMPLETION_OUTPUT) $(DESTDIR)$(PREFIX)/share/bash-completion/completions/gopass
	@install -m 0644 $(FISH_COMPLETION_OUTPUT) $(DESTDIR)$(PREFIX)/share/fish/vendor_completions.d/gopass.fish
	@printf '%s\n' '$(OK)'

install-man: gopass.1
	@install -d -m 0755 $(DESTDIR)$(PREFIX)/share/man/man1
	@install -m 0644 gopass.1 $(DESTDIR)$(PREFIX)/share/man/man1/gopass.1

fulltest: $(GOPASS_OUTPUT)
	@echo ">> TEST, \"full-mode\": race detector off"
	@echo "mode: atomic" > coverage-all.out
	@$(foreach pkg, $(PKGS),\
	    echo -n "     ";\
		$(GO) test -run '(Test|Example)' $(BUILDFLAGS) $(TESTFLAGS) -coverprofile=coverage.out -covermode=atomic $(pkg) || exit 1;\
		tail -n +2 coverage.out >> coverage-all.out;)
	@$(GO) tool cover -html=coverage-all.out -o coverage-all.html
	@which go-cover-treemap > /dev/null; if [ $$? -ne 0 ]; then \
		$(GO) install github.com/nikolaydubina/go-cover-treemap@latest; \
	fi
	@go-cover-treemap -coverprofile coverage-all.out > coverage-all.svg

test: $(GOPASS_OUTPUT)
	@echo ">> TEST, \"fast-mode\": race detector off"
	@$(foreach pkg, $(PKGS),\
	    echo -n "     ";\
		$(GO) test -test.short -run '(Test|Example)' $(BUILDFLAGS) $(TESTFLAGS) $(pkg) || exit 1;)

test-win: $(GOPASS_OUTPUT)
	@echo ">> TEST, \"fast-mode-win\": race detector off"
	@$(foreach pkg, $(PKGS),\
		$(GO) test -test.short -run '(Test|Example)' $(pkg) || exit 1;)

test-integration: $(GOPASS_OUTPUT)
	cd tests && GOPASS_BINARY=$(PWD)/$(GOPASS_OUTPUT) GOPASS_TEST_DIR=$(PWD)/tests $(GO) test -v $(TESTFLAGS)

crosscompile:
	@echo ">> CROSSCOMPILE"
	@which goreleaser > /dev/null; if [ $$? -ne 0 ]; then \
		$(GO) install github.com/goreleaser/goreleaser@latest; \
	fi
	@goreleaser build --snapshot

%.completion: $(GOPASS_OUTPUT)
	@printf ">> $* completion, output = $@"
	@./gopass completion $* > $@
	@printf "%s\n" "$(OK)"

codequality:
	@echo ">> CODE QUALITY"

	# Note: there are 2 different version of golangci-lint used inside the project.
	# https://github.com/gopasspw/gopass/blob/master/.github/workflows/build.yml#L65
	# https://github.com/gopasspw/gopass/blob/master/.github/workflows/golangci-lint.yml#L46
	# https://github.com/gopasspw/gopass/blob/master/Makefile#L136
	@echo -n "     GOLANGCI-LINT "
	@which golangci-lint > /dev/null; if [ $$? -ne 0 ]; then \
		$(GO) install github.com/golangci/golangci-lint/cmd/golangci-lint@latest; \
	fi
	@golangci-lint run --max-issues-per-linter 0 --max-same-issues 0 --sort-results || exit 1

	@printf '%s\n' '$(OK)'

	@echo -n "     LICENSE-LINT "
	@which license-lint > /dev/null; if [ $$? -ne 0 ]; then \
		$(GO) install istio.io/tools/cmd/license-lint@latest; \
	fi
	@license-lint --config .license-lint.yml >/dev/null || exit 1

	@printf '%s\n' '$(OK)'

gen:
	@$(GO) generate ./...

fmt:
	@gofumpt -s -l -w $(GOFILES_NOVENDOR)
	@gci write $(GOFILES_NOVENDOR)
	@$(GO) mod tidy

deps:
	@$(GO) build -v ./...

upgrade: gen fmt
	@$(GO) get -u ./...
	@$(GO) mod tidy

man:
	@$(GO) run helpers/man/main.go > gopass.1

msi:
	@$(GO) run helpers/msipkg/main.go

docker:
	docker build -t gopass:latest .

.PHONY: clean build completion install sysinfo crosscompile test codequality release goreleaser debsign man msi docker
