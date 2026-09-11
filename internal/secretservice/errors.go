//go:build linux

package secretservice

import "github.com/godbus/dbus/v5"

// D-Bus error names defined by the Secret Service specification, §11.
const (
	errNoSession     = "org.freedesktop.Secret.Error.NoSession"
	errNoSuchObject  = "org.freedesktop.Secret.Error.NoSuchObject"
	errIsLocked      = "org.freedesktop.Secret.Error.IsLocked"
	errAlreadyExists = "org.freedesktop.Secret.Error.AlreadyExists"
	errNotSupported  = "org.freedesktop.Secret.Error.NotSupported"
)

// dbusError builds a *dbus.Error with an optional human-readable message.
func dbusError(name string, err error) *dbus.Error {
	body := []any{}
	if err != nil {
		body = append(body, err.Error())
	}

	return dbus.NewError(name, body)
}

// errUnsupported wraps err as NotSupported.
func errUnsupported(err error) *dbus.Error { return dbusError(errNotSupported, err) }

// errNotFound wraps err as NoSuchObject.
func errNotFound(err error) *dbus.Error { return dbusError(errNoSuchObject, err) }

// errExists wraps err as AlreadyExists.
func errExists(err error) *dbus.Error { return dbusError(errAlreadyExists, err) }
