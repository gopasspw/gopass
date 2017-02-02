package action

import (
	"os"
	"path/filepath"
	"testing"
)

func TestPwStoreDir(t *testing.T) {
	home := ""
	if h := os.Getenv("HOME"); h != "" {
		home = h
	}
	for in, out := range map[string]string{
		"":        filepath.Join(home, ".password-store"),
		"work":    filepath.Join(home, ".password-store-work"),
		"foo/bar": filepath.Join(home, ".password-store-foo-bar"),
	} {
		got := pwStoreDir(in)
		if got != out {
			t.Errorf("Mismatch for %s: %s != %s", in, got, out)
		}
	}
}
