// Copyright 2026 The gopass Authors. All rights reserved.
// Use of this source code is governed by the MIT license,
// that can be found in the LICENSE file.

package main

import (
	"strings"
	"testing"

	"github.com/blang/semver/v4"
	"github.com/gopasspw/gopass/helpers/commitmsg"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const migratedFixture = `# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/).

## [Unreleased]

### Security

- fix path traversal in the fs storage layer (C-1)

### Added

- add the gopass doctor diagnostic command (I-4)

## [1.16.1] - 2025-12-13

* fix: Fix version check against latest release (#3292)

## [1.16.0] - 2025-11-12

* [BUGFIX] reorg: List all secrets (#3245)

[Unreleased]: https://github.com/gopasspw/gopass/compare/v1.16.1...HEAD
[1.16.1]: https://github.com/gopasspw/gopass/compare/v1.16.0...v1.16.1
[1.16.0]: https://github.com/gopasspw/gopass/compare/v1.15.18...v1.16.0
`

func TestParseChangelog(t *testing.T) {
	t.Parallel()

	cl, err := parseChangelog(strings.NewReader(migratedFixture))
	require.NoError(t, err)

	assert.Contains(t, strings.Join(cl.header, "\n"), "# Changelog")
	assert.NotContains(t, strings.Join(cl.header, "\n"), "## [")

	assert.Equal(t,
		[]string{"fix path traversal in the fs storage layer (C-1)"},
		cl.unreleased[commitmsg.Security])
	assert.Equal(t,
		[]string{"add the gopass doctor diagnostic command (I-4)"},
		cl.unreleased[commitmsg.Added])

	released := strings.Join(cl.released, "\n")
	assert.Contains(t, released, "## [1.16.1] - 2025-12-13")
	assert.Contains(t, released, "## [1.16.0] - 2025-11-12")
	assert.NotContains(t, released, "[Unreleased]:")

	require.Len(t, cl.links, 3)
	assert.Equal(t, "[Unreleased]: https://github.com/gopasspw/gopass/compare/v1.16.1...HEAD", cl.links[0])
}

// TestReleaseConsumesUnreleased is the regression this change exists for.
//
// The previous writeChangelog inserted the new section before the first "## "
// heading. Once an unreleased section existed, that heading was the unreleased
// one, so the release landed *above* it and the hand-written entries below were
// orphaned: never published, and silently carried into every later release.
func TestReleaseConsumesUnreleased(t *testing.T) {
	t.Parallel()

	cl, err := parseChangelog(strings.NewReader(migratedFixture))
	require.NoError(t, err)

	cl.release(
		semver.MustParse("1.16.1"),
		semver.MustParse("1.17.0"),
		"2026-08-01",
		[]commitmsg.Entry{
			{Section: commitmsg.Fixed, Text: "age: expand ~/ in age.ssh-key-path (#3474)"},
			{Section: commitmsg.Added, Text: "show: add a fuzzy-search toggle (#3449)"},
		},
	)

	var buf strings.Builder

	require.NoError(t, cl.render(&buf))

	got := buf.String()

	// The hand-written entries are now part of the release, not stranded.
	assert.Contains(t, got, "- fix path traversal in the fs storage layer (C-1)")
	assert.Contains(t, got, "- add the gopass doctor diagnostic command (I-4)")

	// ... together with the generated ones.
	assert.Contains(t, got, "- age: expand ~/ in age.ssh-key-path (#3474)")
	assert.Contains(t, got, "- show: add a fuzzy-search toggle (#3449)")

	// The release heading precedes the previous release.
	relIdx := strings.Index(got, "## [1.17.0] - 2026-08-01")
	prevIdx := strings.Index(got, "## [1.16.1] - 2025-12-13")
	require.Positive(t, relIdx)
	require.Positive(t, prevIdx)
	assert.Less(t, relIdx, prevIdx)

	// The unreleased section survives, empty, above the release.
	unrelIdx := strings.Index(got, "## [Unreleased]")
	require.Positive(t, unrelIdx)
	assert.Less(t, unrelIdx, relIdx)
	assert.NotContains(t, got[unrelIdx:relIdx], "- ")
}

func TestReleaseOrdersAndOmitsSections(t *testing.T) {
	t.Parallel()

	cl, err := parseChangelog(strings.NewReader(migratedFixture))
	require.NoError(t, err)

	cl.release(
		semver.MustParse("1.16.1"),
		semver.MustParse("1.17.0"),
		"2026-08-01",
		[]commitmsg.Entry{{Section: commitmsg.Fixed, Text: "a fix"}},
	)

	var buf strings.Builder

	require.NoError(t, cl.render(&buf))

	got := buf.String()

	// Keep a Changelog order: Added, Changed, Deprecated, Removed, Fixed, Security.
	addedIdx := strings.Index(got, "### Added")
	fixedIdx := strings.Index(got, "### Fixed")
	securityIdx := strings.Index(got, "### Security")
	require.Positive(t, addedIdx)
	require.Positive(t, fixedIdx)
	require.Positive(t, securityIdx)
	assert.Less(t, addedIdx, fixedIdx)
	assert.Less(t, fixedIdx, securityIdx)

	// Empty subsections are omitted.
	assert.NotContains(t, got, "### Changed")
	assert.NotContains(t, got, "### Deprecated")
	assert.NotContains(t, got, "### Removed")
}

func TestReleaseDeduplicates(t *testing.T) {
	t.Parallel()

	cl, err := parseChangelog(strings.NewReader(migratedFixture))
	require.NoError(t, err)

	// The same text arrives both by hand and from a commit subject.
	cl.release(
		semver.MustParse("1.16.1"),
		semver.MustParse("1.17.0"),
		"2026-08-01",
		[]commitmsg.Entry{
			{Section: commitmsg.Added, Text: "add the gopass doctor diagnostic command (I-4)"},
		},
	)

	var buf strings.Builder

	require.NoError(t, cl.render(&buf))

	assert.Equal(t, 1, strings.Count(buf.String(), "add the gopass doctor diagnostic command (I-4)"))
}

func TestReleaseUpdatesLinks(t *testing.T) {
	t.Parallel()

	cl, err := parseChangelog(strings.NewReader(migratedFixture))
	require.NoError(t, err)

	cl.release(semver.MustParse("1.16.1"), semver.MustParse("1.17.0"), "2026-08-01", nil)

	var buf strings.Builder

	require.NoError(t, cl.render(&buf))

	got := buf.String()

	assert.Contains(t, got, "[Unreleased]: https://github.com/gopasspw/gopass/compare/v1.17.0...HEAD")
	assert.Contains(t, got, "[1.17.0]: https://github.com/gopasspw/gopass/compare/v1.16.1...v1.17.0")
	// Existing references are preserved.
	assert.Contains(t, got, "[1.16.1]: https://github.com/gopasspw/gopass/compare/v1.16.0...v1.16.1")
	assert.Contains(t, got, "[1.16.0]: https://github.com/gopasspw/gopass/compare/v1.15.18...v1.16.0")
	// The stale Unreleased reference is replaced, not duplicated.
	assert.Equal(t, 1, strings.Count(got, "[Unreleased]: "))
}

// TestParseUnmigratedChangelog covers the transitional state: a file that still
// uses "## Next" and the legacy headings must still parse, with everything from
// the first heading onwards treated as released body.
func TestParseUnmigratedChangelog(t *testing.T) {
	t.Parallel()

	const fixture = `# Changelog

## Next

* [BUGFIX] something not yet released

## 1.16.1 / 2025-12-13

* fix: Fix version check (#3292)
`

	cl, err := parseChangelog(strings.NewReader(fixture))
	require.NoError(t, err)

	assert.Empty(t, cl.unreleased)
	assert.Contains(t, strings.Join(cl.released, "\n"), "## Next")
	assert.Contains(t, strings.Join(cl.released, "\n"), "## 1.16.1 / 2025-12-13")
	assert.Empty(t, cl.links)
}

// TestReleaseBulletOutsideSubsection places an unreleased bullet that sits under
// no "###" heading into Changed rather than dropping it.
func TestReleaseBulletOutsideSubsection(t *testing.T) {
	t.Parallel()

	const fixture = `# Changelog

## [Unreleased]

- a bullet with no subsection

## [1.16.1] - 2025-12-13

* fix: something
`

	cl, err := parseChangelog(strings.NewReader(fixture))
	require.NoError(t, err)

	cl.release(semver.MustParse("1.16.1"), semver.MustParse("1.17.0"), "2026-08-01", nil)

	var buf strings.Builder

	require.NoError(t, cl.render(&buf))

	got := buf.String()
	assert.Contains(t, got, "### Changed")
	assert.Contains(t, got, "- a bullet with no subsection")
}
