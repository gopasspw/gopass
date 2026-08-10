// Copyright 2026 The gopass Authors. All rights reserved.
// Use of this source code is governed by the MIT license,
// that can be found in the LICENSE file.

package main

import (
	"bufio"
	"fmt"
	"io"
	"regexp"
	"slices"
	"strings"

	"github.com/blang/semver/v4"
	"github.com/gopasspw/gopass/helpers/commitmsg"
)

// repoURL is the base for the compare links at the foot of CHANGELOG.md.
const repoURL = "https://github.com/gopasspw/gopass"

var (
	// unreleasedHeadingRE matches the Keep a Changelog unreleased heading.
	unreleasedHeadingRE = regexp.MustCompile(`^## \[Unreleased\]\s*$`)

	// versionHeadingRE matches a released section heading.
	versionHeadingRE = regexp.MustCompile(`^## \[(\d+\.\d+\.\d+[^\]]*)\]`)

	// subsectionRE matches one of the six Keep a Changelog subsections.
	subsectionRE = regexp.MustCompile(`^### (\w+)\s*$`)

	// bulletRE matches a changelog bullet in either the "-" or the legacy "*"
	// form. CHANGELOG.md uses "*" throughout its history.
	bulletRE = regexp.MustCompile(`^[-*]\s+(.*)$`)

	// linkRefRE matches a markdown link reference definition, which Keep a
	// Changelog places in a block at the foot of the file.
	linkRefRE = regexp.MustCompile(`^\[[^\]]+\]:\s+\S+`)
)

// changelog is a parsed CHANGELOG.md.
type changelog struct {
	// header is every line before the first "## " heading.
	header []string
	// unreleased holds the entries of the "## [Unreleased]" section, keyed by
	// subsection. Entries outside any subsection are keyed by commitmsg.Omit.
	unreleased map[commitmsg.Section][]string
	// released is every line from the first versioned heading up to the link
	// reference block, copied through unchanged.
	released []string
	// links is the link reference block at the foot of the file.
	links []string
}

// parseChangelog reads a CHANGELOG.md into its four parts.
//
// It is deliberately tolerant: a file that has not been migrated to Keep a
// Changelog yet still parses, with everything from the first heading onwards
// landing in released.
func parseChangelog(rd io.Reader) (*changelog, error) {
	cl := &changelog{unreleased: map[commitmsg.Section][]string{}}

	const (
		inHeader = iota
		inUnreleased
		inReleased
	)

	state := inHeader
	section := commitmsg.Omit

	s := bufio.NewScanner(rd)
	for s.Scan() {
		line := s.Text()

		switch {
		case unreleasedHeadingRE.MatchString(line):
			state = inUnreleased
			section = commitmsg.Omit

			continue

		case strings.HasPrefix(line, "## "):
			// Any other second-level heading starts the released body. Keep the
			// heading itself.
			state = inReleased

		case state == inHeader:
			cl.header = append(cl.header, line)

			continue
		}

		if state == inUnreleased {
			if m := subsectionRE.FindStringSubmatch(line); m != nil {
				if sec, ok := parseSectionName(m[1]); ok {
					section = sec
				}

				continue
			}

			if m := bulletRE.FindStringSubmatch(line); m != nil {
				cl.unreleased[section] = append(cl.unreleased[section], m[1])
			}

			continue
		}

		if linkRefRE.MatchString(line) {
			cl.links = append(cl.links, line)

			continue
		}

		cl.released = append(cl.released, line)
	}

	if err := s.Err(); err != nil {
		return nil, fmt.Errorf("failed to read changelog: %w", err)
	}

	trimTrailingBlanks(&cl.header)
	trimTrailingBlanks(&cl.released)

	return cl, nil
}

// parseSectionName resolves a "### Foo" heading to a section.
func parseSectionName(in string) (commitmsg.Section, bool) {
	for _, sec := range commitmsg.Sections {
		if strings.EqualFold(in, string(sec)) {
			return sec, true
		}
	}

	return commitmsg.Omit, false
}

