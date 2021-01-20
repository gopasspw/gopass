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
		fmt.Println(line)
	}
}
