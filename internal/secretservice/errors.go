//go:build linux

package secretservice

import (
	"errors"

	"github.com/godbus/dbus/v5"
)

// D-Bus error names defined by the Secret Service specification, §11.
const (
	errNoSession     = "org.freedesktop.Secret.Error.NoSession"
	errNoSuchObject  = "org.freedesktop.Secret.Error.NoSuchObject"
	errIsLocked      = "org.freedesktop.Secret.Error.IsLocked"
	errAlreadyExists = "org.freedesktop.Secret.Error.AlreadyExists"
	errNotSupported  = "org.freedesktop.Secret.Error.NotSupported"
)

// errSessionNotFound marks a session lookup failure (missing, closed or owned
// by another client). Callers map it to the spec's NoSession error instead of
// the generic NotSupported.
var errSessionNotFound = errors.New("session not found")

// errItemNotFound marks a missing item. Store methods wrap it so callers can
// map a missing item to NoSuchObject while reporting other failures (e.g. a
// write error) as NotSupported.
var errItemNotFound = errors.New("item not found")

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

// errNoSessionError wraps err as NoSession.
func errNoSessionError(err error) *dbus.Error { return dbusError(errNoSession, err) }

// itemError maps a store item-update failure to the appropriate D-Bus error:
// NoSuchObject for a missing item, NotSupported for any other failure (e.g. a
// write error).
func itemError(err error) *dbus.Error {
	if errors.Is(err, errItemNotFound) {
		return errNotFound(err)
	}

	return errUnsupported(err)
}