func trimTrailingBlanks(lines *[]string) {
	for len(*lines) > 0 && strings.TrimSpace((*lines)[len(*lines)-1]) == "" {
		*lines = (*lines)[:len(*lines)-1]
	}
}

// release folds the unreleased entries together with the generated ones and
// writes them out as a new released section, leaving an empty Unreleased
// section behind and refreshing the link reference block.
//
// Merging the two sources is what makes the hand-written entries in
// "## [Unreleased]" survive a release. The previous implementation inserted the
// new section before the first "## " heading, which is the unreleased heading,
// so those entries were orphaned below the release and never consumed.
func (c *changelog) release(prev, next semver.Version, date string, generated []commitmsg.Entry) {
	merged := map[commitmsg.Section][]string{}

	for sec, texts := range c.unreleased {
		target := sec
		if target == commitmsg.Omit {
			// A bullet that sat outside any subsection. Changed is the least
			// surprising home for it.
			target = commitmsg.Changed
		}
		merged[target] = append(merged[target], texts...)
	}

	for _, e := range generated {
		merged[e.Section] = append(merged[e.Section], e.Text)
	}

	for sec := range merged {
		slices.Sort(merged[sec])
		merged[sec] = dedupeFold(merged[sec])
	}

	section := []string{fmt.Sprintf("## [%s] - %s", next.String(), date), ""}

	for _, sec := range commitmsg.Sections {
		if len(merged[sec]) == 0 {
			// Keep a Changelog: omit empty subsections.
			continue
		}

		section = append(section, "### "+string(sec), "")
		for _, t := range merged[sec] {
			section = append(section, "- "+t)
		}
		section = append(section, "")
	}

	c.released = append(section, c.released...)
	c.unreleased = map[commitmsg.Section][]string{}
	c.updateLinks(prev, next)
}

// dedupeFold removes entries that differ only in case from one already kept.
// The input must be sorted.
//
// A hand-written unreleased entry and the subject of the commit that
// implemented it often describe the same change with a different capital
// letter, for example "Add gopass doctor diagnostic command (I-4)" against
// "add gopass doctor diagnostic command (I-4)". It cannot catch two genuinely
// different wordings of the same change; those still need a human pass before
// the release.
func dedupeFold(in []string) []string {
	out := in[:0]
	seen := make(map[string]struct{}, len(in))

	for _, s := range in {
		k := strings.ToLower(s)
		if _, ok := seen[k]; ok {
			continue
		}

		seen[k] = struct{}{}
		out = append(out, s)
	}

	return out
}

// updateLinks rewrites the two link references a release changes: Unreleased
// now compares against the new tag, and the new version gets its own compare
// link against the previous one. Every other reference is left alone.
func (c *changelog) updateLinks(prev, next semver.Version) {
	unreleased := fmt.Sprintf("[Unreleased]: %s/compare/v%s...HEAD", repoURL, next.String())
	version := fmt.Sprintf("[%s]: %s/compare/v%s...v%s", next.String(), repoURL, prev.String(), next.String())

	links := []string{unreleased, version}

	for _, l := range c.links {
		if strings.HasPrefix(l, "[Unreleased]:") {
			continue
		}
		if strings.HasPrefix(l, "["+next.String()+"]:") {
			continue
		}
		links = append(links, l)
	}

	c.links = links
}

// render writes the changelog back out.
func (c *changelog) render(w io.Writer) error {
	out := bufio.NewWriter(w)

	for _, l := range c.header {
		fmt.Fprintln(out, l)
	}

	fmt.Fprintln(out)
	fmt.Fprintln(out, "## [Unreleased]")
	fmt.Fprintln(out)

	for _, l := range c.released {
		fmt.Fprintln(out, l)
	}

	if len(c.links) > 0 {
		fmt.Fprintln(out)
		for _, l := range c.links {
			fmt.Fprintln(out, l)
		}
	}

	if err := out.Flush(); err != nil {
		return fmt.Errorf("failed to write changelog: %w", err)
	}

	return nil
}
