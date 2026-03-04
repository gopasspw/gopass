package audit

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFinalize(t *testing.T) {
	r := newReport()
	r.AddPassword("foo", "bar")
	r.AddPassword("baz", "bar")
	r.AddPassword("zab", "bar")
	r.AddPassword("foo", "bar")
	r.AddFinding("foo", "foo", "bar", "warning")
	r.AddFinding("bar", "foo", "bar", "warning")

	sr := r.Finalize()
	assert.NotNil(t, sr)
}
