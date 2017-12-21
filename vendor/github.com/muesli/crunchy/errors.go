/*
 * crunchy - find common flaws in passwords
 *     Copyright (c) 2017, Christian Muehlhaeuser <muesli@gmail.com>
 *
 *   For license see LICENSE
 */

package crunchy

import (
	"errors"
)

// DictionaryError wraps an ErrMangledDictionary with contextual information
type DictionaryError struct {
	Err      error
	Word     string
	Distance int
}

// HashedDictionaryError wraps an ErrHashedDictionary with contextual information
type HashedDictionaryError struct {
	Err  error
	Word string
}

func (e *DictionaryError) Error() string {
	return e.Err.Error()
}

func (e *HashedDictionaryError) Error() string {
	return e.Err.Error()
}

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
