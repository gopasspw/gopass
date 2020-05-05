package main

import (
	"bytes"
	"fmt"
	"runtime"
	"testing"

	"github.com/blang/semver"
	"github.com/stretchr/testify/assert"
)

func TestVersionPrinter(t *testing.T) {
	t.Parallel()

	buf := &bytes.Buffer{}
	vp := makeVersionPrinter(buf, semver.Version{Major: 1})
	vp(nil)
	assert.Equal(t, fmt.Sprintf("gopass 1.0.0 %s %s %s\n", runtime.Version(), runtime.GOOS, runtime.GOARCH), buf.String())
}

func TestGetVersion(t *testing.T) {
	t.Parallel()

	version = "1.9.0"

	if getVersion().LT(semver.Version{Major: 1, Minor: 9}) {
		t.Errorf("invalid version")
	}
}
