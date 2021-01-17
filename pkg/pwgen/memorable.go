package pwgen

import "strings"

// GenerateMemorablePassword will generate a memorable password
// with a minimum length
func GenerateMemorablePassword(minLength int, symbols bool) string {
	var sb strings.Builder
	for sb.Len() < minLength {
		sb.WriteString(randomWord())
		sb.WriteByte(Digits[randomInteger(len(Digits))])
		if !symbols {
			continue
		}
		sb.WriteByte(Syms[randomInteger(len(Syms))])
	}
	return sb.String()
}

func randomWord() string {
	return wordlist[randomInteger(len(wordlist))]
}
