//go:build linux

package secretservice

import "testing"

func TestItemDBusIDRoundTrip(t *testing.T) {
	for _, name := range []string{
		"i0123456789abcdef",
		"cli-inserted",
		"with space",
		"with.dot",
		"x",
		"x6162",
		"ünïcode",
		"",
		"slash/name",
	} {
		id := ItemDBusID(name)

		if name == "" {
			// The empty name cannot be a valid element; the caller filters it.
			if id != "x" {
				t.Fatalf("ItemDBusID(%q) = %q, want %q", name, id, "x")
			}

			continue
		}

		if !isPathSafe(id) {
			t.Fatalf("ItemDBusID(%q) = %q is not a valid path element", name, id)
		}
		if got := ItemNameFromDBusID(id); got != name {
			t.Fatalf("round trip: ItemNameFromDBusID(ItemDBusID(%q)) = %q", name, got)
		}
	}
}

func TestItemDBusIDVerbatimForSafeNames(t *testing.T) {
	// Generated IDs keep their historic layout.
	if got := ItemDBusID("i0123456789abcdef"); got != "i0123456789abcdef" {
		t.Fatalf("safe id was rewritten: %q", got)
	}
	// A name starting with "x" must be encoded so decoding stays unambiguous.
	if got := ItemDBusID("xyz"); got == "xyz" {
		t.Fatalf("x-prefixed name must be encoded, got %q", got)
	}
}

func TestItemDBusPathHasNoHyphens(t *testing.T) {
	p := ItemDBusPath("default", "cli-inserted")
	if str := string(p); len(str) == 0 {
		t.Fatal("empty path")
	}

	for _, r := range string(p) {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9', r == '_', r == '/':
		default:
			t.Fatalf("path %q contains invalid character %q", p, r)
		}
	}
}
