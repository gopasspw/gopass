[![Build Status](https://travis-ci.org/martinhoefling/goxkcdpwgen.svg?branch=master)](https://travis-ci.org/martinhoefling/goxkcdpwgen)

# goxkcdpwgen

xkcd style password generator library and cli tool

## Installation (cli tool)

### Compile

    go install -v github.com/martinhoefling/goxkcdpwgen 

### Package

no package yet :-)

### Run

All params

    $ goxkcdpwgen -h                                                 
    Usage of ./goxkcdpwgen:
    -c    Capitalize words
    -d string
            Delimiter to separate words (default " ")
    -n int
            Number of words to generate a password from (default 4)
    -s    Use eff_short instead of eff_long as wordlist

Sample execution

    $ goxkcdpwgen -c -d "" -n 5 
    VocalistDurableGauntletBluishReputable
    

## Usage as library

Install dependency

    go get github.com/martinhoefling/goxkcdpwgen
    
Use in code

    import (
        ...    
    	"github.com/martinhoefling/goxkcdpwgen/xkcdpwgen"
    )
    

    ...    
    	g := xkcdpwgen.NewGenerator()
    	g.SetNumWords(5)
    	g.SetCapitalize(true)
    	password := g.GeneratePasswordString()
    ...
