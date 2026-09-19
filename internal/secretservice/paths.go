//go:build linux

package secretservice

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"path"
	"strings"

	"github.com/godbus/dbus/v5"
)

// mapper translates between D-Bus object paths and gopass store paths.
type mapper struct {
	prefix string
}

// CollectionPath returns the gopass path of a collection.
func (m mapper) CollectionPath(name string) string {
	return path.Join(m.prefix, name)
}

// ItemPath returns the gopass path of an item.
func (m mapper) ItemPath(collection, id string) string {
	return path.Join(m.prefix, collection, id)
}

// MetaPath returns the gopass path of a collection's metadata entry.
func (m mapper) MetaPath(collection string) string {
	return path.Join(m.prefix, collection, "_meta")
}

// AliasesPath returns the gopass path of the alias map.
func (m mapper) AliasesPath() string {
	return path.Join(m.prefix, "_aliases")
}

// SanitizeName turns a caller-supplied collection name into a single safe
// gopass path segment, which is also a valid D-Bus object path element (that
// is, it matches [A-Za-z0-9_]). Every other character is replaced with "_".
// Because object path elements may not contain "/" this also prevents a name
// from escaping the store prefix.
func SanitizeName(name string) string {
	var b strings.Builder
	b.Grow(len(name))

	for _, r := range name {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9', r == '_':
			b.WriteRune(r)
		default:
			b.WriteByte('_')
		}
	}

	if b.Len() == 0 {
		return "_"
	}

	return b.String()
}

// ItemDBusID returns the D-Bus-safe object path element for a store item name.
//
// D-Bus object path elements must match [A-Za-z0-9_], but items may be created
// out of band by the gopass CLI with arbitrary names (e.g. "cli-inserted").
// Names that are already safe and do not start with "x" are used verbatim. All
// other names — including every name starting with "x" — are hex-encoded with
// an "x" prefix. Encoded and verbatim elements can therefore never collide,
// and ItemNameFromDBusID reverses the mapping exactly.
func ItemDBusID(name string) string {
	if isSafeItemID(name) {
		return name
	}

	return "x" + hex.EncodeToString([]byte(name))
}

// isSafeItemID reports whether name can be used verbatim as an object path
// element. Names starting with "x" are excluded so that the "x"-prefixed
// encoding stays unambiguous.
func isSafeItemID(name string) bool {
	return name != "" && name[0] != 'x' && isPathSafe(name)
}

// isPathSafe reports whether name is a valid D-Bus object path element, i.e.
// it is non-empty and matches [A-Za-z0-9_].
func isPathSafe(name string) bool {
	if name == "" {
		return false
	}

	for i := range len(name) {
		c := name[i]
		switch {
		case c >= 'a' && c <= 'z', c >= 'A' && c <= 'Z', c >= '0' && c <= '9', c == '_':
		default:
			return false
		}
	}

	return true
}

// ItemNameFromDBusID reverses ItemDBusID. Elements that are not a valid
// encoding are returned unchanged.
func ItemNameFromDBusID(id string) string {
	if !strings.HasPrefix(id, "x") {
		return id
	}

	raw := id[1:]
	if len(raw) == 0 || len(raw)%2 != 0 {
		return id
	}

	decoded, err := hex.DecodeString(raw)
	if err != nil {
		return id
	}

	return string(decoded)
}

// CollectionDBusPath returns the object path of a collection.
func CollectionDBusPath(name string) dbus.ObjectPath {
	return dbus.ObjectPath(collectionPathPrefix + name)
}

// AliasDBusPath returns the object path at which a collection alias is served.
//
// libsecret does not call ReadAlias to resolve a collection alias. Instead it
// derives /org/freedesktop/secrets/aliases/<alias> locally and treats that path
// as the collection object, so the Collection interface must be exported there
// as well as at the canonical collection path.
func AliasDBusPath(alias string) dbus.ObjectPath {
	return dbus.ObjectPath(aliasPathPrefix + SanitizeName(alias))
}

// IsAliasPath reports whether p is an alias object path and returns the alias
// name.
func IsAliasPath(p dbus.ObjectPath) (string, bool) {
	s := string(p)
	if !strings.HasPrefix(s, aliasPathPrefix) {
		return "", false
	}

	alias := strings.TrimPrefix(s, aliasPathPrefix)
	if alias == "" || strings.Contains(alias, "/") {
		return "", false
	}

	return alias, true
}

// ItemDBusPath returns the object path of an item. id is the store item name,
// which is mapped to a D-Bus-safe path element by ItemDBusID.
func ItemDBusPath(collection, id string) dbus.ObjectPath {
	return dbus.ObjectPath(collectionPathPrefix + collection + "/" + ItemDBusID(id))
}

// SessionDBusPath returns the object path of a session.
func SessionDBusPath(id string) dbus.ObjectPath {
	return dbus.ObjectPath(sessionPathPrefix + id)
}

// parseCollectionPath extracts the collection name from a collection object
// path.
func parseCollectionPath(p dbus.ObjectPath) (string, error) {
	s := string(p)
	if !strings.HasPrefix(s, collectionPathPrefix) {
		return "", fmt.Errorf("not a collection path: %s", p)
	}

	name := strings.TrimPrefix(s, collectionPathPrefix)
	if name == "" || strings.Contains(name, "/") {
		return "", fmt.Errorf("invalid collection path: %s", p)
	}

	return name, nil
}

// parseItemPath extracts the collection name and the D-Bus item ID from an
// item object path. Use ItemNameFromDBusID to recover the store item name.
func parseItemPath(p dbus.ObjectPath) (string, string, error) {
	s := string(p)
	if !strings.HasPrefix(s, collectionPathPrefix) {
		return "", "", fmt.Errorf("not an item path: %s", p)
	}

	rest := strings.TrimPrefix(s, collectionPathPrefix)
	parts := strings.SplitN(rest, "/", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("invalid item path: %s", p)
	}

	return parts[0], parts[1], nil
}

// parseSessionPath extracts the session ID from a session object path.
func parseSessionPath(p dbus.ObjectPath) (string, error) {
	s := string(p)
	if !strings.HasPrefix(s, sessionPathPrefix) {
		return "", fmt.Errorf("not a session path: %s", p)
	}

	id := strings.TrimPrefix(s, sessionPathPrefix)
	if id == "" || strings.Contains(id, "/") {
		return "", fmt.Errorf("invalid session path: %s", p)
	}

	return id, nil
}

// newID returns a random, hyphen-free identifier prefixed with p. The result is
// safe to use as a D-Bus object-path element, which must match [A-Za-z0-9_].
func newID(prefix byte) (string, error) {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", err
	}

	return fmt.Sprintf("%c%x", prefix, b[:]), nil
}
