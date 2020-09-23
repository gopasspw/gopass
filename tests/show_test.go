package tests

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	goldenQr = "\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\n\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\n\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\n\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\n\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[40m  \x1b[0m\x1b[40m  \x1b[0m\x1b[40m  \x1b[0m\x1b[40m  \x1b[0m\x1b[40m  \x1b[0m\x1b[40m  \x1b[0m\x1b[40m  \x1b[0m\x1b[47m  \x1b[0m\x1b[40m  \x1b[0m\x1b[47m  \x1b[0m\x1b[40m  \x1b[0m\x1b[40m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[40m  \x1b[0m\x1b[40m  \x1b[0m\x1b[40m  \x1b[0m\x1b[40m  \x1b[0m\x1b[40m  \x1b[0m\x1b[40m  \x1b[0m\x1b[40m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\n\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[40m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[40m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[40m  \x1b[0m\x1b[47m  \x1b[0m\x1b[40m  \x1b[0m\x1b[40m  \x1b[0m\x1b[47m  \x1b[0m\x1b[40m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[40m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\n\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[40m  \x1b[0m\x1b[47m  \x1b[0m\x1b[40m  \x1b[0m\x1b[40m  \x1b[0m\x1b[40m  \x1b[0m\x1b[47m  \x1b[0m\x1b[40m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[40m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[40m  \x1b[0m\x1b[47m  \x1b[0m\x1b[40m  \x1b[0m\x1b[40m  \x1b[0m\x1b[40m  \x1b[0m\x1b[47m  \x1b[0m\x1b[40m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\n\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[40m  \x1b[0m\x1b[47m  \x1b[0m\x1b[40m  \x1b[0m\x1b[40m  \x1b[0m\x1b[40m  \x1b[0m\x1b[47m  \x1b[0m\x1b[40m  \x1b[0m\x1b[47m  \x1b[0m\x1b[40m  \x1b[0m\x1b[47m  \x1b[0m\x1b[40m  \x1b[0m\x1b[40m  \x1b[0m\x1b[40m  \x1b[0m\x1b[47m  \x1b[0m\x1b[40m  \x1b[0m\x1b[47m  \x1b[0m\x1b[40m  \x1b[0m\x1b[40m  \x1b[0m\x1b[40m  \x1b[0m\x1b[47m  \x1b[0m\x1b[40m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\n\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[40m  \x1b[0m\x1b[47m  \x1b[0m\x1b[40m  \x1b[0m\x1b[40m  \x1b[0m\x1b[40m  \x1b[0m\x1b[47m  \x1b[0m\x1b[40m  \x1b[0m\x1b[47m  \x1b[0m\x1b[40m  \x1b[0m\x1b[47m  \x1b[0m\x1b[40m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[40m  \x1b[0m\x1b[47m  \x1b[0m\x1b[40m  \x1b[0m\x1b[40m  \x1b[0m\x1b[40m  \x1b[0m\x1b[47m  \x1b[0m\x1b[40m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\n\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[40m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[40m  \x1b[0m\x1b[47m  \x1b[0m\x1b[40m  \x1b[0m\x1b[40m  \x1b[0m\x1b[47m  \x1b[0m\x1b[40m  \x1b[0m\x1b[40m  \x1b[0m\x1b[47m  \x1b[0m\x1b[40m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[40m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\n\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[40m  \x1b[0m\x1b[40m  \x1b[0m\x1b[40m  \x1b[0m\x1b[40m  \x1b[0m\x1b[40m  \x1b[0m\x1b[40m  \x1b[0m\x1b[40m  \x1b[0m\x1b[47m  \x1b[0m\x1b[40m  \x1b[0m\x1b[47m  \x1b[0m\x1b[40m  \x1b[0m\x1b[47m  \x1b[0m\x1b[40m  \x1b[0m\x1b[47m  \x1b[0m\x1b[40m  \x1b[0m\x1b[40m  \x1b[0m\x1b[40m  \x1b[0m\x1b[40m  \x1b[0m\x1b[40m  \x1b[0m\x1b[40m  \x1b[0m\x1b[40m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\n\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[40m  \x1b[0m\x1b[40m  \x1b[0m\x1b[40m  \x1b[0m\x1b[40m  \x1b[0m\x1b[40m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\n\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[40m  \x1b[0m\x1b[40m  \x1b[0m\x1b[40m  \x1b[0m\x1b[47m  \x1b[0m\x1b[40m  \x1b[0m\x1b[47m  \x1b[0m\x1b[40m  \x1b[0m\x1b[47m  \x1b[0m\x1b[40m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[40m  \x1b[0m\x1b[40m  \x1b[0m\x1b[40m  \x1b[0m\x1b[40m  \x1b[0m\x1b[40m  \x1b[0m\x1b[40m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[40m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\n\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[40m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[40m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[40m  \x1b[0m\x1b[40m  \x1b[0m\x1b[40m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[40m  \x1b[0m\x1b[40m  \x1b[0m\x1b[47m  \x1b[0m\x1b[40m  \x1b[0m\x1b[47m  \x1b[0m\x1b[40m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\n\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[40m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[40m  \x1b[0m\x1b[47m  \x1b[0m\x1b[40m  \x1b[0m\x1b[40m  \x1b[0m\x1b[40m  \x1b[0m\x1b[40m  \x1b[0m\x1b[40m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[40m  \x1b[0m\x1b[40m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[40m  \x1b[0m\x1b[40m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\n\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[40m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[40m  \x1b[0m\x1b[47m  \x1b[0m\x1b[40m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[40m  \x1b[0m\x1b[47m  \x1b[0m\x1b[40m  \x1b[0m\x1b[47m  \x1b[0m\x1b[40m  \x1b[0m\x1b[40m  \x1b[0m\x1b[40m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[40m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\n\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[40m  \x1b[0m\x1b[40m  \x1b[0m\x1b[40m  \x1b[0m\x1b[40m  \x1b[0m\x1b[40m  \x1b[0m\x1b[40m  \x1b[0m\x1b[40m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[40m  \x1b[0m\x1b[40m  \x1b[0m\x1b[40m  \x1b[0m\x1b[40m  \x1b[0m\x1b[40m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\n\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[40m  \x1b[0m\x1b[47m  \x1b[0m\x1b[40m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[40m  \x1b[0m\x1b[40m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[40m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\n\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[40m  \x1b[0m\x1b[40m  \x1b[0m\x1b[40m  \x1b[0m\x1b[40m  \x1b[0m\x1b[40m  \x1b[0m\x1b[40m  \x1b[0m\x1b[40m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[40m  \x1b[0m\x1b[40m  \x1b[0m\x1b[47m  \x1b[0m\x1b[40m  \x1b[0m\x1b[47m  \x1b[0m\x1b[40m  \x1b[0m\x1b[47m  \x1b[0m\x1b[40m  \x1b[0m\x1b[47m  \x1b[0m\x1b[40m  \x1b[0m\x1b[40m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\n\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[40m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[40m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[40m  \x1b[0m\x1b[40m  \x1b[0m\x1b[40m  \x1b[0m\x1b[40m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\n\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[40m  \x1b[0m\x1b[47m  \x1b[0m\x1b[40m  \x1b[0m\x1b[40m  \x1b[0m\x1b[40m  \x1b[0m\x1b[47m  \x1b[0m\x1b[40m  \x1b[0m\x1b[47m  \x1b[0m\x1b[40m  \x1b[0m\x1b[40m  \x1b[0m\x1b[40m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[40m  \x1b[0m\x1b[40m  \x1b[0m\x1b[47m  \x1b[0m\x1b[40m  \x1b[0m\x1b[47m  \x1b[0m\x1b[40m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\n\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[40m  \x1b[0m\x1b[47m  \x1b[0m\x1b[40m  \x1b[0m\x1b[40m  \x1b[0m\x1b[40m  \x1b[0m\x1b[47m  \x1b[0m\x1b[40m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[40m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[40m  \x1b[0m\x1b[47m  \x1b[0m\x1b[40m  \x1b[0m\x1b[40m  \x1b[0m\x1b[47m  \x1b[0m\x1b[40m  \x1b[0m\x1b[40m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\n\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[40m  \x1b[0m\x1b[47m  \x1b[0m\x1b[40m  \x1b[0m\x1b[40m  \x1b[0m\x1b[40m  \x1b[0m\x1b[47m  \x1b[0m\x1b[40m  \x1b[0m\x1b[47m  \x1b[0m\x1b[40m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[40m  \x1b[0m\x1b[47m  \x1b[0m\x1b[40m  \x1b[0m\x1b[40m  \x1b[0m\x1b[40m  \x1b[0m\x1b[47m  \x1b[0m\x1b[40m  \x1b[0m\x1b[47m  \x1b[0m\x1b[40m  \x1b[0m\x1b[40m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\n\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[40m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[40m  \x1b[0m\x1b[47m  \x1b[0m\x1b[40m  \x1b[0m\x1b[47m  \x1b[0m\x1b[40m  \x1b[0m\x1b[40m  \x1b[0m\x1b[47m  \x1b[0m\x1b[40m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[40m  \x1b[0m\x1b[47m  \x1b[0m\x1b[40m  \x1b[0m\x1b[40m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\n\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[40m  \x1b[0m\x1b[40m  \x1b[0m\x1b[40m  \x1b[0m\x1b[40m  \x1b[0m\x1b[40m  \x1b[0m\x1b[40m  \x1b[0m\x1b[40m  \x1b[0m\x1b[47m  \x1b[0m\x1b[40m  \x1b[0m\x1b[47m  \x1b[0m\x1b[40m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[40m  \x1b[0m\x1b[40m  \x1b[0m\x1b[47m  \x1b[0m\x1b[40m  \x1b[0m\x1b[40m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\n\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\n\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\n\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\n\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m"
)

