package sub

import (
	"testing"

	"github.com/justwatchcom/gopass/config"
)

func TestConfig(t *testing.T) {
	s := &Store{
		autoPush: true,
		path:     "/tmp/two",
	}
	cfg := s.Config()

	if !cfg.AutoPush {
		t.Errorf("AutoPush should be true")
	}
	if cfg.Path != "/tmp/two" {
		t.Errorf("Path should be /tmp/two")
	}
}

func TestUpdateConfig(t *testing.T) {
	s := &Store{}
	if err := s.UpdateConfig(&config.Config{
		AutoPush: true,
		Path:     "/tmp/foo",
	}); err != nil {
		t.Fatalf("Failed to update config: %s", err)
	}
	if s.Path() != "/tmp/foo" {
		t.Errorf("Wrong value for path")
	}
}
