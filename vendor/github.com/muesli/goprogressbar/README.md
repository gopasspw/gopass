goprogressbar
=============

Golang helper to print one or many progress bars on the console

## Installation

Make sure you have a working Go environment. Follow the [Go install instructions](http://golang.org/doc/install.html).

To install goprogressbar, simply run:

    go get github.com/muesli/goprogressbar

If you want to build it manually:

    cd $GOPATH/src/github.com/muesli/goprogressbar
    go get -u -v
    go build

## Example

```go
package main

import (
	"fmt"
	"strconv"
	"time"

	"github.com/muesli/goprogressbar"
)

func main() {
	mpb := goprogressbar.MultiProgressBar{}

	for i := 0; i < 10; i++ {
		pb := &goprogressbar.ProgressBar{
			Text:    "Progress " + strconv.FormatInt(int64(i+1), 10),
			Total:   100,
			Current: 0,
			Width:   60,
		}

		mpb.AddProgressBar(pb)
	}

	pb := &goprogressbar.ProgressBar{
		Text:    "Overall Progress",
		Total:   1000,
		Current: 0,
		Width:   60,
	}
	mpb.AddProgressBar(pb)

	// fill progress bars one after another
	for j := 0; j < 10; j++ {
		for i := 1; i <= 100; i++ {
			p := mpb.ProgressBars[j]
			p.Current = int64(i)
			p.RightAlignedText = fmt.Sprintf("%d of %d", i, p.Total)

			pb.Current++

			mpb.LazyPrint()
			time.Sleep(23 * time.Millisecond)
		}
	}

	fmt.Println()
}
```

## What it looks like
```
Progress 1                  100 of 100 [#################################################] 100.00%
Progress 2                  100 of 100 [#################################################] 100.00%
Progress 3                   89 of 100 [###########################################>-----]  89.00%
Progress 4                             [#>-----------------------------------------------]   0.00%
Progress 5                             [#>-----------------------------------------------]   0.00%
Progress 6                             [#>-----------------------------------------------]   0.00%
Progress 7                             [#>-----------------------------------------------]   0.00%
Progress 8                             [#>-----------------------------------------------]   0.00%
Progress 9                             [#>-----------------------------------------------]   0.00%
Progress 10                            [#>-----------------------------------------------]   0.00%
Overall Progress                       [#############>-----------------------------------]  28.90%
```

## Development

API docs can be found [here](http://godoc.org/github.com/muesli/goprogressbar).

[![Build Status](https://secure.travis-ci.org/muesli/goprogressbar.png)](http://travis-ci.org/muesli/goprogressbar)
[![Coverage Status](https://coveralls.io/repos/github/muesli/goprogressbar/badge.svg?branch=master)](https://coveralls.io/github/muesli/goprogressbar?branch=master)
[![Go ReportCard](http://goreportcard.com/badge/muesli/goprogressbar)](http://goreportcard.com/report/muesli/goprogressbar)
