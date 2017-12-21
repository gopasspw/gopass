/*
 * crunchy - find common flaws in passwords
 *     Copyright (c) 2017, Christian Muehlhaeuser <muesli@gmail.com>
 *
 *   For license see LICENSE
 */

package crunchy

import (
	"hash"
	"io/ioutil"
	"path/filepath"
	"strings"
	"sync"

	"github.com/xrash/smetrics"
)

// Validator is used to setup a new password validator with options and dictionaries
type Validator struct {
	options     Options
	once        sync.Once
	wordsMaxLen int
	words       map[string]struct{}
	hashedWords map[string]string
}

// Options contains all the settings for a Validator
type Options struct {
	// MinLength is the minimum length required for a valid password (>=1, default is 8)
	MinLength int
	// MinDiff is the minimum amount of unique characters required for a valid password (>=1, default is 5)
	MinDiff int
	// MinDist is the minimum WagnerFischer distance for mangled password dictionary lookups (>=0, default is 3)
	MinDist int
	// Hashers will be used to find hashed passwords in dictionaries
	Hashers []hash.Hash
	// DictionaryPath contains all the dictionaries that will be parsed (default is /usr/share/dict)
	DictionaryPath string
}

// NewValidator returns a new password validator with default settings
func NewValidator() *Validator {
	return NewValidatorWithOpts(Options{
		MinDist:        -1,
		DictionaryPath: "/usr/share/dict",
	})
}

// NewValidatorWithOpts returns a new password validator with custom settings
func NewValidatorWithOpts(options Options) *Validator {
	if options.MinLength <= 0 {
		options.MinLength = 8
	}
	if options.MinDiff <= 0 {
		options.MinDiff = 5
	}
	if options.MinDist < 0 {
		options.MinDist = 3
	}

	return &Validator{
		options:     options,
		words:       make(map[string]struct{}),
		hashedWords: make(map[string]string),
	}
}

// indexDictionaries parses dictionaries/wordlists
func (v *Validator) indexDictionaries() {
	if v.options.DictionaryPath == "" {
		return
	}

	dicts, err := filepath.Glob(filepath.Join(v.options.DictionaryPath, "*"))
	if err != nil {
		return
	}

	for _, dict := range dicts {
		buf, err := ioutil.ReadFile(dict)
		if err != nil {
			continue
		}

		for _, word := range strings.Split(string(buf), "\n") {
			nw := normalize(word)
			nwlen := len(nw)
			if nwlen > v.wordsMaxLen {
				v.wordsMaxLen = nwlen
			}

			// if a word is smaller than the minimum length minus the minimum distance
			// then any collisons would have been rejected by pre-dictionary checks
			if nwlen >= v.options.MinLength-v.options.MinDist {
				v.words[nw] = struct{}{}
			}

			for _, hasher := range v.options.Hashers {
				v.hashedWords[hashsum(nw, hasher)] = nw
			}
		}
	}
}

// foundInDictionaries returns whether a (mangled) string exists in the indexed dictionaries
func (v *Validator) foundInDictionaries(s string) error {
	v.once.Do(v.indexDictionaries)

	pw := normalize(s)   // normalized password
	revpw := reverse(pw) // reversed password
	pwlen := len(pw)

	// let's check perfect matches first
	// we can skip this if the pw is longer than the longest word in our dictionary
	if pwlen <= v.wordsMaxLen {
		if _, ok := v.words[pw]; ok {
			return &DictionaryError{ErrDictionary, pw, 0}
		}
		if _, ok := v.words[revpw]; ok {
			return &DictionaryError{ErrMangledDictionary, revpw, 0}
		}
	}

	// find hashed dictionary entries
	if word, ok := v.hashedWords[pw]; ok {
		return &HashedDictionaryError{ErrHashedDictionary, word}
	}

	// find mangled / reversed passwords
	// we can skip this if the pw is longer than the longest word plus our minimum distance
	if pwlen <= v.wordsMaxLen+v.options.MinDist {
		for word := range v.words {
			if dist := smetrics.WagnerFischer(word, pw, 1, 1, 1); dist <= v.options.MinDist {
				return &DictionaryError{ErrMangledDictionary, word, dist}
			}
			if dist := smetrics.WagnerFischer(word, revpw, 1, 1, 1); dist <= v.options.MinDist {
				return &DictionaryError{ErrMangledDictionary, word, dist}
			}
		}
	}

	return nil
}

// Check validates a password for common flaws
// It returns nil if the password is considered acceptable.
func (v *Validator) Check(password string) error {
	if strings.TrimSpace(password) == "" {
		return ErrEmpty
	}
	if len(password) < v.options.MinLength {
		return ErrTooShort
	}
	if countUniqueChars(password) < v.options.MinDiff {
		return ErrTooFewChars
	}

	// Inspired by cracklib
	maxrepeat := 3.0 + (0.09 * float64(len(password)))
	if countSystematicChars(password) > int(maxrepeat) {
		return ErrTooSystematic
	}

	return v.foundInDictionaries(password)
}
