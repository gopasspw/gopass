// Copyright 2026 The gopass Authors. All rights reserved.
// Use of this source code is governed by the MIT license,
// that can be found in the LICENSE file.

// Package commitmsg parses commit subjects and classifies them into Keep a
// Changelog sections. It is the single source of truth shared by the changelog
// generator in helpers/release and by any commit message linting, so the two
// cannot disagree about what a valid subject looks like.
//
// The conventions it implements are specified in docs/conventions.md.
package commitmsg

import (
	"regexp"
	"strings"
)

// Section is a Keep a Changelog 1.1.0 subsection.
type Section string

// The six sections defined by Keep a Changelog 1.1.0.
const (
	Added      Section = "Added"
	Changed    Section = "Changed"
	Deprecated Section = "Deprecated"
	Removed    Section = "Removed"
	Fixed      Section = "Fixed"
	Security   Section = "Security"

	// Omit marks a commit type that produces no changelog entry.
	Omit Section = ""
)

// Sections lists the sections in the order Keep a Changelog prescribes.
// Render them in this order and skip the empty ones.
var Sections = []Section{Added, Changed, Deprecated, Removed, Fixed, Security}

// Types maps every accepted Conventional Commit type to its changelog section.
// The list is closed: a subject using any other type is not a valid commit
// message. Types mapping to Omit produce no changelog entry.
var Types = map[string]Section{
	"build":    Omit,
	"chore":    Omit,
	"ci":       Omit,
	"deps":     Changed,
	"docs":     Omit,
	"feat":     Added,
	"fix":      Fixed,
	"perf":     Changed,
	"refactor": Changed,
	"revert":   Changed,
	"security": Security,
	"test":     Omit,
}

// legacyTags maps the bracketed tags used before the move to Conventional
// Commits. They are recognised so that a release spanning the transition
// classifies its older commits correctly.
var legacyTags = map[string]Section{
	"BREAKING":      Changed,
	"BUGFIX":        Fixed,
	"CLEANUP":       Changed,
	"DEPRECATION":   Deprecated,
	"DOCUMENTATION": Omit,
	"ENHANCEMENT":   Changed,
	"FEATURE":       Added,
	"PKG-BREAK":     Changed,
	"SECURITY":      Security,
	"TESTING":       Omit,
	"UX":            Changed,
}

// Prefixes applied to an entry's text.
const (
	breakingPrefix = "**BREAKING** "
	pkgBreakPrefix = "[PKG-BREAK] "
	revertPrefix   = "Revert: "
)

var (
	// conventionalRE matches a Conventional Commits 1.0.0 subject.
	// Groups: type, scope, breaking marker, description.
	conventionalRE = regexp.MustCompile(`^([a-z]+)(?:\(([^)]*)\))?(!)?: (.+)$`)

	// legacyRE matches the bracketed subject form used before the move to
	// Conventional Commits, for example "[BUGFIX] reorg: list all secrets".
	legacyRE = regexp.MustCompile(`^\[([A-Za-z-]+)\]\s+(.+)$`)

	// sectionFooterRE matches the Changelog-Section: footer, which overrides
	// the type-to-section mapping. It is the only way to reach the Removed and
	// Deprecated sections, which have no corresponding commit type.
	sectionFooterRE = regexp.MustCompile(`(?m)^Changelog-Section:\s*(\w+)\s*$`)

	// pkgBreakFooterRE matches the PKG-BREAK: footer mandated by ADR A-12 for a
	// breaking change confined to the pkg/gopass Go module.
	pkgBreakFooterRE = regexp.MustCompile(`(?m)^PKG-BREAK:\s*(.+)$`)

	// breakingFooterRE matches the Conventional Commits footer that marks a
	// break in the CLI.
	breakingFooterRE = regexp.MustCompile(`(?m)^BREAKING[ -]CHANGE:\s*(.+)$`)

	// releaseNotesRE matches the pre-existing RELEASE_NOTES= override. A value
	// of "n/a" suppresses the entry entirely.
	releaseNotesRE = regexp.MustCompile(`(?m)^RELEASE_NOTES=(.*)$`)
)

// Commit is a parsed commit subject.
type Commit struct {
	// Type is the Conventional Commit type, or the empty string for a legacy
	// bracketed subject.
	Type string
	// Scope is the optional parenthesised scope. It is empty when absent.
	Scope string
	// Breaking reports whether the subject carried the "!" marker.
	Breaking bool
	// Description is the subject with the type, scope and marker removed. For a
	// legacy subject it is everything after the bracketed tag.
	Description string
	// Legacy reports whether the subject used the bracketed form.
	Legacy bool
	// Section is the section the type maps to, before any footer override.
	Section Section
}

// Entry is one rendered changelog bullet.
type Entry struct {
	Section Section
	Text    string
}

// Disposition is what Classify decided to do with a commit.
type Disposition int

