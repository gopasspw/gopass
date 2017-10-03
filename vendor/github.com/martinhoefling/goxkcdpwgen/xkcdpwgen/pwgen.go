package xkcdpwgen

import "strings"

// Generator encapsulates the password generator configuration
type Generator struct {
	wordlist   []string
	numwords   int
	delimiter  string
	capitalize bool
}

// NewGenerator returns a new password generator with default values set
func NewGenerator() *Generator {
	pwgen := Generator{}
	pwgen.wordlist = effLarge
	pwgen.numwords = 4
	pwgen.delimiter = " "
	pwgen.capitalize = false
	return &pwgen
}

// GeneratePassword creates a randomized password returned as byte slice
func (g *Generator) GeneratePassword() []byte {
	return []byte(g.GeneratePasswordString())
}

// GeneratePasswordString creates a randomized password returned as string
func (g *Generator) GeneratePasswordString() string {
	var words = make([]string, g.numwords)
	for i := 0; i < g.numwords; i++ {
		if g.capitalize {
			words[i] = strings.Title(randomWord(g.wordlist))
		} else {
			words[i] = randomWord(g.wordlist)
		}
	}
	return strings.Join(words, g.delimiter)
}

// SetNumWords sets the word count for the generator
func (g *Generator) SetNumWords(count int) {
	g.numwords = count
}

// SetDelimiter sets the delimiter string. Can also be set to an empty string.
func (g *Generator) SetDelimiter(delimiter string) {
	g.delimiter = delimiter
}

// UseWordlistEFFLarge sets the wordlist from which the passwords are generated to eff_large (https://www.eff.org/de/deeplinks/2016/07/new-wordlists-random-passphrases)
func (g *Generator) UseWordlistEFFLarge() {
	g.wordlist = effLarge
}

// UseWordlistEFFShort sets the wordlist from which the passwords are generated to eff_short (https://www.eff.org/de/deeplinks/2016/07/new-wordlists-random-passphrases)
func (g *Generator) UseWordlistEFFShort() {
	g.wordlist = effShort
}

// UseCustomWordlist sets the wordlist to the wl provided one
func (g *Generator) UseCustomWordlist(wl []string) {
	g.wordlist = wl
}

// SetCapitalize turns on/off capitalization of the first character
func (g *Generator) SetCapitalize(capitalize bool) {
	g.capitalize = capitalize
}

func randomWord(list []string) string {
	return list[randomInteger(len(list))]
}
