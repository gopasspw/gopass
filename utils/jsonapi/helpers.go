package jsonapi

import (
	"regexp"
	"strings"

	"golang.org/x/net/publicsuffix"
)

// isPublicSuffix returns true if this host is one users can or could directly
// register names
func isPublicSuffix(host string) bool {
	suffix, _ := publicsuffix.PublicSuffix(host)
	return host == suffix
}

func regexSafeLower(str string) string {
	return regexp.QuoteMeta(strings.ToLower(str))
}
