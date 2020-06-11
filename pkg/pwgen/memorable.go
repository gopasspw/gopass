package pwgen

import "strings"

// GenerateMemorablePassword will generate a memorable password
// with a minimum length
func GenerateMemorablePassword(minLength int, symbols bool) string {
	var sb strings.Builder
	for sb.Len() < minLength {
		sb.WriteString(randomWord())
		sb.WriteByte(digits[randomInteger(len(digits))])
		if !symbols {
			continue
		}
		sb.WriteByte(syms[randomInteger(len(syms))])
	}
	return sb.String()
}

func randomWord() string {
	return wordlist[randomInteger(len(wordlist))]
}
