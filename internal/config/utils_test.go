package config

import (
	"testing"
)

func TestAsBoolWithDefault(t *testing.T) {
	tests := []struct {
		name     string
		s        string
		def      bool
		expected bool
	}{
		{
			name:     "Empty string with default true",
			s:        "",
			def:      true,
			expected: true,
		},
		{
			name:     "Empty string with default false",
			s:        "",
			def:      false,
			expected: false,
		},
		{
			name:     "Valid string '1' with default true",
			s:        "1",
			def:      true,
			expected: true,
		},
		{
			name:     "Valid string '0' with default true",
			s:        "0",
			def:      true,
			expected: false,
		},
		// Add more test cases here
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := AsBoolWithDefault(test.s, test.def)
			if result != test.expected {
				t.Errorf("Expected %v, but got %v", test.expected, result)
			}
		})
	}
}

func TestAsIntWithDefault(t *testing.T) {
	tests := []struct {
		name     string
		s        string
		def      int
		expected int
	}{
		{
			name:     "Empty string with default 0",
			s:        "",
			def:      0,
			expected: 0,
		},
		{
			name:     "Valid string '123' with default 0",
			s:        "123",
			def:      0,
			expected: 123,
		},
		{
			name:     "Invalid string 'abc' with default 0",
			s:        "abc",
			def:      0,
			expected: 0,
		},
		// Add more test cases here
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := AsIntWithDefault(test.s, test.def)
			if result != test.expected {
				t.Errorf("Expected %v, but got %v", test.expected, result)
			}
		})
	}
}
