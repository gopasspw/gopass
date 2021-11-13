# gpg crypto backend

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

This backend is the single most annoying source of maintenance workload in this project.
We try to keep this backend working as good as possible but there are a lot of reasons
why we'd prefer eventually move beyond GPG.

### GPG Critism

This section is a growing list of references why GPG is bad and why you should avoid it.
That might sound like an unusual thing to say for the authors of a tool whose main use case
relies on GPG but whenever we tried to move beyond GPG we got a lot of backlash. So I guess
first we need to try to make use understand why you shouldn't hold on to GPG and by then we'll
try to have a replacement ready for you.

* [What's the matter with PGP](https://blog.cryptographyengineering.com/2014/08/13/whats-matter-with-pgp/)
* [The PGP Problem](https://latacora.micro.blog/2019/07/16/the-pgp-problem.html)
* [I'm giving up on PGP](https://blog.filippo.io/giving-up-on-long-term-pgp/)
* [GPG and Me](https://moxie.org/2015/02/24/gpg-and-me.html)
