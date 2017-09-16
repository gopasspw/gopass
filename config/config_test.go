package config

import (
	"path/filepath"
	"testing"
)

func TestHomedir(t *testing.T) {
	if home := Homedir(); home == "" {
		t.Fatalf("Homedir must not be empty")
	}
}

func TestPwStoreDir(t *testing.T) {
	for in, out := range map[string]string{
		"":     filepath.Join(Homedir(), ".password-store"),
		"work": filepath.Join(Homedir(), ".password-store-work"),
		filepath.Join("foo", "bar"): filepath.Join(Homedir(), ".password-store-foo-bar"),
	} {
		got := PwStoreDir(in)
		if got != out {
			t.Errorf("Mismatch for %s: %s != %s", in, got, out)
		}
	}
}
