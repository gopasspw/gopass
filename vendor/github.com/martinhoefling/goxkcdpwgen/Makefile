GO := CGO_ENABLED=0 go
PACKAGES ?= $(shell go list ./... | grep -v /vendor/ | grep -v /tests)

.PHONY: all
all: tests

.PHONY: fmt
fmt:
	$(GO) fmt $(PACKAGES)

.PHONY: tests
tests: test vet lint errcheck megacheck

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

default: tests
