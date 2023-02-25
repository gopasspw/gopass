package create

import (
	"context"
	"errors"
	"fmt"

	"github.com/gopasspw/gopass/internal/store"
	"github.com/gopasspw/gopass/pkg/debug"
	"gopkg.in/yaml.v3"
)

var defaultTemplates = []string{
	`---
priority: 0
name: "Website login"
prefix: "websites"
name_from:
  - "url"
  - "username"
welcome: "🧪 Creating Website login"
attributes:
  - name: "url"
    type: "hostname"
    prompt: "Website URL"
    min: 1
    max: 255
  - name: "username"
    type: "string"
    prompt: "Username"
    min: 1
  - name: "password"
    type: "password"
    prompt: "Password for the Website"
`,
	`---
priority: 1
name: "PIN Code (numerical)"
prefix: "pin"
name_from:
  - "authority"
  - "application"
welcome: "🧪 Creating numerical PIN"
attributes:
  - name: "authority"
    type: "string"
    prompt: "Authority"
    min: 1
  - name: "application"
    type: "string"
    prompt: "Entity"
    min: 1
  - name: "password"
    type: "password"
    prompt: "Pin"
    charset: "0123456789"
    min: 1
    max: 64
  - name: "comment"
    type: "string"
`,
}

type storageSetter interface {
	Set(context.Context, string, []byte) error
	Add(context.Context, ...string) error
	Commit(context.Context, string) error
}

func (w *Wizard) writeTemplates(ctx context.Context, s storageSetter) error {
	for _, y := range defaultTemplates {
		by := []byte(y)
		tpl := Template{}
		if err := yaml.Unmarshal(by, &tpl); err != nil {
			return fmt.Errorf("invalid default template. Please report a bug! %w", err)
		}

		path := fmt.Sprintf("%s%d-%s.yml", tplPath, tpl.Priority, tpl.Prefix)
		if err := s.Set(ctx, path, by); err != nil {
			if errors.Is(err, store.ErrMeaninglessWrite) {
				debug.Log("got unexpected error for %s (ignoring): %s", path, err)

				continue
			}

			return fmt.Errorf("failed to write default template %s: %w", path, err)
		}

		if err := s.Add(ctx, path); err != nil && !errors.Is(err, store.ErrGitNotInit) {
			return fmt.Errorf("failed to stage changes %s: %w", path, err)
		}

		debug.Log("wrote default template to %s", path)
	}

	if err := s.Commit(ctx, "Added default wizard templates"); err != nil && !errors.Is(err, store.ErrGitNotInit) {
		return fmt.Errorf("failed to commit changes: %w", err)
	}

	return nil
}
