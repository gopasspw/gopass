//go:build linux

// Package secretservice implements the freedesktop.org Secret Service D-Bus
// API on top of a gopass password store.
//
// It lets desktop applications that speak org.freedesktop.secrets — Firefox,
// Chromium, VS Code, Electron apps, NetworkManager, secret-tool, … — store
// their secrets in the same GPG-encrypted, git-backed store as the gopass CLI.
//
// The package is Linux-only. It is pure Go: D-Bus comes from
// github.com/godbus/dbus/v5 and all crypto from the standard library plus
// golang.org/x/crypto. See docs/adr/A-11-secret-service.md.
package secretservice

import (
	"github.com/godbus/dbus/v5"
	"github.com/gopasspw/gopass/internal/secretservice/crypto"
)

// Transport encryption algorithms defined by the Secret Service specification,
// re-exported for callers that negotiate a session.
const (
	// AlgorithmPlain performs no transport encryption.
	AlgorithmPlain = crypto.AlgorithmPlain
	// AlgorithmDHAES is Diffie-Hellman with AES-128-CBC.
	AlgorithmDHAES = crypto.AlgorithmDHAES
)

// Well-known D-Bus names, interfaces and object paths.
const (
	// ServiceName is the well-known bus name of the Secret Service.
	ServiceName = "org.freedesktop.secrets"

	// ServiceIface and friends are the Secret Service interfaces.
	ServiceIface    = "org.freedesktop.Secret.Service"
	CollectionIface = "org.freedesktop.Secret.Collection"
	ItemIface       = "org.freedesktop.Secret.Item"
	SessionIface    = "org.freedesktop.Secret.Session"
	PromptIface     = "org.freedesktop.Secret.Prompt"

	// PropertiesIface is the standard D-Bus properties interface.
	PropertiesIface = "org.freedesktop.DBus.Properties"
	// IntrospectableIface is the standard introspection interface.
	IntrospectableIface = "org.freedesktop.DBus.Introspectable"

	// ServicePath is the object path of the service object.
	ServicePath = dbus.ObjectPath("/org/freedesktop/secrets")

	collectionPathPrefix = "/org/freedesktop/secrets/collection/"
	aliasPathPrefix      = "/org/freedesktop/secrets/aliases/"
	sessionPathPrefix    = "/org/freedesktop/secrets/session/"

	// NullPath is the "no object / no prompt" path defined by the spec.
	NullPath = dbus.ObjectPath("/")
)

// Secret is the wire format of a secret as defined by the Secret Service
// specification (the struct with signature "(oayays)").
type Secret struct {
	// Session is the session used for transport encryption.
	Session dbus.ObjectPath
	// Parameters is the IV (empty for the "plain" algorithm).
	Parameters []byte
	// Value is the encrypted (or plaintext) secret value.
	Value []byte
	// ContentType is the MIME type, e.g. "text/plain".
	ContentType string
}
