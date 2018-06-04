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

func convertMixedMapInterfaces(i interface{}) interface{} {
	switch x := i.(type) {
	case map[string]interface{}:
		stringMap := map[string]interface{}{}
		for k, v := range x {
			stringMap[k] = convertMixedMapInterfaces(v)
		}
		return stringMap
	case map[interface{}]interface{}:
		stringMap := map[string]interface{}{}
		for k, v := range x {
			stringMap[k.(string)] = convertMixedMapInterfaces(v)
		}
		return stringMap
	case []interface{}:
		for i, v := range x {
			x[i] = convertMixedMapInterfaces(v)
		}
	}
	return i
}
