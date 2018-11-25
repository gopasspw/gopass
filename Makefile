FIRST_GOPATH              := $(firstword $(subst :, ,$(GOPATH)))
PKGS                      := $(shell go list ./... | grep -v /tests | grep -v /xcpb | grep -v /openpgp)
GOFILES_NOVENDOR          := $(shell find . -name vendor -prune -o -type f -name '*.go' -not -name '*.pb.go' -print)
GOFILES_BUILD             := $(shell find . -type f -name '*.go' -not -name '*_test.go')
PROTOFILES                := $(shell find . -name vendor -prune -o -type f -name '*.proto' -print)
GOPASS_VERSION            ?= $(shell cat VERSION)
GOPASS_OUTPUT             ?= gopass
GOPASS_REVISION           := $(shell cat COMMIT 2>/dev/null || git rev-parse --short=8 HEAD)
BASH_COMPLETION_OUTPUT    := bash.completion
FISH_COMPLETION_OUTPUT    := fish.completion
ZSH_COMPLETION_OUTPUT     := zsh.completion
# Support reproducible builds by embedding date according to SOURCE_DATE_EPOCH if present
DATE                      := $(shell date -u -d "@$(SOURCE_DATE_EPOCH)" '+%FT%T%z' 2>/dev/null || date -u '+%FT%T%z')
BUILDFLAGS_NOPIE                := -ldflags="-s -w -X main.version=$(GOPASS_VERSION) -X main.commit=$(GOPASS_REVISION) -X main.date=$(DATE)" -gcflags="-trimpath=$(GOPATH)" -asmflags="-trimpath=$(GOPATH)"
BUILDFLAGS                := $(BUILDFLAGS_NOPIE) -buildmode=pie
TESTFLAGS                 ?=
PWD                       := $(shell pwd)
PREFIX                    ?= $(GOPATH)
BINDIR                    ?= $(PREFIX)/bin
GO                        := go
GOOS                      ?= $(shell go version | cut -d' ' -f4 | cut -d'/' -f1)
GOARCH                    ?= $(shell go version | cut -d' ' -f4 | cut -d'/' -f2)
TAGS                      ?= netgo

OK := $(shell tput setaf 6; echo ' [OK]'; tput sgr0;)

all: build completion
build: $(GOPASS_OUTPUT)
completion: $(BASH_COMPLETION_OUTPUT) $(FISH_COMPLETION_OUTPUT) $(ZSH_COMPLETION_OUTPUT)
travis: sysinfo crosscompile build install legal fulltest codequality completion manifests full

