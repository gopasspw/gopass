<p align="center">
    <img src="docs/logo.png" height="250" alt="gopass Gopher by Vincent Leinweber, remixed from the Renée French original Gopher" title="gopass Gopher by Vincent Leinweber, remixed from the Renée French original Gopher" />
</p>

# gopass

[![Build Status](https://travis-ci.org/gopasspw/gopass.svg?branch=master)](https://travis-ci.org/gopasspw/gopass)
[![Go Report Card](https://goreportcard.com/badge/github.com/gopasspw/gopass)](https://goreportcard.com/report/github.com/gopasspw/gopass)
[![MIT License](https://img.shields.io/badge/license-MIT-blue.svg)](https://github.com/gopasspw/gopass/blob/master/LICENSE)
[![Github All Releases](https://img.shields.io/github/downloads/gopasspw/gopass/total.svg)](https://github.com/gopasspw/gopass/releases)
[![codecov](https://codecov.io/gh/gopasspw/gopass/branch/master/graph/badge.svg)](https://codecov.io/gh/gopasspw/gopass)
[![CII Best Practices](https://bestpractices.coreinfrastructure.org/projects/1899/badge)](https://bestpractices.coreinfrastructure.org/projects/1899)

The slightly more awesome Standard Unix Password Manager for Teams. Written in Go.

## Table of Contents

1. [Abstract](#abstract)
2. [Demo](#demo)
3. [Features](#features)
4. [Installation](#installation)
5. [Development](#development)
6. [Credit & License](#credit-&-license)
7. [Community](#community)
8. [Integrations](#integrations)
9. [Mobile apps](#mobile-apps)
10. [Contributing](#contributing)
11. [Acknowledgements](#acknowledgements)
12. [Further Documentation](#further-documentation)

## Abstract

> Password management should be simple and follow [Unix philosophy](http://en.wikipedia.org/wiki/Unix_philosophy). With `pass`, each secret lives inside of a `gpg` encrypted file whose filename is the title of the website or resource that requires the secret. These encrypted files may be organized into meaningful folder hierarchies, copied from computer to computer, and, in general, manipulated using standard command line file management utilities. - [passwordstore.org](https://www.passwordstore.org/)

*gopass* is a rewrite of the *pass* password manager in [Go](https://golang.org/) with the aim of making it cross-platform and [adding additional features](#features). Our target audience are professional developers and sysadmins (and especially teams of those) who are well versed with a command line interface. One explicit goal for this project is to make it more approachable to non-technical users. We go by the UNIX philosophy and try to do one thing and do it well, providing a stellar user experience and a sane, simple interface.

## Demo

[![asciicast](https://asciinema.org/a/172749.png)](https://asciinema.org/a/172749)

## Features

Please see [docs/features.md](https://github.com/gopasspw/gopass/blob/master/docs/features.md) for an extensive list of all features along with several usage examples.

| **Feature**                 | *pass* | *gopass* | **State** | **Description**                                                   |
| --------------------------- | ------ | -------- | --------- | ----------------------------------------------------------------- |
| Secure secret storage       | ✔      | ✔        | *stable*  | Securely storing secrets encrypted with GPG                       |
| Recipient management        | ❌      | ✔        | *beta*    | Easily manage multiple users of each store                        |
| Multiple stores             | ❌      | ✔        | *beta*    | Mount multiple stores in your root store, like file systems       |
| password quality assistance | ❌      | ✔        | *beta*    | Checks existing or new passwords for common flaws                 |
| password leak checker       | ❌      | ✔        | *integration*   | Perform **offline** checks against known leaked passwords         |
| PAGER support               | ❌      | ✔        | *stable*  | Automatically invoke a pager on long output                       |
| JSON API                    | ❌      | ✔        | *integration*   | Allow gopass to be used as a native extension for browser plugins |
| Automatic fuzzy search      | ❌      | ✔        | *stable*  | Automatically search for matching store entries if a literal entry was not found |
| gopass sync                 | ❌      | ✔        | *beta*    | Easy to use syncing of remote repos and GPG keys |
| Desktop Notifications       | ❌      | ✔        | *beta*    | Display desktop notifications and completing long running operations |
| Editing Recipients per Secret    | ❌      | ✔        | *beta*   | Select recipients per secret when encrypting |
| Extensions                  | ✔      | ✔        |           | Extend gopass with custom commands using our API |

## Installation

If you have [Go](https://golang.org/) 1.14 (or greater) installed:

```bash
go get github.com/gopasspw/gopass
```

Otherwise, please see [docs/setup.md](https://github.com/gopasspw/gopass/blob/master/docs/setup.md).


## Upgrade

To upgrade with Go installed, run:
```bash
go get -u github.com/gopasspw/gopass
```

Otherwise, use the setup docs mentioned in the installation section to reinstall the latest version.

## Development

This project uses [GitHub Flow](https://guides.github.com/introduction/flow/). In other words, create feature branches from master, open an PR against master, and rebase onto master if necessary.

We aim for compatibility with the [latest stable Go Release](https://golang.org/dl/) only.

While this project is maintained by volunteers in their free time we aim to triage issues weekly and release a new version at least every quarter.

## Credit & License

gopass was created by the nice folks from [JustWatch](https://www.justwatch.com/gopass) and licensed under the terms of the MIT license.

Maintainers of this repository:

* Dominik Schulz <dominik.schulz@gmail.com> [@dominikschulz](https://github.com/dominikschulz)
* Martin Hoefling [@martinhoefling](https://github.com/martinhoefling)

Please refer to the Git commit log for a complete list of contributors.

## Community

gopass is developed in the open. Here are some of the channels we use to communicate and contribute:

* IRC: #gopass on [irc.freenode.net](https://freenode.net) ([join via Riot](https://riot.im/app/#/room/#freenode_#gopass:matrix.org))
* Usage mailing list: [gopass-users](https://groups.google.com/forum/#!forum/gopass-users), for discussions around gopass usage and community support
* Issue tracker: Use the [GitHub issue tracker](https://github.com/gopasspw/gopass/issues) to file bugs and feature requests. If you need support, please send your questions to [gopass-users](https://groups.google.com/forum/#!forum/gopass-users) or ask on IRC rather than filing a GitHub issue.

## Integrations

- [gopassbridge](https://github.com/gopasspw/gopassbridge): Browser plugin for Firefox, Chrome and other Chromium based browsers
- [kubectl gopass](https://github.com/gopasspw/kubectl-gopass): Kubernetes / kubectl plugin to support reading and writing secrets directly from/to gopass.
- [gopass alfred](https://github.com/gopasspw/gopass-alfred): Alfred workflow to use gopass from the Alfred Mac launcher
- `gopass-git-credentials`: Integrate gopass as an git-credential helper
- `gopass-hibp`: haveibeenpwned.com leak checker
- `gopass-jsonapi`: native messaging for browser plugins, e.g. gopassbridge

## Mobile apps

- [Pass - Password Store](https://apps.apple.com/us/app/pass-password-store/id1205820573) - iOS, [source code](https://github.com/mssun/passforios), [supports only 1 repository now](https://github.com/mssun/passforios/issues/88)
- [Password Store](https://play.google.com/store/apps/details?id=dev.msfjarvis.aps) - Android

## Contributing

We welcome any contributions. Please see the [CONTRIBUTING.md](https://github.com/gopasspw/gopass/blob/master/CONTRIBUTING.md) file for instructions on how to submit changes. If your are planning on making more elaborate or controversial changes, please discuss them on the [gopass-developers mailing list](https://groups.google.com/forum/#!forum/gopass-developers) or on IRC before sending a pull request.

## Acknowledgements

gopass was initially started by Matthias Loibl and Dominik Schulz.
The majority of its development has been sponsored by [JustWatch](https://www.justwatch.com/).

## Further Documentation

* [Security, Known Limitations, and Caveats](https://github.com/gopasspw/gopass/blob/master/docs/security.md)
* [Configuration](https://github.com/gopasspw/gopass/blob/master/docs/config.md)
* [FAQ](https://github.com/gopasspw/gopass/blob/master/docs/faq.md)
* [JSON API](https://github.com/gopasspw/gopass/blob/master/docs/jsonapi.md)
* [Gopass as Summon provider](https://github.com/gopasspw/gopass/blob/master/docs/summon-provider.md)

## External Documentation
* [gopass cheat sheet](https://woile.github.io/gopass-cheat-sheet/) ([source on github](https://github.com/Woile/gopass-presentation))
* [gopass presentation](https://woile.github.io/gopass-presentation/) ([source on github](https://github.com/Woile/gopass-presentation))
