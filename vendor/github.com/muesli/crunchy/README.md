crunchy
=======

Finds common flaws in passwords. Like cracklib, but written in Go.

Detects:
 - Empty passwords: `ErrEmpty`
 - Too short passwords: `ErrTooShort`
 - Too few different characters, like "aabbccdd": `ErrTooFewChars`
 - Systematic passwords, like "abcdefgh" or "87654321": `ErrTooSystematic`
 - Passwords from a dictionary / wordlist: `ErrDictionary`
 - Mangled / reversed passwords, like "p@ssw0rd" or "drowssap": `ErrMangledDictionary`
 - Hashed dictionary words, like "5f4dcc3b5aa765d61d8327deb882cf99" (the md5sum of "password"): `ErrHashedDictionary`

Your system dictionaries from /usr/share/dict will be indexed. If no dictionaries were found, crunchy only relies on the
regular sanity checks (ErrEmpty, ErrTooShort, ErrTooFewChars and ErrTooSystematic). On Ubuntu it is recommended to install
the wordlists distributed with `cracklib-runtime`, on macOS you can install `cracklib-words` from brew. You could also
install various other language dictionaries or wordlists, e.g. from skullsecurity.org.

crunchy uses the WagnerFischer algorithm to find mangled passwords in your dictionaries.

## Installation

Make sure you have a working Go environment. See the [install instructions](http://golang.org/doc/install.html).

To install crunchy, simply run:

    go get github.com/muesli/crunchy

To compile it from source:

    cd $GOPATH/src/github.com/muesli/crunchy
    go get -u -v
    go build && go test -v

## Example
```go
package main

import (
	"github.com/muesli/crunchy"
	"fmt"
)

func main() {
    validator := crunchy.NewValidator()
    // there's also crunchy.NewValidatorWithOpts()

    err := validator.Check("12345678")
    if err != nil {
        fmt.Printf("The password '%s' is considered unsafe: %v\n", "12345678", err)
    }

    err = validator.Check("d1924ce3d0510b2b2b4604c99453e2e1")
    if err == nil {
        // Password is considered acceptable
        ...
    }
}
```

## Development

API docs can be found [here](http://godoc.org/github.com/muesli/crunchy).

[![Build Status](https://secure.travis-ci.org/muesli/crunchy.png)](http://travis-ci.org/muesli/crunchy)
[![Coverage Status](https://coveralls.io/repos/github/muesli/crunchy/badge.svg?branch=master)](https://coveralls.io/github/muesli/crunchy?branch=master)
[![Go ReportCard](http://goreportcard.com/badge/muesli/crunchy)](http://goreportcard.com/report/muesli/crunchy)
