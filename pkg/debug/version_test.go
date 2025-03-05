package debug

import (
	"testing"

	rdebug "runtime/debug"

	"github.com/blang/semver/v4"
	"github.com/stretchr/testify/assert"
)

func TestModuleVersion(t *testing.T) {
	tests := []struct {
		name   string
		pkg    string
		module string
		modver string
		want   semver.Version
		noBI   bool
	}{
		{
			name:   "valid module version semver",
			pkg:    "github.com/blang/semver/v4",
			module: "github.com/blang/semver/",
			modver: "v4.0.0",
			want:   semver.MustParse("4.0.0"),
		},
		{
			name:   "valid module version gopass",
			module: "github.com/gopasspw/gopass",
			pkg:    "github.com/gopasspw/gopass/internal/backend/storage/fs",
			modver: "v4.0.0",
			want:   semver.MustParse("4.0.0"),
		},
		{
			name:   "invalid module version",
			module: "invalid/module",
			modver: "",
			want:   semver.Version{},
		},
		{
			name:   "non-existent module",
			module: "non/existent/module",
			modver: "",
			want:   semver.Version{},
		},
		{
			name:   "build-info failure",
			module: "non/existent/module",
			modver: "",
			want:   semver.Version{},
			noBI:   true,
		},

		{
			name:   "invalid version",
			module: "some/module/with/invalid/version",
			modver: "devel",
			want:   semver.Version{Build: []string{"devel"}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			biFunc = func() (*rdebug.BuildInfo, bool) {
				return &rdebug.BuildInfo{
					Main: rdebug.Module{
						Version: "v4.0.0",
					},
					Deps: []*rdebug.Module{
						{
							Path:    tt.module,
							Version: tt.modver,
						},
					},
				}, true
			}
			if tt.noBI {
				biFunc = func() (*rdebug.BuildInfo, bool) {
					return nil, false
				}
			}
			ask := tt.pkg
			if ask == "" {
				ask = tt.module
			}
			got := ModuleVersion(ask)
			assert.True(t, got.Equals(tt.want), "ModuleVersion(%s) = %v, want %v", ask, got, tt.want)
		})
	}
}
