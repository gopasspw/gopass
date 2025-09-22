package pwgen

import "strings"

// GenerateMemorablePassword will generate a memorable password
// with a minimum length.
// It will use a wordlist to generate the password.
// If symbols is true, it will add symbols to the password.
// If capitals is true, it will capitalize some words.
func GenerateMemorablePassword(minLength int, symbols bool, capitals bool) string {
	var sb strings.Builder

	upper := false

	for sb.Len() < minLength {
		// when requesting uppercase, we randomly uppercase words
		if capitals && randomInteger(2) == 0 {
			// We control the input so we can safely ignore the linter.
			sb.WriteString(strings.Title(randomWord())) //nolint:staticcheck

			upper = true
		} else {
			sb.WriteString(randomWord())
		}

		sb.WriteByte(Digits[randomInteger(len(Digits))])

		if !symbols {
			continue
		}

		sb.WriteByte(Syms[randomInteger(len(Syms))])
	}
	// If there isn't already a capitalized word, capitalize the first letter
	if capitals && !upper {
		str := sb.String()

		return strings.Title(string(str[0])) + str[1:] //nolint:staticcheck
	}

	return sb.String()
}

func randomWord() string {
	return wordlist[randomInteger(len(wordlist))]
}
