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

Your system dictionaries from `/usr/share/dict` will be indexed. If no dictionaries were found, crunchy only relies on
the regular sanity checks (`ErrEmpty`, `ErrTooShort`, `ErrTooFewChars` and `ErrTooSystematic`). On Ubuntu it is
recommended to install the wordlists distributed with `cracklib-runtime`, on macOS you can install `cracklib-words` from
brew. You could also install various other language dictionaries or wordlists, e.g. from skullsecurity.org.

crunchy uses the WagnerFischer algorithm to find mangled passwords in your dictionaries.

## Installation

Make sure you have a working Go environment (Go 1.2 or higher is required).
See the [install instructions](http://golang.org/doc/install.html).

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

    err := validator.Check("12345678")
    if err != nil {
        fmt.Printf("The password '12345678' is considered unsafe: %v\n", err)
    }

    err = validator.Check("p@ssw0rd")
    if dicterr, ok := err.(*crunchy.DictionaryError); ok {
        fmt.Printf("The password 'p@ssw0rd' is too similar to dictionary word '%s' (distance %d)\n",
            dicterr.Word, dicterr.Distance)
    }

    err = validator.Check("d1924ce3d0510b2b2b4604c99453e2e1")
    if err == nil {
        // Password is considered acceptable
        ...
    }
}
```

## Custom Options
```go
package main

import (
	"github.com/muesli/crunchy"
	"fmt"
)

func main() {
    validator := crunchy.NewValidatorWithOpts(crunchy.Options{
        // MinLength is the minimum length required for a valid password
        // (must be >= 1, default is 8)
        MinLength: 10,

        // MinDiff is the minimum amount of unique characters required for a valid password
        // (must be >= 1, default is 5)
        MinDiff: 8,

        // MinDist is the minimum WagnerFischer distance for mangled password dictionary lookups
        // (must be >= 0, default is 3)
        MinDist: 4,

        // Hashers will be used to find hashed passwords in dictionaries
        Hashers: []hash.Hash{md5.New(), sha1.New(), sha256.New(), sha512.New()},

        // DictionaryPath contains all the dictionaries that will be parsed
        // (default is /usr/share/dict)
        DictionaryPath: "/var/my/own/dicts",
    })
    ...
}
```

## Development

[![GoDoc](https://godoc.org/github.com/golang/gddo?status.svg)](https://godoc.org/github.com/muesli/crunchy)
[![Build Status](https://travis-ci.org/muesli/crunchy.svg?branch=master)](https://travis-ci.org/muesli/crunchy)
[![Coverage Status](https://coveralls.io/repos/github/muesli/crunchy/badge.svg?branch=master)](https://coveralls.io/github/muesli/crunchy?branch=master)
[![Go ReportCard](http://goreportcard.com/badge/muesli/crunchy)](http://goreportcard.com/report/muesli/crunchy)
