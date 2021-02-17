package main

import (
	"strings"

	"github.com/blang/semver/v4"
)

func getVersion() semver.Version {
	sv, err := semver.Parse(strings.TrimPrefix(version, "v"))
	if err == nil {
		if commit != "" {
			sv.Build = []string{commit}
		}
		return sv
	}
	return semver.Version{
		Major: 1,
		Minor: 12,
		Patch: 1,
		Pre: []semver.PRVersion{
			{VersionStr: "git"},
		},
		Build: []string{"HEAD"},
	}
}
