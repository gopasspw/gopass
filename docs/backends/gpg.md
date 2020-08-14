# gpg backend

The `gpgcli` backend is the default crypto backend based on the `gpg` CLI. It depends on the GPG installation to be working and having a properly initialized keyring.

## Getting started

WARNING: This backend suffers from myriads of different configuration options, a poor scripting interface and not pure-Go libarary bindings being available.

To start using the `gpgcli` backend initialize a new (sub) store with the `--crypto=gpgcli` flag:

```
gopass init --crypto gpgcli
gopass recipients add 0xDEADBEEF
```

## Features

* Compatible with other password store implementations
* Support for all GPG features, like smart-cards or hardware tokens

## Caveats

* Using long key sizes (e.g. 4096 bit or longer) can make many operations a lot slower
* Some GPG installations don't work well with concurrent operations

## Roadmap

This backend is the single most annoying source of maintenance workload in this project. Iff a viable replacement becomes available this backend might
be dropped entirely. Until then we try to keep it working as good as
possible.

