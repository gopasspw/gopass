package gpgid

import "github.com/justwatchcom/gopass/pkg/pwgen"

// Token is an private encryption token
type Token []byte

// String implement fmt.Stringer
func (t Token) String() string {
	return string(t)
}

// NewToken creates a new random token
func NewToken() Token {
	return Token(pwgen.GeneratePasswordCharset(128, pwgen.CharAlphaNum))
}

// TokenList is a list of tokens
type TokenList []Token

// Current returns the latest token
func (tl TokenList) Current() []byte {
	if len(tl) < 1 {
		panic("No token")
	}
	return tl[len(tl)-1]
}
