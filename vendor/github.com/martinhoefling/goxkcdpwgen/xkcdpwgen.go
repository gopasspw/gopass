package main

//go:generate go run _generator/main.go

import (
	"flag"
	"fmt"

	"github.com/martinhoefling/goxkcdpwgen/xkcdpwgen"
)

var wordcount = flag.Int("n", 4, "Number of words to generate a password from")
var delim = flag.String("d", " ", "Delimiter to separate words")
var lang = flag.String("l", "en", "Use non english language with custom list, currently only de = german is supported")
var effshort = flag.Bool("s", false, "Use eff_short instead of eff_long as wordlist")
var capitalize = flag.Bool("c", false, "Capitalize words")

func main() {
	flag.Parse()
	g := xkcdpwgen.NewGenerator()
	g.SetNumWords(*wordcount)
	g.SetDelimiter(*delim)
	g.SetCapitalize(*capitalize)
	if *effshort {
		g.UseWordlistEFFShort()
	}
	if *lang != "en" {
		if err := g.UseLangWordlist(*lang); err != nil {
			panic(err)
		}
	}
	fmt.Println(g.GeneratePasswordString())
}
