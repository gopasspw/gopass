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


## Pull Request Checklist

* Use that [latest stable Go release](https://golang.org/dl/)

* Branch from master and, if needed, rebase to the current master branch before submitting your pull request.
  If it doesn't merge cleanly with master you will be asked to rebase your changes.

* Commits should be as small as possible, while ensuring that each commit is correct independently.

* Add tests relevant to the fixed bug or new feature.


## Building & Testing

* Build via `go build` to create the binary file `./gopass`.
* Run unit tests with: `make test`
* Run meta tests with: `make codequality`
* Run integration tests `make test-integration`

If any of the above don't work check out the [troubleshooting section](#troubleshooting-build).

## Troubleshooting

### Docker Approach

Building and testing commands can be run in a docker container.  
Change to the directory of your gopass checkout and run:
```
cd $GOPATH/src/github.com/justwatchcom/gopass
docker run --rm -v "$PWD":/go/src/github.com/justwatchcom/gopass -w /go/src/github.com/justwatchcom/gopass golang:stretch go build -v
```

Replace `go build -v` in the above command with `make test` or any other command you'd like to run inside the docker container.

You can also run an interactive shell inside the container via:
```
docker run -it -v "$PWD":/go/src/github.com/justwatchcom/gopass -w /go/src/github.com/justwatchcom/gopass golang:stretch bash
```

### Setup of your local environment

- `go env` shows helpful info about the current env setup for go.
- See https://github.com/golang/go/wiki/GOPATH for more info on setting `$GOPATH` and `$GOROOT` env vars.

Quick Start:
- `mkdir -p $HOME/go/src`
- `export GOPATH=$HOME/go`
- `go get -u github.com/justwatchcom/gopass`
- Set `$GOROOT` depending on your OS and Go installation method:
  - MacOS, Go installed via brew: `export GOROOT=/usr/local/opt/go/libexec/`
- Now you should be able to build from the gopass dir:
  - `cd $GOPATH/src/github.com/justwatchcom/`
  - `go build -v`



