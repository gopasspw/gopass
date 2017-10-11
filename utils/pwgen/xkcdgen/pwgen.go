package xkcdgen

import "github.com/martinhoefling/goxkcdpwgen/xkcdpwgen"

// Random returns a random passphrase combined from four words
func Random() string {
	return RandomLength(4)
}

// RandomLength returns a random passphrase combined from the desired number
// of words
func RandomLength(length int) string {
	return RandomLengthDelim(length, " ")
}

// RandomLengthDelim returns a random passphrase combined from the desired number
// of words and the given delimiter
func RandomLengthDelim(length int, delim string) string {
	g := xkcdpwgen.NewGenerator()
	g.SetNumWords(length)
	g.SetDelimiter(delim)
	g.SetCapitalize(delim == "")
	return string(g.GeneratePassword())
}