func TestShow(t *testing.T) {
	ts := newTester(t)
	defer ts.teardown()

	_, err := ts.run("show")
	assert.Error(t, err)

	ts.initStore()

	t.Run("test usage", func(t *testing.T) {
		out, err := ts.run("show")
		assert.Error(t, err)
		assert.Equal(t, "\nError: Usage: "+filepath.Base(ts.Binary)+" show [name]\n", out)
	})

	t.Run("test show with non-existing secret", func(t *testing.T) {
		out, err := ts.run("show foo")
		assert.Error(t, err)
		assert.Equal(t, "\nError: failed to retrieve secret 'foo': Entry is not in the password store\n", out)
	})

	ts.initSecrets("")

	t.Run("show folder foo", func(t *testing.T) {
		_, err = ts.run("show foo")
		assert.NoError(t, err)
		_, err = ts.run("show -u foo")
		assert.NoError(t, err)
		_, err = ts.run("show foo -unsafe")
		assert.NoError(t, err)
	})

	t.Run("show w/o safecontent", func(t *testing.T) {
		_, err = ts.run("config safecontent false")
		assert.NoError(t, err)

		out, err := ts.run("show fixed/secret")
		assert.NoError(t, err)
		assert.Equal(t, "Password: moar", out)

		out, err = ts.run("show fixed/twoliner")
		assert.NoError(t, err)
		assert.Equal(t, "Password: and\n\nmore stuff", out)

		out, err = ts.run("show --qr fixed/secret")
		assert.NoError(t, err)
		assert.Equal(t, goldenQr, out)
	})

	t.Run("show w/o autoclip", func(t *testing.T) {
		_, err = ts.run("config autoclip false")
		assert.NoError(t, err)
		_, err = ts.run("show fixed/secret")
		assert.NoError(t, err)
	})

	t.Run("show with safecontent", func(t *testing.T) {
		_, err = ts.run("config safecontent true")
		assert.NoError(t, err)
		_, err = ts.run("config mime true")
		assert.NoError(t, err)

		out, err := ts.run("show fixed/secret")
		assert.NoError(t, err)
		assert.Contains(t, out, "Password: ***")

		out, err = ts.run("show fixed/twoliner password")
		assert.NoError(t, err)
		assert.Equal(t, "and", out)

		out, err = ts.run("show fixed/twoliner")
		assert.NoError(t, err)
		assert.Contains(t, out, "more stuff")
		assert.NotContains(t, out, "and")
	})

	t.Run("force showing full secret", func(t *testing.T) {
		_, err = ts.run("config safecontent true")
		assert.NoError(t, err)
		_, err = ts.run("config mime true")
		assert.NoError(t, err)

		out, err := ts.run("show -u fixed/secret")
		assert.NoError(t, err)
		assert.Equal(t, "Password: moar", out)

		out, err = ts.run("show -o fixed/secret")
		assert.NoError(t, err)
		assert.Equal(t, "moar", out)

		out, err = ts.run("show -u fixed/twoliner")
		assert.NoError(t, err)
		assert.Equal(t, "Password: and\n\nmore stuff", out)

		out, err = ts.run("show -o fixed/twoliner")
		assert.NoError(t, err)
		assert.Equal(t, "and", out)

		out, err = ts.run("show -c fixed/twoliner")
		assert.NoError(t, err)
		assert.NotContains(t, out, "***")
		assert.NotContains(t, out, "and")

		out, err = ts.run("show -C fixed/twoliner")
		assert.NoError(t, err)
		assert.Contains(t, out, "***")
		assert.NotContains(t, out, "and")
	})

	// what if we disable the new secret type after having been using it?
	t.Run("after disabling mime", func(t *testing.T) {
		_, err = ts.run("config mime false")
		assert.NoError(t, err)
		_, err = ts.run("config safecontent true")
		assert.NoError(t, err)

		// No flags
		out, err := ts.run("show fixed/secret")
		assert.NoError(t, err)
		assert.Contains(t, out, "***")
		assert.NotContains(t, out, "Warning: safecontent=true")

		out, err = ts.run("show fixed/twoliner password")
		assert.NoError(t, err)
		assert.Equal(t, out, "and")

		out, err = ts.run("show fixed/twoliner")
		assert.NoError(t, err)
		assert.Contains(t, out, "more stuff")
		assert.NotContains(t, out, "and")

		// with flags now
		out, err = ts.run("show -o fixed/secret")
		assert.NoError(t, err)
		assert.Equal(t, "moar", out)

		out, err = ts.run("show -u fixed/secret")
		assert.NoError(t, err)
		assert.Equal(t, "moar", out)

		out, err = ts.run("show -o fixed/twoliner")
		assert.NoError(t, err)
		assert.Equal(t, "and", out)

		out, err = ts.run("show -u fixed/twoliner")
		assert.NoError(t, err)
		assert.Equal(t, "and\n\nmore stuff", out)
	})

	t.Run("Regression test for #1574 and #1575", func(t *testing.T) {
		_, err = ts.run("config safecontent true")
		assert.NoError(t, err)
		_, err = ts.run("config mime false")
		assert.NoError(t, err)

		_, err := ts.run("generate fo2 5")
		assert.NoError(t, err)

		out, err := ts.run("show fo2")
		assert.Error(t, err)
		assert.Contains(t, out, "Warning: safecontent=true")

		out, err = ts.run("show -u fo2")
		assert.NoError(t, err)
		assert.Equal(t, out, "aaaaa")

		_, err = ts.run("config mime true")
		assert.NoError(t, err)

		_, err = ts.run("generate fo6 5")
		assert.NoError(t, err)

		out, err = ts.run("show fo6")
		assert.NoError(t, err)
		assert.Contains(t, out, "Password: ***")
		assert.NotContains(t, out, "\n\n")

		out, err = ts.run("show -u fo6")
		assert.NoError(t, err)
		assert.Equal(t, out, "Password: aaaaa")
		assert.NotContains(t, out, "\n\n")
	})
}
