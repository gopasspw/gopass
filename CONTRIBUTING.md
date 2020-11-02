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

* Branch from master and, if needed, rebase to the current master branch before submitting your pull request.
  If it doesn't merge cleanly with master you will be asked to rebase your changes.

* Commits should be as small as possible, while ensuring that each commit is correct independently.

* Add tests relevant to the fixed bug or new feature.

* Commit messages must contain both a [Developer Certificate of Origin](https://developercertificate.org/) / `Signed-off-by` line and a `RELEASE_NOTES=` entry, for example:

      One line description of commit

      More detailed description of commit, if needed.

      RELEASE_NOTES=Description for release notes, or n/a if trivial.

      Signed-off-by: Your Name <your@email.com>


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

### Releasing a new minor release

This subsection applies to major or minor releases, i.e. incrementing
X or Y in X.Y.Z. This is the regular release process.

We develop new features and fixes and feature branches which are frequently
merged into master in our own forks of the repository.

**Important: Do not push and feature branches to the main repo.**

```bash
# Change in to the repository
cd $GOPATH/src/github.com/gopasspw/gopass

# Update the Changelog
# TODO: Update CHANGELOG.md

# Update the version
echo v1.X.Y > VERSION
git commit -am'Tag v1.X.Y'

# Tag the new version
git tag -s v1.X.Y

# Generate shell completion files
make completion

# Do a release dry run to detect possible issues
goreleaser --skip-publish

# Push the tag to GitHub
git push origin v1.X.Y

# Build and push the release
GITHUB_TOKEN=ABC goreleaser

# Update the gopass website
# TODO: Update gopasspw.github.io
```

After these steps are complete please edit the auto-generated GitHub release
description and make it match the current CHANGELOG entry.

### Releasing a patch level release

This subsection applies to patch level releases, i.e. incrementing
Z in X.Y.Z.

If we need to release a patch release and can not base this upon the master
branch because there have been changes which should not be included in the patch
release (e.g. new features) we need to summon a new release branch from a past
release tag. Then we cherry-pick or port the required fixes to this branch and
create a release from it.

Tips for cherry-picking:
* Keep the changes small and self contained
* Squashed commits per feature help (one commit per fix/feature)
* Keep them in order

```bash
git checkout vX.Y.Z
git checkout -b release-X.Y
git cherry-pick ABC
git cherry-pick DEF
git cherry-pick FFF
make travis
# TODO: Update CHANGELOG.md and VERSION in ONE COMMIT
git commit -am'Tag X.Y.Z+1'
git tag -s vX.Y.Z+1
goreleaser --skip-publish
git push origin vX.Y.Z+1
GITHUB_TOKEN=ABC goreleaser
git push origin release-X.Y
```

After these steps are complete please edit the auto-generated GitHub release
description and make it match the current CHANGELOG entry.

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



