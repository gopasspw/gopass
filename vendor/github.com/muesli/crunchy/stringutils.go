/*
 * crunchy - find common flaws in passwords
 *     Copyright (c) 2017, Christian Muehlhaeuser <muesli@gmail.com>
 *
 *   For license see LICENSE
 */

package crunchy

import (
	"encoding/hex"
	"hash"
	"strings"
)

// countUniqueChars returns the amount of unique runes in a string
func countUniqueChars(s string) int {
	m := make(map[rune]struct{})

	for _, c := range s {
		if _, ok := m[c]; !ok {
			m[c] = struct{}{}
		}
	}

	return len(m)
}

// countSystematicChars returns how many runes in a string are part of a sequence ('abcdef', '654321')
func countSystematicChars(s string) int {
	var x int
	rs := []rune(s)

	for i, c := range rs {
		if i == 0 {
			continue
		}
		if c == rs[i-1]+1 || c == rs[i-1]-1 {
			x++
		}
	}

	return x
}

// reverse returns the reversed form of a string
func reverse(s string) string {
	rs := []rune(s)
	for i, j := 0, len(rs)-1; i < j; i, j = i+1, j-1 {
		rs[i], rs[j] = rs[j], rs[i]
	}
	return string(rs)
}

// normalize returns the trimmed and lowercase version of a string
func normalize(s string) string {
	return strings.TrimSpace(strings.ToLower(s))
}

// hashsum returns the hashed sum of a string
func hashsum(s string, hasher hash.Hash) string {
	hasher.Reset()
	hasher.Write([]byte(s))
	return hex.EncodeToString(hasher.Sum(nil))
}
