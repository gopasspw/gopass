//go:build linux

package secretservice

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestTemplatesMatchContrib ensures the embedded templates stay in sync with
// the reference files shipped in contrib/secret-service/.
func TestTemplatesMatchContrib(t *testing.T) {
	for _, tc := range []struct {
		file string
		tmpl string
	}{
		{"gopass-secret-service.service", systemdUnitTemplate},
		{"org.freedesktop.secrets.service", dbusActivationTemplate},
	} {
		got, err := os.ReadFile(filepath.Join("..", "..", "contrib", "secret-service", tc.file))
		if err != nil {
			t.Fatalf("read contrib file %s: %v", tc.file, err)
		}
		if string(got) != tc.tmpl {
			t.Fatalf("contrib/%s differs from embedded template", tc.file)
		}
	}
}

func TestRenderTemplateQuotesSpaces(t *testing.T) {
	out := renderTemplate(systemdUnitTemplate, "/opt/gopass bin/gopass")
	if !strings.Contains(out, `ExecStart="/opt/gopass bin/gopass" secret-service serve`) {
		t.Fatalf("binary path with spaces was not quoted:\n%s", out)
	}

	plain := renderTemplate(systemdUnitTemplate, "/usr/bin/gopass")
	if !strings.Contains(plain, "ExecStart=/usr/bin/gopass secret-service serve") {
		t.Fatalf("plain binary path should not be quoted:\n%s", plain)
	}
}

func TestInstallUninstall(t *testing.T) {
	dir := t.TempDir()
	paths := InstallPaths{
		SystemdUnit:    filepath.Join(dir, "systemd", "user", systemdUnitName),
		DBusActivation: filepath.Join(dir, "dbus-1", "services", dbusServiceName),
	}

	written, err := Install(paths, "/usr/bin/gopass")
	if err != nil {
		t.Fatalf("Install: %v", err)
	}
	if len(written) != 2 {
		t.Fatalf("wrote %d files, want 2", len(written))
	}

	body, err := os.ReadFile(paths.SystemdUnit)
	if err != nil {
		t.Fatalf("read unit: %v", err)
	}
	if strings.Contains(string(body), binaryPlaceholder) {
		t.Fatal("placeholder was not substituted")
	}
	if !strings.Contains(string(body), "/usr/bin/gopass secret-service serve") {
		t.Fatalf("unit does not reference the binary:\n%s", body)
	}

	// Uninstalling removes both files.
	removed, err := Uninstall(paths)
	if err != nil {
		t.Fatalf("Uninstall: %v", err)
	}
	if len(removed) != 2 {
		t.Fatalf("removed %d files, want 2", len(removed))
	}

	// A second uninstall is a no-op, not an error.
	removed, err = Uninstall(paths)
	if err != nil {
		t.Fatalf("second Uninstall: %v", err)
	}
	if len(removed) != 0 {
		t.Fatalf("second Uninstall removed %d files, want 0", len(removed))
	}
}
