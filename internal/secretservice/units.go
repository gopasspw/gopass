//go:build linux

package secretservice

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/gopasspw/gopass/pkg/appdir"
)

// binaryPlaceholder is replaced with the absolute path of the running gopass
// binary when the unit/activation templates are installed.
const binaryPlaceholder = "@GOPASS@"

// Unit and activation file names.
const (
	systemdUnitName = "gopass-secret-service.service"
	dbusServiceName = "org.freedesktop.secrets.service"
)

// systemdUnitTemplate is the systemd user unit. It is kept byte-identical to
// contrib/secret-service/gopass-secret-service.service; a unit test enforces
// this.
const systemdUnitTemplate = `[Unit]
Description=gopass Secret Service D-Bus daemon
Documentation=https://github.com/gopasspw/gopass/blob/master/docs/commands/secret-service.md
After=graphical-session.target
PartOf=graphical-session.target

[Service]
Type=dbus
BusName=org.freedesktop.secrets
ExecStart=` + binaryPlaceholder + ` secret-service serve
Restart=on-failure

[Install]
WantedBy=graphical-session.target
`

// dbusActivationTemplate is the D-Bus session activation file. It is kept
// byte-identical to contrib/secret-service/org.freedesktop.secrets.service; a
// unit test enforces this.
const dbusActivationTemplate = `[D-BUS Service]
Name=org.freedesktop.secrets
Exec=` + binaryPlaceholder + ` secret-service serve
`

// InstallPaths are the XDG locations used by Install and Uninstall.
type InstallPaths struct {
	// SystemdUnit is the path of the systemd user unit.
	SystemdUnit string
	// DBusActivation is the path of the D-Bus session activation file.
	DBusActivation string
}

// DefaultInstallPaths returns the per-user paths used for installation. They
// honor XDG_CONFIG_HOME / XDG_DATA_HOME (and GOPASS_HOMEDIR) via pkg/appdir so
// tests can redirect them.
func DefaultInstallPaths() InstallPaths {
	configBase := filepath.Dir(appdir.UserConfig())
	dataBase := filepath.Dir(appdir.UserData())

	return InstallPaths{
		SystemdUnit:    filepath.Join(configBase, "systemd", "user", systemdUnitName),
		DBusActivation: filepath.Join(dataBase, "dbus-1", "services", dbusServiceName),
	}
}

// Install writes the systemd user unit and the D-Bus session activation file
// for the Secret Service daemon. It returns the paths it wrote.
func Install(paths InstallPaths, binary string) ([]string, error) {
	if binary == "" {
		exe, err := os.Executable()
		if err != nil {
			return nil, fmt.Errorf("determine gopass binary: %w", err)
		}
		binary = exe
	}

	written := make([]string, 0, 2)
	for _, f := range []struct {
		path    string
		content string
	}{
		{paths.SystemdUnit, renderTemplate(systemdUnitTemplate, binary)},
		{paths.DBusActivation, renderTemplate(dbusActivationTemplate, binary)},
	} {
		if f.path == "" {
			continue
		}
		if err := os.MkdirAll(filepath.Dir(f.path), 0o755); err != nil {
			return written, fmt.Errorf("create %s: %w", filepath.Dir(f.path), err)
		}
		if err := os.WriteFile(f.path, []byte(f.content), 0o644); err != nil {
			return written, fmt.Errorf("write %s: %w", f.path, err)
		}
		written = append(written, f.path)
	}

	return written, nil
}

// Uninstall removes the systemd user unit and the D-Bus session activation
// file. It returns the paths it removed. Missing files are not an error.
func Uninstall(paths InstallPaths) ([]string, error) {
	removed := make([]string, 0, 2)
	for _, p := range []string{paths.SystemdUnit, paths.DBusActivation} {
		if p == "" {
			continue
		}
		if err := os.Remove(p); err != nil {
			if os.IsNotExist(err) {
				continue
			}

			return removed, fmt.Errorf("remove %s: %w", p, err)
		}
		removed = append(removed, p)
	}

	return removed, nil
}

// renderTemplate substitutes the binary placeholder with the given path. The
// path is quoted so that it survives the systemd and D-Bus unit parsers even
// when it contains spaces.
func renderTemplate(tmpl, binary string) string {
	return strings.ReplaceAll(tmpl, binaryPlaceholder, quoteUnitArg(binary))
}

// quoteUnitArg quotes a single command argument if it needs quoting.
func quoteUnitArg(arg string) string {
	if arg != "" && !strings.ContainsAny(arg, " \t\"'\\") {
		return arg
	}

	escaped := strings.ReplaceAll(arg, `\`, `\\`)
	escaped = strings.ReplaceAll(escaped, `"`, `\"`)

	return `"` + escaped + `"`
}
