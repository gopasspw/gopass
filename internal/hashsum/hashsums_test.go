package hashsum

import (
	"testing"
)

func TestMD5Hex(t *testing.T) {
	t.Parallel()

	in := "test"
	want := "098f6bcd4621d373cade4e832627b4f6"
	got := MD5Hex(in)
	if got != want {
		t.Errorf("MD5Hex(%q) = %q, want %q", in, got, want)
	}
}

func TestSHA1Hex(t *testing.T) {
	t.Parallel()

	in := "test"
	want := "a94a8fe5ccb19ba61c4c0873d391e987982fbbd3"
	got := SHA1Hex(in)
	if got != want {
		t.Errorf("SHA1Hex(%q) = %q, want %q", in, got, want)
	}
}

func TestSHA256Hex(t *testing.T) {
	t.Parallel()

	in := "test"
	want := "9f86d081884c7d659a2feaa0c55ad015a3bf4f1b2b0b822cd15d6c15b0f00a08"
	got := SHA256Hex(in)
	if got != want {
		t.Errorf("SHA256Hex(%q) = %q, want %q", in, got, want)
	}
}

func TestSHA512Hex(t *testing.T) {
	t.Parallel()

	in := "test"
	want := "ee26b0dd4af7e749aa1a8ee3c10ae9923f618980772e473f8819a5d4940e0db27ac185f8a0e1d5f84f88bc887fd67b143732c304cc5fa9ad8e6f57f50028a8ff"
	got := SHA512Hex(in)
	if got != want {
		t.Errorf("SHA512Hex(%q) = %q, want %q", in, got, want)
	}
}

func TestBlake3Hex(t *testing.T) {
	t.Parallel()

	in := "test"
	want := "4878ca0425c739fa427f7eda20fe845f6b2e46ba5fe2a14df5b1e32f50603215"
	got := Blake3Hex(in)
	if got != want {
		t.Errorf("Blake3Hex(%q) = %q, want %q", in, got, want)
	}
}
