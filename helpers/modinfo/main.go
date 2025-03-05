// modinfo a small helper to print the build info and module versions.
//
// Test builds don't have build info, so this will only work in a real build.
package main

import (
	"fmt"
	rd "runtime/debug"

	_ "github.com/blang/semver/v4"
	"github.com/gopasspw/gopass/pkg/debug"
)

func main() {
	info, ok := rd.ReadBuildInfo()
	if !ok {
		panic("could not read build info")
	}

	fmt.Printf("Build Info: %+v\n", info)

	for _, v := range []string{
		"github.com/blang/semver/v4",
		"github.com/gopasspw/gopass/internal/backend/storage/fs",
	} {
		mv := debug.ModuleVersion(v)
		fmt.Printf("Module Version: %s %s\n", v, mv)
	}
}
