// +build xc

package main

func init() {
	for _, cmd := range []string{
		".xc.decrypt",
		".xc.encrypt",
		".xc.export",
		".xc.export-private-key",
		".xc.generate",
		".xc.import",
		".xc.import-private-key",
		".xc.remove",
	} {
		commandsWithError[cmd] = struct{}{}
	}
}
