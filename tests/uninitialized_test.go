package tests

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUninitialized(t *testing.T) {
	ts := newTester(t)
	defer ts.teardown()

	commands := []string{
		"",
		"copy",
		"cp",
		"delete",
		"edit",
		"find",
		"generate",
		"grep",
		"insert",
		"list",
		"ls",
		"mount",
		"move",
		"mv",
		"remove",
		"rm",
		"show",
	}

	for _, command := range commands {
		t.Run(command, func(t *testing.T) {
			out, err := ts.run(command)
			assert.Error(t, err)
			assert.Contains(t, out, "password-store is not initialized. Try ")
		})
	}
}
