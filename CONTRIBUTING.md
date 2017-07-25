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

Please check the [`help-wanted`](https://github.com/justwatchcom/gopass/issues?q=is%3Aissue+is%3Aopen+label%3A%22help+wanted%22) label to find issues that need help.
If you have questions about one of the issues please comment on them and one of the maintainers
will try to clarify it.

For quickly compiling and testing your changes do:
```
go build
./gopass

# run unit and meta tests
make tests

# run integration tests
make test-integration
```

## Pull Request Checklist

* Use that [latest stable Go release](https://golang.org/dl/)

* Branch from master and, if needed, rebase to the current master branch before submitting your pull request.
  If it doesn't merge cleanly with master you will be asked to rebase your changes.

* Commits should be as small as possible, while ensuring that each commit is correct independently.

* Add tests relevant to the fixed bug or new feature.
