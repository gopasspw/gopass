<p align="center">
    <img src="logo.png" height="500" alt="gopass Gopher by Vincent Leinweber, remixed from the Renée French original Gopher" title="gopass Gopher by Vincent Leinweber, remixed from the Renée French original Gopher" />
</p>

# gopass

[![Build Status](https://travis-ci.org/justwatchcom/gopass.svg?branch=master)](https://travis-ci.org/justwatchcom/gopass)
[![Go Report Card](https://goreportcard.com/badge/github.com/justwatchcom/gopass)](https://goreportcard.com/report/github.com/justwatchcom/gopass)
[![Code Climate](https://codeclimate.com/github/justwatchcom/gopass/badges/gpa.svg)](https://codeclimate.com/github/justwatchcom/gopass)

The slightly more awesome Standard Unix Password Manager for Teams. Written in Go.

> Password management should be simple and follow [Unix philosophy](http://en.wikipedia.org/wiki/Unix_philosophy). With `pass`, each secret lives inside of a `gpg` encrypted file whose filename is the title of the website or resource that requires the secret. These encrypted files may be organized into meaningful folder hierarchies, copied from computer to computer, and, in general, manipulated using standard command line file management utilities. - [passwordstore.org](https://www.passwordstore.org/)

Our target audience are professional developers and sysadmins (and especially teams of those) who are well versed with a
command line interface. One explicit goal for this project is to make it more approachable to semi- and non-technical users
in the long term as well. We go by the UNIX philosophy and try to do one thing and do it well - always providing stellar
user experience and sane, simple interfaces.

Warning: _gopass_ currently works on Linux & macOS. Please feel free to help with others.

## Demo

[![asciicast](https://asciinema.org/a/101688.png)](https://asciinema.org/a/101688)

## Features

Please see `docs/features.md` for an extensive list of all features along with
several usage examples.

| **Feature** | *State* | Description |
| ----------- | ------- | ----------- |
| **Secure secret storage** | *stable* | Securely storing secrets encrypted with GPG
| **Recipient management** | *beta* | Easily manage multiple users of each store
| **Multiple stores** | *beta* | Mount multiple stores in your root store, like filesystems
| **password quality assistance** | *beta* | Checks existing or new passwords for common flaws
| **Binary support** | *alpha* | Special handling of binary files (automatic Base64 encoding)
| **YAML support** | *alpha* | Special handling for YAML content in secrets
| **password leak checker** | *alpha* | Perform **offline** checks against known leaked passwords
| **PAGER support** | *stable* | Automatically invoke a pager on long output
| **JSON API** | *alpha* | Allow gopass to be used as a native extension for browser plugins

## Installation

If you have a Go development environment installed please build from source:

```bash
go get github.com/justwatchcom/gopass
```

Otherwise please see `docs/setup.md` or the [gopass website](https://www.justwatch.com/gopass/#install) for further instructions.

## Development

This project uses github-flow, i.e. create feature branches from master, open an PR against master
and rebase onto master if necessary.

We aim for compatibility with the [latest stable Go Release](https://golang.org/dl/) only.

## Security

Please see `docs/security.md`.

## Configuration

Please see `docs/config.md`.

## Credit & License

`gopass` is maintained by the nice folks from [JustWatch](https://www.justwatch.com/gopass)
and licensed under the terms of the MIT license.

Maintainers of this repository:

* Matthias Loibl <mail@matthiasloibl.com> [@metalmatze](https://github.com/metalmatze)
* Dominik Schulz <dominik.schulz@justwatch.com> [@dominikschulz](https://github.com/dominikschulz)

Please refer to the Git commit log for a complete list of contributors.

## Community

`gopass` is developed in the open. Here are some of the channels we use to communicate and contribute:

**IRC**: `#gopass` on [irc.freenode.net](https://freenode.net) ([join via Riot](https://riot.im/app/#/room/#freenode_#gopass:matrix.org))

**Usage mailing list:** [gopass-users](https://groups.google.com/forum/#!forum/gopass-users) - for discussions around gopass usage and community support

**Issue tracker:** Use the [GitHub issue tracker](https://github.com/justwatchcom/gopass/issues) to file bugs and feature requests. If you need support, please send your questions to [gopass-user](https://groups.google.com/forum/#!forum/gopass-users) or ask on IRC rather than filing a GitHub issue.

## Contributing

We welcome any contributions. Please see the `CONTRIBUTING.md` file for
instructions on how to submit changes. If your are planning on making
more elaborate or controversial changes, please discuss them on the
mailing list or on IRC before sending a pull request.

**Development mailing list:** [gopass-developers](https://groups.google.com/forum/#!forum/gopass-developers) - for discussions around gopass development

## Acknowledgements

`gopass` was initially started by Matthias Loibl and Dominik Schulz. The majority of its development has been sponsored by [JustWatch](https://www.justwatch.com/).
