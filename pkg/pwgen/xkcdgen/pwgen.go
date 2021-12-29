package xkcdgen

import "github.com/martinhoefling/goxkcdpwgen/xkcdpwgen"

// Random returns a random passphrase combined from four words.
func Random() string {
	password, _ := RandomLength(4, "en")
	return password
}

// RandomLength returns a random passphrase combined from the desired number.
// of words. Words are drawn from lang.
func RandomLength(length int, lang string) (string, error) {
	return RandomLengthDelim(length, " ", lang)
}

// RandomLengthDelim returns a random passphrase combined from the desired number
// of words and the given delimiter. Words are drawn from lang.
func RandomLengthDelim(length int, delim, lang string) (string, error) {
	g := xkcdpwgen.NewGenerator()
	g.SetNumWords(length)
	g.SetDelimiter(delim)
	g.SetCapitalize(delim == "")
	if err := g.UseLangWordlist(lang); err != nil {
		return "", err
	}
	return string(g.GeneratePassword()), nil
}
