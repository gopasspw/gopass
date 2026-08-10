// Copyright 2026 The gopass Authors. All rights reserved.
// Use of this source code is governed by the MIT license,
// that can be found in the LICENSE file.

package main

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const keepAChangelogFixture = `# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/).

## [Unreleased]

## [1.17.0] - 2026-08-01

### Added

- show: add a configurable fuzzy-search fallback toggle (#3449)

### Fixed

- age: expand ~/ in age.ssh-key-path (#3474)

## [1.16.1] - 2025-12-13

* fix: Fix version check against latest release (#3292)

[Unreleased]: https://github.com/gopasspw/gopass/compare/v1.17.0...HEAD
[1.17.0]: https://github.com/gopasspw/gopass/compare/v1.16.1...v1.17.0
`

// TestExtractSkipsUnreleased covers the reason this tool needed changing: the
// previous implementation printed everything between the first and the second
// "## " heading. After the Keep a Changelog migration the first heading is
// "## [Unreleased]", which the release helper leaves empty, so the extractor
// would have produced empty release notes.
func TestExtractSkipsUnreleased(t *testing.T) {
	t.Parallel()

	var buf strings.Builder

	require.NoError(t, extract(strings.NewReader(keepAChangelogFixture), &buf))

	got := buf.String()

	assert.Contains(t, got, "## [1.17.0] - 2026-08-01")
	assert.Contains(t, got, "### Added")
	assert.Contains(t, got, "age: expand ~/ in age.ssh-key-path (#3474)")

	assert.NotContains(t, got, "[Unreleased]")
	assert.NotContains(t, got, "1.16.1")
	assert.NotContains(t, got, "# Changelog")
	assert.NotContains(t, got, "Keep a Changelog")
}

// TestExtractLegacyHeadings keeps the tool working against a changelog that has
// not been migrated.
func TestExtractLegacyHeadings(t *testing.T) {
	t.Parallel()

	const fixture = `# Changelog

## 1.16.1 / 2025-12-13

* fix: Fix version check against latest release (#3292)

## 1.16.0 / 2025-11-12

* [BUGFIX] reorg: List all secrets (#3245)
`

	var buf strings.Builder

	require.NoError(t, extract(strings.NewReader(fixture), &buf))

	got := buf.String()

	assert.Contains(t, got, "## 1.16.1 / 2025-12-13")
	assert.Contains(t, got, "Fix version check against latest release")
	assert.NotContains(t, got, "1.16.0")
}

func TestExtractNoReleasedSection(t *testing.T) {
	t.Parallel()

	const fixture = `# Changelog

## [Unreleased]

### Added

- something not yet released
`

	var buf strings.Builder

	require.NoError(t, extract(strings.NewReader(fixture), &buf))
	assert.Empty(t, buf.String())
}
