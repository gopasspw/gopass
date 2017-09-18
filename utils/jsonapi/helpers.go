package jsonapi

import (
	"regexp"
	"strings"

	"golang.org/x/net/publicsuffix"
)

func isPublicSuffix(host string) bool {
	suffix, _ := publicsuffix.PublicSuffix(host)
	return host == suffix
}

func regexSafeLower(str string) string {
	return regexp.QuoteMeta(strings.ToLower(str))
}
