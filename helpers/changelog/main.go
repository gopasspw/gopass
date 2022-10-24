// Copyright 2021 The gopass Authors. All rights reserved.
// Use of this source code is governed by the MIT license,
// that can be found in the LICENSE file.

// Changelog implements the changelog extractor that is called by the autorelease GitHub action
// and used to extract the changelog from the CHANGELOG.md file. It's content is then used to
// populate the release description on GitHub.
//
// This tool will extract every line between the first and the second subheading (##).
// This way the changelog can have a common header under the top most heading (#) and we
// still only get the content of the latest release in the GitHub release notes.
package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func main() {
	fh, err := os.Open("CHANGELOG.md")
	if err != nil {
		panic(err)
	}
	defer fh.Close()

	s := bufio.NewScanner(fh)
	var in bool
	for s.Scan() {
		line := s.Text()
		if strings.HasPrefix(line, "## ") {
			if in {
				break
			}
			in = true
		}

		if !in {
			continue
		}

		fmt.Println(line)
	}
}
