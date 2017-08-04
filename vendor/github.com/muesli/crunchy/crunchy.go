package crunchy

import (
	"encoding/hex"
	"errors"
	"hash"
	"io/ioutil"
	"path/filepath"
	"strings"
	"sync"

	"github.com/xrash/smetrics"
)

var (
	// ErrEmpty gets returned when the password is empty or all whitespace
	ErrEmpty = errors.New("Password is empty or all whitespace")
	// ErrTooShort gets returned when the password is not long enough
	ErrTooShort = errors.New("Password is too short")
	// ErrTooFewChars gets returned when the password does not contain enough unique characters
	ErrTooFewChars = errors.New("Password does not contain enough different/unique characters")
	// ErrTooSystematic gets returned when the password is too systematic (e.g. 123456, abcdef)
	ErrTooSystematic = errors.New("Password is too systematic")
	// ErrDictionary gets returned when the password is found in a dictionary
	ErrDictionary = errors.New("Password is too common / from a dictionary")
	// ErrMangledDictionary gets returned when the password is mangled, but found in a dictionary
	ErrMangledDictionary = errors.New("Password is mangled, but too common / from a dictionary")
	// ErrHashedDictionary gets returned when the password is hashed, but found in a dictionary
	ErrHashedDictionary = errors.New("Password is hashed, but too common / from a dictionary")
)

// Validator is used to setup a new password validator with options and dictionaries
type Validator struct {
	options     Options
	once        sync.Once
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
	// DictionaryPath contains all the dictionaries that will be parsed
	DictionaryPath string // = "/usr/share/dict"
}

// NewValidator returns a new password validator with default settings
func NewValidator() *Validator {
	return NewValidatorWithOpts(Options{
		MinDist:        -1,
		DictionaryPath: "/usr/share/dict"})
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

			// if a word is smaller than the minimum length minus the minimum distance
			// then any collisons would have been rejected by pre-dictionary checks
			if len(nw) >= v.options.MinLength-v.options.MinDist {
				v.words[nw] = struct{}{}
			}

			for _, hasher := range v.options.Hashers {
				v.hashedWords[hashsum(nw, hasher)] = nw
			}
		}
	}
}

// foundInDictionaries returns whether a (mangled) string exists in the indexed dictionaries
func (v *Validator) foundInDictionaries(s string) (string, error) {
	v.once.Do(v.indexDictionaries)

	pw := normalize(s)   // normalized password
	revpw := reverse(pw) // reversed password

	// let's check perfect matches first
	if _, ok := v.words[pw]; ok {
		return pw, ErrDictionary
	}
	if _, ok := v.words[revpw]; ok {
		return revpw, ErrMangledDictionary
	}

	// find hashed dictionary entries
	if _, ok := v.hashedWords[pw]; ok {
		return pw, ErrHashedDictionary
	}

	// find mangled / reversed passwords
	for word := range v.words {
		if dist := smetrics.WagnerFischer(word, pw, 1, 1, 1); dist <= v.options.MinDist {
			// fmt.Printf("%s is too similar to %s: %d\n", pw, word, dist)
			return word, ErrMangledDictionary
		}
		if dist := smetrics.WagnerFischer(word, revpw, 1, 1, 1); dist <= v.options.MinDist {
			// fmt.Printf("Reversed %s (%s) is too similar to %s: %d\n", pw, revpw, word, dist)
			return word, ErrMangledDictionary
		}
	}

	return "", nil
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

	if _, err := v.foundInDictionaries(password); err != nil {
		return err
	}

	return nil
}
