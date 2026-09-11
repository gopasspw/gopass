// Package goeol contains canary tests that fail once the Go version targeted by
// this repository reaches its end of life according to endoflife.date.
//
// The tests are guarded by the "canary" build tag and are not part of the
// regular test suite. They run in their own GitHub Action workflow, see
// ../../.github/workflows/go-eol-canary.yml.
package goeol
