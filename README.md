<p align="center">
    <img src="logo.png" height="250" alt="gopass Gopher by Vincent Leinweber, remixed from the Renée French original Gopher" title="gopass Gopher by Vincent Leinweber, remixed from the Renée French original Gopher" />
</p>

# gopass

[![Build Status](https://travis-ci.org/justwatchcom/gopass.svg?branch=master)](https://travis-ci.org/justwatchcom/gopass)
[![Go Report Card](https://goreportcard.com/badge/github.com/justwatchcom/gopass)](https://goreportcard.com/report/github.com/justwatchcom/gopass)
[![Code Climate](https://codeclimate.com/github/justwatchcom/gopass/badges/gpa.svg)](https://codeclimate.com/github/justwatchcom/gopass)
[![MIT License](https://img.shields.io/badge/license-MIT-blue.svg)](https://github.com/justwatchcom/gopass/blob/master/LICENSE)
[![Github All Releases](https://img.shields.io/github/downloads/justwatchcom/gopass/total.svg)](https://github.com/justwatchcom/gopass/releases)
[![codecov](https://codecov.io/gh/justwatchcom/gopass/branch/master/graph/badge.svg)](https://codecov.io/gh/justwatchcom/gopass)
[![FOSSA Status](https://app.fossa.io/api/projects/git%2Bgithub.com%2Fjustwatchcom%2Fgopass.svg?type=shield)](https://app.fossa.io/projects/git%2Bgithub.com%2Fjustwatchcom%2Fgopass?ref=badge_shield)

The slightly more awesome Standard Unix Password Manager for Teams. Written in Go.

## Table of Contents

1. [Abstract](#abstract)
2. [Demo](#demo)
3. [Features](#features)
4. [Installation](#installation)
5. [Development](#development)
6. [Credit & License](#credit-&-license)
7. [Community](#community)
8. [Contributing](#contributing)
9. [Acknowledgements](#acknowledgements)
10. [Further Documentation](#further-documentation)

## Abstract

> Password management should be simple and follow [Unix philosophy](http://en.wikipedia.org/wiki/Unix_philosophy). With `pass`, each secret lives inside of a `gpg` encrypted file whose filename is the title of the website or resource that requires the secret. These encrypted files may be organized into meaningful folder hierarchies, copied from computer to computer, and, in general, manipulated using standard command line file management utilities. - [passwordstore.org](https://www.passwordstore.org/)

*gopass* is a rewrite of the *pass* password manager in [Go](https://golang.org/) with the aim of making it cross-platform and [adding additional features](#features). Our target audience are professional developers and sysadmins (and especially teams of those) who are well versed with a command line interface. One explicit goal for this project is to make it more approachable to non-technical users. We go by the UNIX philosophy and try to do one thing and do it well, providing a stellar user experience and a sane, simple interface.

## Demo

[![asciicast](https://asciinema.org/a/172749.png)](https://asciinema.org/a/172749)

## Features

Please see [docs/features.md](https://github.com/justwatchcom/gopass/blob/master/docs/features.md) for an extensive list of all features along with several usage examples.

| **Feature**                 | *pass* | *gopass* | **State** | **Description**                                                   |
| --------------------------- | ------ | -------- | --------- | ----------------------------------------------------------------- |
| Secure secret storage       | ✔      | ✔        | *stable*  | Securely storing secrets encrypted with GPG                       |
| Recipient management        | ❌      | ✔        | *beta*    | Easily manage multiple users of each store                        |
| Multiple stores             | ❌      | ✔        | *beta*    | Mount multiple stores in your root store, like file systems       |
| password quality assistance | ❌      | ✔        | *beta*    | Checks existing or new passwords for common flaws                 |
| Binary support              | ❌      | ✔        | *alpha*   | Special handling of binary files (automatic Base64 encoding)      |
| K/V and YAML support        | ❌      | ✔        | *alpha*   | Special handling for Key/Value and YAML content in secrets        |
| password leak checker       | ❌      | ✔        | *alpha*   | Perform **offline** checks against known leaked passwords         |
| PAGER support               | ❌      | ✔        | *stable*  | Automatically invoke a pager on long output                       |
| JSON API                    | ❌      | ✔        | *alpha*   | Allow gopass to be used as a native extension for browser plugins |
| Automatic fuzzy search      | ❌      | ✔        | *stable*  | Automatically search for matching store entries if a literal entry was not found |
| gopass sync                 | ❌      | ✔        | *beta*    | Easy to use syncing of remote repos and GPG keys |
| Desktop Notifications       | ❌      | ✔        | *beta*    | Display desktop notifications and completing long running operations |
| OTP support                 | (✔)    | ✔        | *stable*  | Generate HOTP/TOTP tokens based on the stored secret               |
| Multiple Crypto Backends    | ❌      | ✔        | *alpha*   | Extensible crypto backend support (GPG, NaCl)                      |
| Editing Recipients per Secret    | ❌      | ✔        | *beta*   | Select recipients per secret when encrypting |

## Installation

If you have [Go](https://golang.org/) 1.10 (or greater) installed:

```bash
go get github.com/justwatchcom/gopass
```

Otherwise, please see [docs/setup.md](https://github.com/justwatchcom/gopass/blob/master/docs/setup.md).

## Development

This project uses [GitHub Flow](https://guides.github.com/introduction/flow/). In other words, create feature branches from master, open an PR against master, and rebase onto master if necessary.

We aim for compatibility with the [latest stable Go Release](https://golang.org/dl/) only.

## Credit & License

gopass is maintained by the nice folks from [JustWatch](https://www.justwatch.com/gopass) and licensed under the terms of the MIT license.

Maintainers of this repository:

* Christoph Oelmüller <christoph.oelmueller@justwatch.com> [@zwopir](https://github.com/zwopir)
* Matthias Loibl <mail@matthiasloibl.com> [@metalmatze](https://github.com/metalmatze)
* Dominik Schulz <dominik.schulz@justwatch.com> [@dominikschulz](https://github.com/dominikschulz)

Please refer to the Git commit log for a complete list of contributors.


[![FOSSA Status](https://app.fossa.io/api/projects/git%2Bgithub.com%2Fjustwatchcom%2Fgopass.svg?type=large)](https://app.fossa.io/projects/git%2Bgithub.com%2Fjustwatchcom%2Fgopass?ref=badge_large)

## Community

gopass is developed in the open. Here are some of the channels we use to communicate and contribute:

* IRC: #gopass on [irc.freenode.net](https://freenode.net) ([join via Riot](https://riot.im/app/#/room/#freenode_#gopass:matrix.org))
* Usage mailing list: [gopass-users](https://groups.google.com/forum/#!forum/gopass-users), for discussions around gopass usage and community support
* Issue tracker: Use the [GitHub issue tracker](https://github.com/justwatchcom/gopass/issues) to file bugs and feature requests. If you need support, please send your questions to [gopass-user](https://groups.google.com/forum/#!forum/gopass-users) or ask on IRC rather than filing a GitHub issue.

## Contributing

We welcome any contributions. Please see the [CONTRIBUTING.md](https://github.com/justwatchcom/gopass/blob/master/CONTRIBUTING.md) file for instructions on how to submit changes. If your are planning on making more elaborate or controversial changes, please discuss them on the [gopass-developers mailing list](https://groups.google.com/forum/#!forum/gopass-developers) or on IRC before sending a pull request.

## Acknowledgements

gopass was initially started by Matthias Loibl and Dominik Schulz. The majority of its development has been sponsored by [JustWatch](https://www.justwatch.com/).

## Further Documentation

* [Security, Known Limitations, and Caveats](https://github.com/justwatchcom/gopass/blob/master/docs/security.md)
* [Configuration](https://github.com/justwatchcom/gopass/blob/master/docs/config.md)
* [FAQ](https://github.com/justwatchcom/gopass/blob/master/docs/faq.md)
* [JSON API](https://github.com/justwatchcom/gopass/blob/master/docs/jsonapi.md)
* [Gopass as Summon provider](https://github.com/justwatchcom/gopass/blob/master/docs/summon-provider.md)