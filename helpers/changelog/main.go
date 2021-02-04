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

	fw, err := os.Create("../RELEASE_NOTES")
	if err != nil {
		panic(err)
	}
	defer fw.Close()

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

		_, err := fmt.Fprintln(fw, line)
		if err != nil {
			panic(err)
		}
	}
}
