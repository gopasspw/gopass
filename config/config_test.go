package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestHomedir(t *testing.T) {
	if home := Homedir(); home == "" {
		t.Fatalf("Homedir must not be empty")
	}
}

func TestNewConfig(t *testing.T) {
	if err := os.Setenv("GOPASS_CONFIG", filepath.Join(os.TempDir(), ".gopass.yml")); err != nil {
		t.Fatalf("Failed to set GOPASS_CONFIG: %s", err)
	}

	cfg := New()
	if cfg.Root.AskForMore {
		t.Errorf("AskForMore should be false")
	}
}

func TestSetConfigValue(t *testing.T) {
	if err := os.Setenv("GOPASS_CONFIG", filepath.Join(os.TempDir(), ".gopass.yml")); err != nil {
		t.Fatalf("Failed to set GOPASS_CONFIG: %s", err)
	}

	cfg := New()
	if err := cfg.SetConfigValue("", "autosync", "false"); err != nil {
		t.Errorf("Error: %s", err)
	}
	if err := cfg.SetConfigValue("", "askformore", "true"); err != nil {
		t.Errorf("Error: %s", err)
	}
	if err := cfg.SetConfigValue("", "askformore", "yo"); err == nil {
		t.Errorf("Should fail")
	}
	if err := cfg.SetConfigValue("", "cliptimeout", "900"); err != nil {
		t.Errorf("Error: %s", err)
	}
	if err := cfg.SetConfigValue("", "path", "/tmp"); err != nil {
		t.Errorf("Error: %s", err)
	}
	cfg.Mounts["foo"] = &StoreConfig{}
	if err := cfg.SetConfigValue("foo", "autosync", "true"); err != nil {
		t.Errorf("Error: %s", err)
	}
	if err := cfg.SetConfigValue("foo", "askformore", "true"); err != nil {
		t.Errorf("Error: %s", err)
	}
	if err := cfg.SetConfigValue("foo", "askformore", "yo"); err == nil {
		t.Errorf("Should fail")
	}
	if err := cfg.SetConfigValue("foo", "cliptimeout", "900"); err != nil {
		t.Errorf("Error: %s", err)
	}
}
