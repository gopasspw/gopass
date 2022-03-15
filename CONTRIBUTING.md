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

See [docs/releases.md](docs/releases.md).

