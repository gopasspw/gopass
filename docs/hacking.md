# Hacking on gopass

Note: See [CONTRIBUTING.md](../CONTRIBUTING.md) for an overview.

This document provides an overview on how to develop on gopass.

## Development

This project uses [GitHub Flow](https://guides.github.com/introduction/flow/). In other words, create feature branches from master, open an PR against master, and rebase onto master if necessary.

We aim for compatibility with the [latest stable Go Release](https://golang.org/dl/) only.

While this project is maintained by volunteers in their free time we aim to triage issues weekly and release a new version at least every quarter.

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

The main extension model are small binaries that use the [gopass API](https://pkg.go.dev/github.com/gopasspw/gopass/pkg/gopass/api) package. This package provides a small and easy to use API that should work with any up to date gopass setup.

This API encapsulates the exact same implementation that the CLI uses in a more nicely packaged format that's easier to use.

Note: The API is operating directly on the password store. It does not involve network operations or connecting to a gopass instance.

The API does not support setting up a new password store (yet). Users will need have an existing password store
or use `gopass setup` to create a new one. The API will attempt to load an existing configuration or use it's built-in
defaults. Then it initializes an existing password store and provides a simple set of CRUD operations.

Our API has some [examples](../pkg/gopass/api/api_test.go) on how to use the API. The [gopass-hibp](https://github.com/gopasspw/gopass-hibp/blob/master/main.go) binary should provide a more complete example that can be used as a blueprint.

```go
import (
 "context"
 "fmt"

 "github.com/gopasspw/gopass/pkg/gopass/api"
 "github.com/gopasspw/gopass/pkg/gopass/secrets"
)

 ctx := context.Background()

 gp, err := api.New(ctx)
 if err != nil {
  panic(err)
 }

 // Listing secrets by their names (path within the store).
 ls, err := gp.List(ctx)
 if err != nil {
  panic(err)
 }

 for _, s := range ls {
  fmt.Printf("Secret: %s", s)
 }

 // Writing secrets to a specific location (path) in the store.
 sec := secrets.New()
 sec.SetPassword("foobar")
 if err := gp.Set(ctx, "my/new/secret", sec); err != nil {
  panic(err)
 }

 // Reading secrets by their name and revision from within the store.
 sec, err = gp.Get(ctx, "my/new/secret", "latest")
 if err != nil {
  panic(err)
 }
 fmt.Printf("content of %s: %s\n", "my/new/secret", string(sec.Bytes()))

 // Removing a secret by their name.
 if err := gp.Remove(ctx, "my/new/secret"); err != nil {
  panic(err)
 }

 // Cleaning up (waiting for background processing to complete).
 if err := gp.Close(ctx); err != nil {
  panic(err)
 }
```
