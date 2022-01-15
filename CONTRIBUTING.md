# Contributing

`gopass` uses GitHub to manage reviews of pull requests.

* If you are a new contributor see: [Steps to Contribute](#steps-to-contribute)

* If you have a trivial fix or improvement, go ahead and create a pull request.

* If you plan to do something more involved, first raise an issue to discuss
  your idea. This will avoid unnecessary work.

* Relevant coding style guidelines are  the [Go Code Review Comments](https://code.google.com/p/go-wiki/wiki/CodeReviewComments)
  and the _Formatting and style_ section of Peter Bourgon's [Go: Best Practices for Production Environments](http://peter.bourgon.org/go-in-production/#formatting-and-style).

## Steps to Contribute

Should you wish to work on an issue, please claim it first by commenting on the GitHub issue you want to work on it.
This will prevent duplicated efforts from contributors.

Please check the [`help-wanted`](https://github.com/gopasspw/gopass/issues?q=is%3Aissue+is%3Aopen+label%3A%22help+wanted%22) label to find issues that need help.
If you have questions about one of the issues please comment on them and one of the maintainers
will try to clarify it.

## Pull Request Checklist

* Use that [latest stable Go release](https://golang.org/dl/)

**Note: This repository is already using features from - as of today unreleased - Go 1.18.
We expect this to be an exception that will resolve itself once Go 1.18 is released.**

* Branch from master and, if needed, rebase to the current master branch before submitting your pull request.
  If it doesn't merge cleanly with master you will be asked to rebase your changes.

* Commits should be as small as possible, while ensuring that each commit is correct independently.

* Add tests relevant to the fixed bug or new feature.

* Commit messages must contain both a [Developer Certificate of Origin](https://developercertificate.org/) / `Signed-off-by` line and a `RELEASE_NOTES=` entry, for example:

      One line description of commit

      More detailed description of commit, if needed.

      RELEASE_NOTES=[TAG] Description for release notes.

      Signed-off-by: Your Name <your@email.com>

  Valid `[TAG]`s are `[BREAKING]`, `[BUGFIX]`, `[CLEANUP]`, `[DEPRECATION]`,
  `[DOCUMENTATION]`, `[ENHANCEMENT]`, `[FEATURE]`, `[TESTING]`, and `[UX]`.
  Trivial changes should have no tag and the description `n/a`, i.e.
  `RELEASE_NOTES=n/a`.

## Building & Testing

* Build via `go build` to create the binary file `./gopass`.
* Run unit tests with: `make test`
* Run meta tests with: `make codequality`
* Run integration tests `make test-integration`

If any of the above don't work check out the [troubleshooting section](#troubleshooting-build).

## Releasing

This section is a reference for contributors with write access to the gopass
repository.

### Preparation

gopass release should work with the latest upstream version of goreleaser.

```bash
go get -u github.com/goreleaser/goreleaser
cd $GOPATH/src/github.com/goreleaser/goreleaser
go install
```

### Releasing a new release

This subsection applies to a new release in direct succession of the previous one, i.e. releasing what's in the master branch. If you need to cherry-pick
and base a release off of a previous one see the next subsection.

We develop new features and fixes and feature branches which are frequently
merged into master in our own forks of the repository.

**Important: Do not push feature branches to the main repo.**

We have some helpers and automation in place to help us release new versions.
A new release can be prepared by anyone (doesn't need to be a maintainer).
Releasing it involves sending a PR that needs to be reviwed by the maintainers.
Once approved the PR will be merged and a tag needs to be pushed to trigger
the release process automation (using GitHub actions).

```bash
$ go run helpers/release/main.go
# Follow the instructions to release a new minor version.
# If you want to skip a patch level or bump the minor version
# specify a version argument.
$ go run helpers/release/main.go v1.18.2
# If that confuses the changelog parser, you can specify a previous version
# as well.
$ go run helpers/release/main.go v1.18.2 v1.17.2
```

After the helper ran it will show you instructions how to push your release
branch and create a PR for review. Once it's merged a maintainer only needs
to tag it and push the tag to trigger the release process automation.

Afterwards a maintainer should run the post-release automation that will
perform some cleanup, create new GitHub milestones and send out PRs to
rolling release distributions.

### Releasing a cherry-pick release

This subsection applies to a new release that should be based on a previous
release that is not a direct ancestor of the master branch, i.e. because
breaking changes were introduced or other releases have happend in between.

This can still use our release automation but it will require some adjustments:

* Check out the previous release tag (we usually only publish a release branch if we need to): `git checkout v1.12.2`
* Create a new release preparation branch (the release automation will create the actual release branch later): `git checkout -b prep/v.1.12.3`
* Cherry-pick the changes from the previous release into the release preparation branch: `git cherry-pick -x HASH1 HASH2 ...`
  * Resolve any conflicts, make sure all tests pass and `git cherry-pick --continue`
* Trigger the release preparation, it should pick up any changelog entries from the cherry-picked commit messages: `PATCH_RELEASE=true go run helpers/release/main.go`
  * `PATCH_RELEASE=true` instructs it to not change the current branch to master
* Push the release branch printed at the end to the repository (or your fork) and open a PR.
  * IMPORANT: This PR will not be merged into master! We will just use it to create a tag and trigger the release automation.

## Troubleshooting

### Vendoring

This project use `dep` to manage it's dependencies. See this [gist](https://gist.github.com/subfuzion/12342599e26f5094e4e2d08e9d4ad50d) for a quick overview.

### Docker Approach

gopass ships a ready to use Dockerfile based on Alpine. It allows to run tests
and build gopass without having to setup a Go stack on the host.

```bash
cd $GOPATH/src/github.com/gopasspw/gopass
make docker-test
```

You can also run an interactive shell inside the container via:

```bash
docker run --rm -ti gopass sh
```

It is also possible mount a local directory into the container to copy files in
and out of it, but please pay attention to permissions.

```bash
docker run -it -v "$PWD":/go/src/github.com/gopasspw/gopass -w /go/src/github.com/gopasspw/gopass gopass sh
```

Please note that it is not recommended to actually *use* gopass inside Docker
as there are issues with random number generation in general and GnuPG.

### Setup of your local environment

- `go env` shows helpful info about the current env setup for go.
- See https://github.com/golang/go/wiki/GOPATH for more info on setting `$GOPATH` and `$GOROOT` env vars.

Quick Start:
- `mkdir -p $HOME/go/src`
- `export GOPATH=$HOME/go`
- `go get -u github.com/gopasspw/gopass`
- Set `$GOROOT` depending on your OS and Go installation method:
  - MacOS, Go installed via brew: `export GOROOT=/usr/local/opt/go/libexec/`
- Now you should be able to build from the gopass dir:
  - `cd $GOPATH/src/github.com/gopasspw/`
  - `go build -v`



