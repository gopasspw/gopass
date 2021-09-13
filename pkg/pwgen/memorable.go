package pwgen

import "strings"

// GenerateMemorablePassword will generate a memorable password
// with a minimum length
func GenerateMemorablePassword(minLength int, symbols bool, capitals bool) string {
	var sb strings.Builder
	var upper = false
	for sb.Len() < minLength {
		// when requesting uppercase, we randomly uppercase words
		if capitals && randomInteger(2) == 0 {
			sb.WriteString(strings.Title(randomWord()))
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
		var str = sb.String()
		return strings.Title(string(str[0])) + str[1:]
	}
	return sb.String()
}

func randomWord() string {
	return wordlist[randomInteger(len(wordlist))]
}