const (
	// Include means the commit produced a changelog entry.
	Include Disposition = iota
	// Omitted means the commit was deliberately excluded: its type maps to
	// Omit, or its body carried RELEASE_NOTES=n/a. This is the expected
	// outcome for dependency bumps, CI changes and documentation.
	Omitted
	// Unrecognised means the subject is not a valid commit message, so no
	// section could be chosen. Callers must surface these rather than drop
	// them silently: a real user-facing change written as "otp: hide --snip
	// flag" -- a scope used as a type -- would otherwise vanish from the
	// changelog without a trace.
	Unrecognised
)

// Parse splits a commit subject into its parts. It reports false when the
// subject is neither a Conventional Commit with a known type nor a legacy
// bracketed subject with a known tag.
func Parse(subject string) (Commit, bool) {
	subject = strings.TrimSpace(subject)

	if m := conventionalRE.FindStringSubmatch(subject); m != nil {
		sec, ok := Types[m[1]]
		if !ok {
			return Commit{}, false
		}

		return Commit{
			Type:        m[1],
			Scope:       m[2],
			Breaking:    m[3] == "!",
			Description: m[4],
			Section:     sec,
		}, true
	}

	if m := legacyRE.FindStringSubmatch(subject); m != nil {
		if sec, ok := legacyTags[strings.ToUpper(m[1])]; ok {
			return Commit{
				Description: m[2],
				Legacy:      true,
				Section:     sec,
				Breaking:    strings.EqualFold(m[1], "BREAKING"),
			}, true
		}

		// The history also contains a hybrid form that puts a Conventional
		// Commit type in brackets, for example "[fix] Support HW Age
		// identities" and "[chore] Update gopasspw/clipboard". Accept it so a
		// release spanning the transition does not lose those entries.
		if sec, ok := Types[strings.ToLower(m[1])]; ok {
			return Commit{
				Type:        strings.ToLower(m[1]),
				Description: m[2],
				Legacy:      true,
				Section:     sec,
			}, true
		}

		return Commit{}, false
	}

	return Commit{}, false
}

// Text renders the changelog bullet text for a parsed commit: the description,
// prefixed with the scope when there is one. The type is dropped because the
// section heading already carries it -- "fix(age): expand ~/" under a "Fixed"
// heading renders as "age: expand ~/".
func (c Commit) Text() string {
	if c.Scope == "" {
		return c.Description
	}

	return c.Scope + ": " + c.Description
}

// Classify turns a commit subject and body into a changelog entry, and reports
// why it did so.
//
// Precedence, highest first: RELEASE_NOTES=n/a suppresses everything;
// RELEASE_NOTES=<text> replaces the text; Changelog-Section: replaces the
// section; PKG-BREAK: forces an entry and adds a prefix; "!" or
// BREAKING CHANGE: adds a prefix.
func Classify(subject, body string) (Entry, Disposition) {
	notes, hasNotes := releaseNotes(body)
	if hasNotes && notes == "" {
		// RELEASE_NOTES=n/a: an explicit request to stay out of the changelog.
		return Entry{}, Omitted
	}

	c, ok := Parse(subject)
	if !ok && !hasNotes {
		return Entry{}, Unrecognised
	}

	sec := c.Section
	text := c.Text()

	if hasNotes {
		text = notes
		if !ok {
			// An unparseable subject with an explicit release note still
			// deserves an entry; Changed is the least surprising home for it.
			sec = Changed
		}
	}

	pkgBreak := pkgBreakFooterRE.FindStringSubmatch(body)
	if pkgBreak != nil && sec == Omit {
		// A pkg/gopass break must always be visible, whatever the type says.
		sec = Changed
	}

	if m := sectionFooterRE.FindStringSubmatch(body); m != nil {
		if s, valid := parseSection(m[1]); valid {
			sec = s
		}
	}

	if sec == Omit {
		return Entry{}, Omitted
	}

	if c.Breaking || breakingFooterRE.MatchString(body) {
		text = breakingPrefix + text
	}

	if pkgBreak != nil {
		text = pkgBreakPrefix + text
	}

	if c.Type == "revert" {
		text = revertPrefix + text
	}

	return Entry{Section: sec, Text: text}, Include
}

// parseSection resolves a Changelog-Section: footer value, case-insensitively.
func parseSection(in string) (Section, bool) {
	for _, s := range Sections {
		if strings.EqualFold(in, string(s)) {
			return s, true
		}
	}

	return Omit, false
}

// releaseNotes extracts the RELEASE_NOTES= override from a commit body. It
// returns an empty string with ok set when the value is "n/a", which means the
// commit is deliberately excluded from the changelog.
func releaseNotes(body string) (string, bool) {
	m := releaseNotesRE.FindStringSubmatch(body)
	if m == nil {
		return "", false
	}

	val := strings.TrimSpace(m[1])
	if strings.EqualFold(val, "n/a") {
		return "", true
	}

	if val == "" {
		return "", false
	}

	return val, true
}
