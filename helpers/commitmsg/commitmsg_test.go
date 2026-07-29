// Copyright 2026 The gopass Authors. All rights reserved.
// Use of this source code is governed by the MIT license,
// that can be found in the LICENSE file.

package commitmsg

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseConventional(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name    string
		subject string
		want    Commit
	}{
		{
			name:    "type only",
			subject: "fix: avoid NPE when editing a non-existing secret",
			want: Commit{
				Type:        "fix",
				Description: "avoid NPE when editing a non-existing secret",
				Section:     Fixed,
			},
		},
		{
			name:    "type and scope",
			subject: "fix(age): expand ~/ in age.ssh-key-path",
			want: Commit{
				Type:        "fix",
				Scope:       "age",
				Description: "expand ~/ in age.ssh-key-path",
				Section:     Fixed,
			},
		},
		{
			name:    "breaking marker",
			subject: "feat(show)!: drop the -f alias",
			want: Commit{
				Type:        "feat",
				Scope:       "show",
				Breaking:    true,
				Description: "drop the -f alias",
				Section:     Added,
			},
		},
		{
			name:    "slash in scope",
			subject: "refactor(pkg/gopass): drop Store.GetRevision",
			want: Commit{
				Type:        "refactor",
				Scope:       "pkg/gopass",
				Description: "drop Store.GetRevision",
				Section:     Changed,
			},
		},
		{
			name:    "dependabot subject",
			subject: "chore(deps): bump actions/checkout from 6.0.3 to 7.0.0 (#3485)",
			want: Commit{
				Type:        "chore",
				Scope:       "deps",
				Description: "bump actions/checkout from 6.0.3 to 7.0.0 (#3485)",
				Section:     Omit,
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, ok := Parse(tc.subject)
			require.True(t, ok, tc.subject)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestParseLegacy(t *testing.T) {
	t.Parallel()

	got, ok := Parse("[BUGFIX] reorg: List all secrets instead of just top-level folders (#3245)")
	require.True(t, ok)
	assert.True(t, got.Legacy)
	assert.Equal(t, Fixed, got.Section)
	assert.Equal(t, "reorg: List all secrets instead of just top-level folders (#3245)", got.Description)
	assert.Empty(t, got.Type)
}

func TestParseRejects(t *testing.T) {
	t.Parallel()

	// These appear as commit types in the history but are scopes, so they must
	// not parse as Conventional Commits.
	for _, subject := range []string{
		"otp: hide --snip flag when built with noscreenshot tag (#3445)",
		"age: fix YubiKey identity persistence via raw-append (ADR-0002) (#3399)",
		"bug: reload identities on unlock command (#3430)",
		"fscopy: derive copy direction from the source argument (#3462)",
		"openbsd: something",
		"[NOSUCHTAG] whatever",
		"Improve Team Workflows (#3460)",
		"",
	} {
		_, ok := Parse(subject)
		assert.False(t, ok, subject)
	}
}

func TestText(t *testing.T) {
	t.Parallel()

	c, ok := Parse("fix(age): expand ~/ in age.ssh-key-path")
	require.True(t, ok)
	assert.Equal(t, "age: expand ~/ in age.ssh-key-path", c.Text())

	c, ok = Parse("fix: restore clip flag through fuzzy search")
	require.True(t, ok)
	assert.Equal(t, "restore clip flag through fuzzy search", c.Text())
}

func TestClassify(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name     string
		subject  string
		body     string
		want     Entry
		wantOK   bool
		wantDisp Disposition
	}{
		{
			name:    "feat to Added",
			subject: "feat(show): add configurable fuzzy-search fallback toggle",
			want:    Entry{Section: Added, Text: "show: add configurable fuzzy-search fallback toggle"},
			wantOK:  true,
		},
		{
			name:    "fix to Fixed",
			subject: "fix: restore clip flag through fuzzy search",
			want:    Entry{Section: Fixed, Text: "restore clip flag through fuzzy search"},
			wantOK:  true,
		},
		{
			name:    "security to Security",
			subject: "security: bound symlink traversal to the store root",
			want:    Entry{Section: Security, Text: "bound symlink traversal to the store root"},
			wantOK:  true,
		},
		{
			name:    "refactor to Changed",
			subject: "refactor(store): split the leaf recipient handling",
			want:    Entry{Section: Changed, Text: "store: split the leaf recipient handling"},
			wantOK:  true,
		},
		{
			name:    "deps to Changed",
			subject: "deps: bump golang.org/x/crypto to 0.54.0",
			want:    Entry{Section: Changed, Text: "bump golang.org/x/crypto to 0.54.0"},
			wantOK:  true,
		},
		{
			name:     "docs omitted",
			wantDisp: Omitted,
			subject:  "docs(adr): add the MCP server record",
			wantOK:   false,
		},
		{
			name:     "chore omitted",
			wantDisp: Omitted,
			subject:  "chore(deps): bump actions/checkout from 6.0.3 to 7.0.0",
			wantOK:   false,
		},
		{
			name:     "ci omitted",
			wantDisp: Omitted,
			subject:  "ci: bump golangci-lint to v2.12.2",
			wantOK:   false,
		},
		{
			name:    "breaking marker prefixes the text",
			subject: "feat(show)!: drop the -f alias",
			want:    Entry{Section: Added, Text: "**BREAKING** show: drop the -f alias"},
			wantOK:  true,
		},
		{
			name:    "breaking footer prefixes the text",
			subject: "feat(show): drop the -f alias",
			body:    "BREAKING CHANGE: -f no longer aliases --unsafe\n",
			want:    Entry{Section: Added, Text: "**BREAKING** show: drop the -f alias"},
			wantOK:  true,
		},
		{
			name:    "hyphenated breaking footer is accepted",
			subject: "feat(show): drop the -f alias",
			body:    "BREAKING-CHANGE: -f no longer aliases --unsafe\n",
			want:    Entry{Section: Added, Text: "**BREAKING** show: drop the -f alias"},
			wantOK:  true,
		},
		{
			name:    "pkg break prefixes the text without forcing a major",
			subject: "refactor(pkg/gopass): drop Store.GetRevision",
			body:    "PKG-BREAK: pkg/gopass: remove Store.GetRevision, use Store.History\n",
			want:    Entry{Section: Changed, Text: "[PKG-BREAK] pkg/gopass: drop Store.GetRevision"},
			wantOK:  true,
		},
		{
			name:    "pkg break forces an entry for an otherwise omitted type",
			subject: "chore(pkg/gopass): tidy up",
			body:    "PKG-BREAK: pkg/gopass: remove Store.Foo\n",
			want:    Entry{Section: Changed, Text: "[PKG-BREAK] pkg/gopass: tidy up"},
			wantOK:  true,
		},
		{
			name:    "section footer overrides the type",
			subject: "refactor(hook): delete the dead hook system",
			body:    "Changelog-Section: Removed\n",
			want:    Entry{Section: Removed, Text: "hook: delete the dead hook system"},
			wantOK:  true,
		},
		{
			name:    "section footer reaches Deprecated",
			subject: "refactor(env): mark GOPASS_AUTOSYNC_INTERVAL as going away",
			body:    "Changelog-Section: Deprecated\n",
			want:    Entry{Section: Deprecated, Text: "env: mark GOPASS_AUTOSYNC_INTERVAL as going away"},
			wantOK:  true,
		},
		{
			name:    "section footer rescues an omitted type",
			subject: "chore(env): drop the legacy variable",
			body:    "Changelog-Section: Removed\n",
			want:    Entry{Section: Removed, Text: "env: drop the legacy variable"},
			wantOK:  true,
		},
		{
			name:    "invalid section footer falls back to the type",
			subject: "fix(env): correct the lookup order",
			body:    "Changelog-Section: Nonsense\n",
			want:    Entry{Section: Fixed, Text: "env: correct the lookup order"},
			wantOK:  true,
		},
		{
			name:    "release notes override the text",
			subject: "fix: something terse",
			body:    "RELEASE_NOTES=A much better description\n",
			want:    Entry{Section: Fixed, Text: "A much better description"},
			wantOK:  true,
		},
		{
			name:     "release notes n/a suppresses the entry",
			wantDisp: Omitted,
			subject:  "feat(show): add a thing",
			body:     "RELEASE_NOTES=n/a\n",
			wantOK:   false,
		},
		{
			name:    "release notes rescue an unparseable subject",
			subject: "Improve Team Workflows (#3460)",
			body:    "RELEASE_NOTES=Improve team workflows\n",
			want:    Entry{Section: Changed, Text: "Improve team workflows"},
			wantOK:  true,
		},
		{
			name:    "revert is prefixed",
			subject: "revert: undo the fuzzy search toggle",
			want:    Entry{Section: Changed, Text: "Revert: undo the fuzzy search toggle"},
			wantOK:  true,
		},
		{
			name:    "legacy BUGFIX to Fixed",
			subject: "[BUGFIX] reorg: List all secrets instead of just top-level folders",
			want:    Entry{Section: Fixed, Text: "reorg: List all secrets instead of just top-level folders"},
			wantOK:  true,
		},
		{
			name:    "legacy SECURITY to Security",
			subject: "[SECURITY] Fix path traversal vulnerability in fs storage layer (C-1)",
			want:    Entry{Section: Security, Text: "Fix path traversal vulnerability in fs storage layer (C-1)"},
			wantOK:  true,
		},
		{
			name:     "legacy DOCUMENTATION omitted",
			wantDisp: Omitted,
			subject:  "[DOCUMENTATION] Fix documentation vs. implementation mismatches",
			wantOK:   false,
		},
		{
			name:    "legacy BREAKING is prefixed",
			subject: "[BREAKING] Remove the old config format",
			want:    Entry{Section: Changed, Text: "**BREAKING** Remove the old config format"},
			wantOK:  true,
		},
		{
			name:     "unrecognised subject produces nothing",
			wantDisp: Unrecognised,
			subject:  "Improve Team Workflows (#3460)",
			wantOK:   false,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, disp := Classify(tc.subject, tc.body)
			assert.Equal(t, tc.wantOK, disp == Include, tc.subject)
			if !tc.wantOK {
				assert.Equal(t, tc.wantDisp, disp, tc.subject)

				return
			}
			assert.Equal(t, tc.want, got)
		})
	}
}

// TestEveryTypeHasASection guards against adding a type to docs/conventions.md
// without deciding where its entries go.
func TestEveryTypeHasASection(t *testing.T) {
	t.Parallel()

	want := []string{
		"build", "chore", "ci", "deps", "docs", "feat",
		"fix", "perf", "refactor", "revert", "security", "test",
	}

	assert.Len(t, Types, len(want))

	for _, ty := range want {
		sec, ok := Types[ty]
		require.True(t, ok, ty)

		if sec == Omit {
			continue
		}

		assert.Contains(t, Sections, sec, ty)
	}
}

func TestSectionsAreKeepAChangelogOrder(t *testing.T) {
	t.Parallel()

	assert.Equal(
		t,
		[]Section{Added, Changed, Deprecated, Removed, Fixed, Security},
		Sections,
	)
}

// TestParseBracketedType covers the hybrid form found in the history, which
// puts a Conventional Commit type in brackets rather than a legacy tag.
func TestParseBracketedType(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		subject string
		wantSec Section
		wantTxt string
	}{
		{"[fix] Support HW Age identities (#3389)", Fixed, "Support HW Age identities (#3389)"},
		{"[feat] Add --safe flag to set safecontent on demand (#3318)", Added, "Add --safe flag to set safecontent on demand (#3318)"},
		{"[chore] Update gopasspw/clipboard (#3436)", Omit, ""},
	} {
		c, ok := Parse(tc.subject)
		require.True(t, ok, tc.subject)
		assert.Equal(t, tc.wantSec, c.Section, tc.subject)

		e, disp := Classify(tc.subject, "")
		if tc.wantSec == Omit {
			assert.Equal(t, Omitted, disp, tc.subject)

			continue
		}
		require.Equal(t, Include, disp, tc.subject)
		assert.Equal(t, tc.wantTxt, e.Text)
	}
}

// TestUnrecognisedIsReported guards the distinction that keeps real changes
// from vanishing: a scope written where a type belongs must be reported, not
// silently dropped like an intentional omission.
func TestUnrecognisedIsReported(t *testing.T) {
	t.Parallel()

	for _, subject := range []string{
		"otp: hide --snip flag when built with noscreenshot tag (#3445)",
		"fscopy: derive copy direction from the source argument (#3462)",
		"Improve Team Workflows (#3460)",
	} {
		_, disp := Classify(subject, "")
		assert.Equal(t, Unrecognised, disp, subject)
	}

	// An intentional omission must not be reported.
	_, disp := Classify("chore(deps): bump actions/checkout from 6.0.3 to 7.0.0", "")
	assert.Equal(t, Omitted, disp)
}