sysinfo:
	@echo ">> SYSTEM INFORMATION"
	@echo -n "     PLATFORM: $(shell uname -a)"
	@printf '%s\n' '$(OK)'
	@echo -n "     PWD:    : $(shell pwd)"
	@printf '%s\n' '$(OK)'
	@echo -n "     GO      : $(shell go version)"
	@printf '%s\n' '$(OK)'
	@echo -n "     BUILDFLAGS: $(BUILDFLAGS)"
	@printf '%s\n' '$(OK)'
	@echo -n "     GIT     : $(shell git version)"
	@printf '%s\n' '$(OK)'
	@echo -n "     GPG1    : $(shell gpg --version | head -1)"
	@printf '%s\n' '$(OK)'
	@echo -n "     GPG2    : $(shell gpg2 --version | head -1)"
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
	@rm -rf dist/*
	@printf '%s\n' '$(OK)'

$(GOPASS_OUTPUT): $(GOFILES_BUILD)
	@echo -n ">> BUILD, version = $(GOPASS_VERSION)/$(GOPASS_REVISION), output = $@"
	@$(GO) build -o $@ $(BUILDFLAGS)
	@printf '%s\n' '$(OK)'

install: all install-completion
	@echo -n ">> INSTALL, version = $(GOPASS_VERSION)"
	@install -m 0755 -d $(DESTDIR)$(BINDIR)
	@install -m 0755 $(GOPASS_OUTPUT) $(DESTDIR)$(BINDIR)/gopass
	@printf '%s\n' '$(OK)'

fulltest: $(GOPASS_OUTPUT)
	@echo ">> TEST, \"full-mode\": race detector off, build tags: xc, gogit, consul"
	@echo "mode: atomic" > coverage-all.out
	@$(foreach pkg, $(PKGS),\
	    echo -n "     ";\
		go test -run '(Test|Example)' $(BUILDFLAGS) $(TESTFLAGS) -coverprofile=coverage.out -covermode=atomic $(pkg) -tags 'xc gogit consul' || exit 1;\
		tail -n +2 coverage.out >> coverage-all.out;)
	@$(GO) tool cover -html=coverage-all.out -o coverage-all.html

racetest: $(GOPASS_OUTPUT)
	@echo ">> TEST, \"full-mode\": race detector on"
	@echo "mode: atomic" > coverage-all.out
	@$(foreach pkg, $(PKGS),\
	    echo -n "     ";\
		go test -run '(Test|Example)' $(BUILDFLAGS) $(TESTFLAGS) -race -coverprofile=coverage.out -covermode=atomic $(pkg) -tags 'xc gogit consul' || exit 1;\
		tail -n +2 coverage.out >> coverage-all.out;)
	@$(GO) tool cover -html=coverage-all.out -o coverage-all.html
test: $(GOPASS_OUTPUT)
	@echo ">> TEST, \"fast-mode\": race detector off"
	@echo "mode: count" > coverage-all.out
	@$(foreach pkg, $(PKGS),\
	    echo -n "     ";\
		$(GO) test  -run '(Test|Example)' $(BUILDFLAGS) $(TESTFLAGS) -coverprofile=coverage.out -covermode=count $(pkg) || exit 1;\
		tail -n +2 coverage.out >> coverage-all.out;)
	@$(GO) tool cover -html=coverage-all.out -o coverage-all.html

test-integration: $(GOPASS_OUTPUT)
	cd tests && GOPASS_BINARY=$(PWD)/$(GOPASS_OUTPUT) GOPASS_TEST_DIR=$(PWD)/tests go test -v

crosscompile:
	@echo -n ">> CROSSCOMPILE linux/amd64"
	@GOOS=linux GOARCH=amd64 $(GO) build -o $(GOPASS_OUTPUT)-linux-amd64
	@printf '%s\n' '$(OK)'
	@echo -n ">> CROSSCOMPILE darwin/amd64"
	@GOOS=darwin GOARCH=amd64 $(GO) build -o $(GOPASS_OUTPUT)-darwin-amd64
	@printf '%s\n' '$(OK)'
	@echo -n ">> CROSSCOMPILE windows/amd64"
	@GOOS=windows GOARCH=amd64 $(GO) build -o $(GOPASS_OUTPUT)-windows-amd64
	@printf '%s\n' '$(OK)'

full:
	@echo -n ">> COMPILE linux/amd64 xc gogit consul"
	$(GO) build -o $(GOPASS_OUTPUT)-linux-amd64-full -tags "xc gogit consul"

%.completion: $(GOPASS_OUTPUT)
	@printf ">> $* completion, output = $@"
	@./gopass completion $* > $@
	@printf "%s\n" "$(OK)"

install-completion: completion
	@install -d $(DESTDIR)$(PREFIX)/share/zsh/site-functions $(DESTDIR)$(PREFIX)/share/bash-completion/completions $(DESTDIR)$(PREFIX)/share/fish/vendor_completions.d
	@install -m 0755 $(ZSH_COMPLETION_OUTPUT) $(DESTDIR)$(PREFIX)/share/zsh/site-functions/_gopass
	@install -m 0755 $(BASH_COMPLETION_OUTPUT) $(DESTDIR)$(PREFIX)/share/bash-completion/completions/gopass
	@install -m 0755 $(FISH_COMPLETION_OUTPUT) $(DESTDIR)$(PREFIX)/share/fish/vendor_completions.d/gopass.fish
	@printf '%s\n' '$(OK)'

manifests: $(GOPASS_OUTPUT)
	@./gopass --yes jsonapi configure --path=. --manifest-path=manifest-chrome.json --browser=chrome --gopass-path=gopass --print=false
	@./gopass --yes jsonapi configure --path=. --manifest-path=manifest-chromium.json --browser=chromium --gopass-path=gopass --print=false
	@./gopass --yes jsonapi configure --path=. --manifest-path=manifest-firefox.json --browser=firefox --gopass-path=gopass --print=false

legal:
	@echo ">> LEGAL"
	@echo -n "   LICENSES   "
	@which licenses > /dev/null; if [ $$? -ne 0 ]; then \
		$(GO) get -u github.com/pmezard/licenses; \
	fi
	@GOOS=linux GOARCH=amd64 licenses ./... > NOTICE.new
	@diff NOTICE.txt NOTICE.new || exit 1
	@printf '%s\n' '$(OK)'

codequality:
	@echo ">> CODE QUALITY"

	@echo -n "     REVIVE    "
	@which revive > /dev/null; if [ $$? -ne 0 ]; then \
		$(GO) get -u github.com/mgechev/revive; \
	fi
	@revive -formatter friendly -exclude vendor/... ./...
	@printf '%s\n' '$(OK)'

	@echo -n "     FMT       "
	@$(foreach gofile, $(GOFILES_NOVENDOR),\
			out=$$(gofmt -s -l -d -e $(gofile) | tee /dev/stderr); if [ -n "$$out" ]; then exit 1; fi;)
	@printf '%s\n' '$(OK)'

	@echo -n "     CLANGFMT  "
	@$(foreach pbfile, $(PROTOFILES),\
			if [ $$(clang-format -output-replacements-xml $(pbfile) | wc -l) -gt 3  ]; then exit 1; fi;)
	@printf '%s\n' '$(OK)'

	@echo -n "     VET       "
	@$(GO) vet ./...
	@printf '%s\n' '$(OK)'

	@echo -n "     CYCLO     "
	@which gocyclo > /dev/null; if [ $$? -ne 0 ]; then \
		$(GO) get -u github.com/fzipp/gocyclo; \
	fi
	@$(foreach gofile, $(GOFILES_NOVENDOR),\
			gocyclo -over 22 $(gofile) || exit 1;)
	@printf '%s\n' '$(OK)'

	@echo -n "     LINT      "
	@which golint > /dev/null; if [ $$? -ne 0 ]; then \
		$(GO) get -u github.com/golang/lint/golint; \
	fi
	@$(foreach pkg, $(PKGS),\
			golint -set_exit_status $(pkg) || exit 1;)
	@printf '%s\n' '$(OK)'

	@echo -n "     INEFF     "
	@which ineffassign > /dev/null; if [ $$? -ne 0 ]; then \
		$(GO) get -u github.com/gordonklaus/ineffassign; \
	fi
	@ineffassign . || exit 1
	@printf '%s\n' '$(OK)'

	@echo -n "     SPELL     "
	@which misspell > /dev/null; if [ $$? -ne 0 ]; then \
		$(GO) get -u github.com/client9/misspell/cmd/misspell; \
	fi
	@$(foreach gofile, $(GOFILES_NOVENDOR),\
			misspell --error $(gofile) || exit 1;)
	@printf '%s\n' '$(OK)'

	@echo -n "     MEGACHECK "
	@which megacheck > /dev/null; if [ $$? -ne 0  ]; then \
		$(GO) get -u honnef.co/go/tools/cmd/megacheck; \
	fi
	@megacheck $(PKGS) || exit 1
	@printf '%s\n' '$(OK)'

	@echo -n "     ERRCHECK  "
	@which errcheck > /dev/null; if [ $$? -ne 0  ]; then \
		$(GO) get -u github.com/kisielk/errcheck; \
	fi
	@errcheck -exclude .errcheck.excl $(PKGS) || exit 1
	@printf '%s\n' '$(OK)'

	@echo -n "     INTERFACER"
	@which interfacer > /dev/null; if [ $$? -ne 0 ]; then \
		$(GO) get -u mvdan.cc/interfacer; \
	fi
	@interfacer $(PKGS)
	@printf '%s\n' '$(OK)'

	@echo -n "     UNCONVERT "
	@which unconvert > /dev/null; if [ $$? -ne 0  ]; then \
		$(GO) get -u github.com/mdempsky/unconvert; \
	fi
	@unconvert -v $(PKGS) || exit 1
	@printf '%s\n' '$(OK)'

fmt:
	@gofmt -s -l -w $(GOFILES_NOVENDOR)
	@clang-format -i $(PROTOFILES)

fuzz-gpg:
	mkdir -p workdir/gpg-cli/corpus
	go-fuzz-build github.com/gopasspw/gopass/backend/gpg/cli
	go-fuzz -bin=cli-fuzz.zip -workdir=workdir/gpg-cli

fuzz-jsonapi:
	mkdir -p workdir/jsonapi/corpus
	go-fuzz-build github.com/gopasspw/gopass/utils/jsonapi
	go-fuzz -bin=jsonapi-fuzz.zip -workdir=workdir/jsonapi

docker-test:
	docker build -t gopass:$(GOPASS_REVISION) .
	docker run --rm gopass:$(GOPASS_REVISION) make test

check-release-env:
ifndef GITHUB_TOKEN
	$(error GITHUB_TOKEN is undefined)
endif
ifndef BINTRAY_USER
	$(error BINTRAY_USER is undefined)
endif
ifndef BINTRAY_GPG_PASSPHRASE
	$(error BINTRAY_GPG_PASSPHRASE is undefined)
endif
ifndef BINTRAY_API_KEY
	$(error BINTRAY_API_KEY is undefined)
endif

release: goreleaser bintray

goreleaser: check-release-env travis clean
	@echo ">> RELEASE, goreleaser"
	@goreleaser

bintray: check-release-env
	@echo ">> RELEASE, deb packages"
	@$(eval AMD64DEB:=$(shell ls ./dist/gopass-*-amd64.deb | xargs -n1 basename))
	@curl -f -T ./dist/$(AMD64DEB) -H "X-GPG-PASSPHRASE:$(BINTRAY_GPG_PASSPHRASE)" -u$(BINTRAY_USER):$(BINTRAY_API_KEY) "https://api.bintray.com/content/gopasspw/gopass/gopass/v$(GOPASS_VERSION)/pool/main/g/gopass/$(AMD64DEB);deb_distribution=trusty,xenial,bionic,wheezy,jessie,buster,sid;deb_component=main;deb_architecture=amd64;publish=1"
	@echo ""

	@$(eval I386DEB:=$(shell ls ./dist/gopass-*-386.deb | xargs -n1 basename))
	@curl -f -T ./dist/$(I386DEB) -H "X-GPG-PASSPHRASE:$(BINTRAY_GPG_PASSPHRASE)" -u$(BINTRAY_USER):$(BINTRAY_API_KEY) "https://api.bintray.com/content/gopasspw/gopass/gopass/v$(GOPASS_VERSION)/pool/main/g/gopass/$(I386DEB);deb_distribution=trusty,xenial,bionic,wheezy,jessie,buster,sid;deb_component=main;deb_architecture=i386;publish=1"
	@echo ""

	@echo "   CALCULATE METADATA, deb repository"
	@curl -f -X POST -H "X-GPG-PASSPHRASE:$(BINTRAY_GPG_PASSPHRASE)" -u$(BINTRAY_USER):$(BINTRAY_API_KEY) https://api.bintray.com/calc_metadata/gopasspw/gopass
	@echo ""
	@echo ">> DONE"

.PHONY: clean build completion install sysinfo crosscompile test codequality release goreleaser debsign bintray
