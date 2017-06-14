package root

import (
	"testing"

	"github.com/justwatchcom/gopass/config"
)

func TestConfig(t *testing.T) {
	s := &Store{
		autoPush:    true,
		clipTimeout: 10,
		path:        "/tmp/two",
	}
	cfg := s.Config()

	if !cfg.AutoPush {
		t.Errorf("AutoPush should be true")
	}
	if cfg.ClipTimeout != 10 {
		t.Errorf("ClipTimeout should be 10")
	}
	if cfg.Path != "/tmp/two" {
		t.Errorf("Path should be /tmp/two")
	}
}

func TestUpdateConfig(t *testing.T) {
	s := &Store{}
	if err := s.UpdateConfig(&config.Config{
		Path:        "/tmp/foo",
		NoConfirm:   true,
		ClipTimeout: 23,
	}); err != nil {
		t.Fatalf("Failed to update config: %s", err)
	}
	if s.Path() != "/tmp/foo" {
		t.Errorf("Wrong value for path")
	}
	if !s.NoConfirm() {
		t.Errorf("Wrong value for NoConfirm")
	}
	if s.ClipTimeout() != 23 {
		t.Errorf("Wrong clip timeout")
	}
}
