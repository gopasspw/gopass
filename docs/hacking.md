# Hacking on gopass

Note: See [CONTRIBUTING.md](../CONTRIBUTING.md) for an overview.

This document provides an overview on how to develop on gopass.

## Setting up an isolated development environment

### With GPG

`gopass` should fully respect `GOPASS_HOMEDIR` overriding all gopass internal paths.
However it will still use your normal GPG keyring and configuration. To override this
you will need to set `GNUPGHOME` as well and possibly generate a new keyring.

```bash
$ export GOPASS_DEBUG_LOG=/tmp/gp1.log
$ export GOPASS_HOMEDIR=/tmp/gp1
$ mkdir -p $GOPASS_HOMEDIR
$ export GNUPGHOME=$GOPASS_HOMEDIR/.gnupg
# Make sure that you're using the correct keyring.
$ gpg -K
gpg: directory '/tmp/gp1/.gnupg' created
gpg: keybox '/tmp/gp1/.gnupg/pubring.kbx' created
gpg: /tmp/gp1/.gnupg/trustdb.gpg: trustdb created
$ gpg --gen-key
$ go build && ./gopass setup --crypto gpg --storage gitfs
```

### With age

Using `age` is recommended for development since it's easier to set up. Setting
`GOPASS_HOMEDIR` should be sufficient to ensure an isolated environment.

```bash
$ export GOPASS_DEBUG_LOG=/tmp/gp1.log
$ export GOPASS_HOMEDIR=/tmp/gp1
$ mkdir -p $GOPASS_HOMEDIR
$ go build && ./gopass setup --crypto age --storage gitfs
```

## Extending gopass

The main extension model small binaries that use the [gopass API](https://pkg.go.dev/github.com/gopasspw/gopass/pkg/gopass/api) package. This package provides a small and easy to use API that should work with any up to date gopass setup.

We don't have extensive documentation for this, yet. But the [gopass-hibp](https://github.com/gopasspw/gopass-hibp/blob/master/main.go) binary should provide an easy example that can be used as a blueprint.
