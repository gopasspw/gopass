// Copyright 2021 The gopass Authors. All rights reserved.
// Use of this source code is governed by the MIT license,
// that can be found in the LICENSE file.

// Changelog implements the changelog extractor that is called by the autorelease
// GitHub action and used to extract the changelog from the CHANGELOG.md file.
// It's content is then used to populate the release description on GitHub.
//
// This tool extracts the most recent *released* section: everything from the
// first "## [X.Y.Z]" heading up to the next "## " heading. The "## [Unreleased]"
// section is skipped, because at release time the release helper has already
// moved its content into the new version's section and left it empty.
package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"
)

var filename = "CHANGELOG.md"

// versionHeadingRE matches a released section heading in Keep a Changelog form,
// for example "## [1.16.1] - 2025-12-13". The legacy form
// "## 1.16.1 / 2025-12-13" is accepted as well so the tool keeps working on a
// changelog that has not been migrated.
var versionHeadingRE = regexp.MustCompile(`^## \[?\d+\.\d+\.\d+`)

func main() {
	fh, err := os.Open(filename)
	if err != nil {
		panic(err)
	}
	defer fh.Close()

	if err := extract(fh, os.Stdout); err != nil {
		panic(err)
	}
}

// extract copies the first released section of a changelog to w.
func extract(rd io.Reader, w io.Writer) error {
	s := bufio.NewScanner(rd)
	out := bufio.NewWriter(w)

	var in bool

	for s.Scan() {
		line := s.Text()

		if strings.HasPrefix(line, "## ") {
			if in {
				break
			}

			// Skip everything before the first released section: the file
			// header and the (empty) Unreleased section.
			if !versionHeadingRE.MatchString(line) {
				continue
			}

			in = true
		}

		if !in {
			continue
		}

		fmt.Fprintln(out, line)
	}

	if err := s.Err(); err != nil {
		return fmt.Errorf("failed to read %s: %w", filename, err)
	}

	if err := out.Flush(); err != nil {
		return fmt.Errorf("failed to write release notes: %w", err)
	}

	return nil
}
