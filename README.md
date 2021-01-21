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
[![Gopass Slack](https://img.shields.io/badge/%23gopass-Slack-brightgreen)](https://docs.google.com/forms/d/e/1FAIpQLScxOPX_OLDaG5ak2E1kNdcFw9fJvPCr8xUaPGLyW8cyNUEnJw/viewform?usp=sf_link)

## Introduction

gopass is a password manager for the command line written in Go. It supports all major operating systems (Linux, MacOS, BSD) as well as Windows.

For detailed usage and installation instructions please check out our [documentation](docs/).

## Design Principles

Gopass is a versatile command line based password manager that is being developed with the following principles in mind:

- **Easy**: For technical users (i.e. those who are used to the command line) it should be easy to get started with gopass.
- **Secure**: Security is hard. We aim to make it as easy as possible while still providing a good level of protection against common adversaries. *Caution*: If your personal threat level is very high, we might not offer a good tool for you.
- **Extensible**: While Gopass includes a fair amount of useful features, we can't cover every use-case. To support more special use cases we want to provide a clean and simple API to integration gopass into your own binaries.

## Screenshot

![screenshot](docs/showcase.png)

## Features

Please see [docs/features.md](https://github.com/gopasspw/gopass/blob/master/docs/features.md) for an extensive list of all features along with several usage examples. Some examples are available in our
[example password store](https://github.com/gopasspw/password-store-example).

| **Feature**                 | **State**     | **Description**                                                   |
| --------------------------- | ------------- | ----------------------------------------------------------------- |
| Secure secret storage       | *stable*      | Securely storing encrypted secrets                                |
| Recipient management        | *beta*        | Easily manage multiple users of each store                        |
| Multiple stores             | *stable*      | Mount multiple stores in your root store, like file systems       |
| password quality assistance | *beta*        | Checks existing or new passwords for common flaws                 |
| password leak checker       | *integration* | Perform **offline** checks against known leaked passwords         |
| PAGER support               | *stable*      | Automatically invoke a pager on long output                       |
| JSON API                    | *integration* | Allow gopass to be used as a native extension for browser plugins |
| Automatic fuzzy search      | *stable*      | Automatically search for matching store entries if a literal entry was not found |
| gopass sync                 | *stable*      | Easy to use syncing of remote repos and GPG keys                  |
| Desktop Notifications       | *stable*      | Display desktop notifications and completing long running operations |
| REPL                        | *beta*        | Integrated Read-Eval-Print-Loop shell with autocompletion. |
| Extensions                  |               | Extend gopass with custom commands using our API                  |

## Installation

Please see [docs/setup.md](https://github.com/gopasspw/gopass/blob/master/docs/setup.md).

If you have [Go](https://golang.org/) 1.14 (or greater) installed:

```bash
GO111MODULE=on go get github.com/gopasspw/gopass
```

WARNING: Please prefer releases, unless you want to contribute to the
development of gopass. The master branch might not be very well tested and
can contain breaking changes without further notice.


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

gopass is licensed under the terms of the MIT license. You can find the complete text in `LICENSE`.

Please refer to the Git commit log for a complete list of contributors.

## Community

gopass is developed in the open. Here are some of the channels we use to communicate and contribute:

* Issue tracker: Use the [GitHub issue tracker](https://github.com/gopasspw/gopass/issues) to file bugs and feature requests.

## Integrations

- [gopassbridge](https://github.com/gopasspw/gopassbridge): Browser plugin for Firefox, Chrome and other Chromium based browsers
- [kubectl gopass](https://github.com/gopasspw/kubectl-gopass): Kubernetes / kubectl plugin to support reading and writing secrets directly from/to gopass.
- [gopass alfred](https://github.com/gopasspw/gopass-alfred): Alfred workflow to use gopass from the Alfred Mac launcher
- [git-credential-gopass](https://github.com/gopasspw/git-credential-gopass): Integrate gopass as an git-credential helper
- [gopass-hibp](https://github.com/gopasspw/gopass-hibp): haveibeenpwned.com leak checker
- [gopass-jsonapi](https://github.com/gopasspw/gopass-jsonapi): native messaging for browser plugins, e.g. gopassbridge
- [gopass-summon-prover](https://github.com/gopasspw/gopass-summon-provider): gopass as a summon provider
- [`terraform-provider-gopass`](https://github.com/camptocamp/terraform-provider-pass): a Terraform provider to interact with gopass
- [chezmoi](https://github.com/twpayne/chezmoi): dotfile manager with gopass support

## Mobile apps

- [Pass - Password Store](https://apps.apple.com/us/app/pass-password-store/id1205820573) - iOS, [source code](https://github.com/mssun/passforios), [supports only 1 repository now](https://github.com/mssun/passforios/issues/88)
- [Password Store](https://play.google.com/store/apps/details?id=dev.msfjarvis.aps) - Android

## Contributing

We welcome any contributions. Please see the [CONTRIBUTING.md](https://github.com/gopasspw/gopass/blob/master/CONTRIBUTING.md) file for instructions on how to submit changes.

## Further Documentation

* [Security, Known Limitations, and Caveats](https://github.com/gopasspw/gopass/blob/master/docs/security.md)
* [Configuration](https://github.com/gopasspw/gopass/blob/master/docs/config.md)
* [FAQ](https://github.com/gopasspw/gopass/blob/master/docs/faq.md)
* [JSON API](https://github.com/gopasspw/gopass/blob/master/docs/jsonapi.md)
* [Gopass as Summon provider](https://github.com/gopasspw/gopass/blob/master/docs/summon-provider.md)

## External Documentation
* [gopass cheat sheet](https://woile.github.io/gopass-cheat-sheet/) ([source on github](https://github.com/Woile/gopass-cheat-sheet))
* [gopass presentation](https://woile.github.io/gopass-presentation/) ([source on github](https://github.com/Woile/gopass-presentation))
