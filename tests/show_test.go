package tests

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var goldenQr = "\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\n\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\n\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\n\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\n\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[40m  \x1b[0m\x1b[40m  \x1b[0m\x1b[40m  \x1b[0m\x1b[40m  \x1b[0m\x1b[40m  \x1b[0m\x1b[40m  \x1b[0m\x1b[40m  \x1b[0m\x1b[47m  \x1b[0m\x1b[40m  \x1b[0m\x1b[47m  \x1b[0m\x1b[40m  \x1b[0m\x1b[40m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[40m  \x1b[0m\x1b[40m  \x1b[0m\x1b[40m  \x1b[0m\x1b[40m  \x1b[0m\x1b[40m  \x1b[0m\x1b[40m  \x1b[0m\x1b[40m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\n\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[40m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[40m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[40m  \x1b[0m\x1b[47m  \x1b[0m\x1b[40m  \x1b[0m\x1b[40m  \x1b[0m\x1b[47m  \x1b[0m\x1b[40m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[40m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\n\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[40m  \x1b[0m\x1b[47m  \x1b[0m\x1b[40m  \x1b[0m\x1b[40m  \x1b[0m\x1b[40m  \x1b[0m\x1b[47m  \x1b[0m\x1b[40m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[40m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[40m  \x1b[0m\x1b[47m  \x1b[0m\x1b[40m  \x1b[0m\x1b[40m  \x1b[0m\x1b[40m  \x1b[0m\x1b[47m  \x1b[0m\x1b[40m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\n\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[40m  \x1b[0m\x1b[47m  \x1b[0m\x1b[40m  \x1b[0m\x1b[40m  \x1b[0m\x1b[40m  \x1b[0m\x1b[47m  \x1b[0m\x1b[40m  \x1b[0m\x1b[47m  \x1b[0m\x1b[40m  \x1b[0m\x1b[47m  \x1b[0m\x1b[40m  \x1b[0m\x1b[40m  \x1b[0m\x1b[40m  \x1b[0m\x1b[47m  \x1b[0m\x1b[40m  \x1b[0m\x1b[47m  \x1b[0m\x1b[40m  \x1b[0m\x1b[40m  \x1b[0m\x1b[40m  \x1b[0m\x1b[47m  \x1b[0m\x1b[40m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\n\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[40m  \x1b[0m\x1b[47m  \x1b[0m\x1b[40m  \x1b[0m\x1b[40m  \x1b[0m\x1b[40m  \x1b[0m\x1b[47m  \x1b[0m\x1b[40m  \x1b[0m\x1b[47m  \x1b[0m\x1b[40m  \x1b[0m\x1b[47m  \x1b[0m\x1b[40m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[40m  \x1b[0m\x1b[47m  \x1b[0m\x1b[40m  \x1b[0m\x1b[40m  \x1b[0m\x1b[40m  \x1b[0m\x1b[47m  \x1b[0m\x1b[40m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\n\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[40m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[40m  \x1b[0m\x1b[47m  \x1b[0m\x1b[40m  \x1b[0m\x1b[40m  \x1b[0m\x1b[47m  \x1b[0m\x1b[40m  \x1b[0m\x1b[40m  \x1b[0m\x1b[47m  \x1b[0m\x1b[40m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[40m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\n\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[40m  \x1b[0m\x1b[40m  \x1b[0m\x1b[40m  \x1b[0m\x1b[40m  \x1b[0m\x1b[40m  \x1b[0m\x1b[40m  \x1b[0m\x1b[40m  \x1b[0m\x1b[47m  \x1b[0m\x1b[40m  \x1b[0m\x1b[47m  \x1b[0m\x1b[40m  \x1b[0m\x1b[47m  \x1b[0m\x1b[40m  \x1b[0m\x1b[47m  \x1b[0m\x1b[40m  \x1b[0m\x1b[40m  \x1b[0m\x1b[40m  \x1b[0m\x1b[40m  \x1b[0m\x1b[40m  \x1b[0m\x1b[40m  \x1b[0m\x1b[40m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\n\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[40m  \x1b[0m\x1b[40m  \x1b[0m\x1b[40m  \x1b[0m\x1b[40m  \x1b[0m\x1b[40m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\n\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[40m  \x1b[0m\x1b[40m  \x1b[0m\x1b[40m  \x1b[0m\x1b[47m  \x1b[0m\x1b[40m  \x1b[0m\x1b[47m  \x1b[0m\x1b[40m  \x1b[0m\x1b[47m  \x1b[0m\x1b[40m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[40m  \x1b[0m\x1b[40m  \x1b[0m\x1b[40m  \x1b[0m\x1b[40m  \x1b[0m\x1b[40m  \x1b[0m\x1b[40m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[40m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\n\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[40m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[40m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[40m  \x1b[0m\x1b[40m  \x1b[0m\x1b[40m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[40m  \x1b[0m\x1b[40m  \x1b[0m\x1b[47m  \x1b[0m\x1b[40m  \x1b[0m\x1b[47m  \x1b[0m\x1b[40m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\n\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[40m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[40m  \x1b[0m\x1b[47m  \x1b[0m\x1b[40m  \x1b[0m\x1b[40m  \x1b[0m\x1b[40m  \x1b[0m\x1b[40m  \x1b[0m\x1b[40m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[40m  \x1b[0m\x1b[40m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[40m  \x1b[0m\x1b[40m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\n\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[40m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[40m  \x1b[0m\x1b[47m  \x1b[0m\x1b[40m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[40m  \x1b[0m\x1b[47m  \x1b[0m\x1b[40m  \x1b[0m\x1b[47m  \x1b[0m\x1b[40m  \x1b[0m\x1b[40m  \x1b[0m\x1b[40m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[40m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\n\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[40m  \x1b[0m\x1b[40m  \x1b[0m\x1b[40m  \x1b[0m\x1b[40m  \x1b[0m\x1b[40m  \x1b[0m\x1b[40m  \x1b[0m\x1b[40m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[40m  \x1b[0m\x1b[40m  \x1b[0m\x1b[40m  \x1b[0m\x1b[40m  \x1b[0m\x1b[40m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\n\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[40m  \x1b[0m\x1b[47m  \x1b[0m\x1b[40m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[40m  \x1b[0m\x1b[40m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[40m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\n\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[40m  \x1b[0m\x1b[40m  \x1b[0m\x1b[40m  \x1b[0m\x1b[40m  \x1b[0m\x1b[40m  \x1b[0m\x1b[40m  \x1b[0m\x1b[40m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[40m  \x1b[0m\x1b[40m  \x1b[0m\x1b[47m  \x1b[0m\x1b[40m  \x1b[0m\x1b[47m  \x1b[0m\x1b[40m  \x1b[0m\x1b[47m  \x1b[0m\x1b[40m  \x1b[0m\x1b[47m  \x1b[0m\x1b[40m  \x1b[0m\x1b[40m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\n\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[40m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[40m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[40m  \x1b[0m\x1b[40m  \x1b[0m\x1b[40m  \x1b[0m\x1b[40m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\n\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[40m  \x1b[0m\x1b[47m  \x1b[0m\x1b[40m  \x1b[0m\x1b[40m  \x1b[0m\x1b[40m  \x1b[0m\x1b[47m  \x1b[0m\x1b[40m  \x1b[0m\x1b[47m  \x1b[0m\x1b[40m  \x1b[0m\x1b[40m  \x1b[0m\x1b[40m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[40m  \x1b[0m\x1b[40m  \x1b[0m\x1b[47m  \x1b[0m\x1b[40m  \x1b[0m\x1b[47m  \x1b[0m\x1b[40m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\n\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[40m  \x1b[0m\x1b[47m  \x1b[0m\x1b[40m  \x1b[0m\x1b[40m  \x1b[0m\x1b[40m  \x1b[0m\x1b[47m  \x1b[0m\x1b[40m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[40m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[40m  \x1b[0m\x1b[47m  \x1b[0m\x1b[40m  \x1b[0m\x1b[40m  \x1b[0m\x1b[47m  \x1b[0m\x1b[40m  \x1b[0m\x1b[40m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\n\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[40m  \x1b[0m\x1b[47m  \x1b[0m\x1b[40m  \x1b[0m\x1b[40m  \x1b[0m\x1b[40m  \x1b[0m\x1b[47m  \x1b[0m\x1b[40m  \x1b[0m\x1b[47m  \x1b[0m\x1b[40m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[40m  \x1b[0m\x1b[47m  \x1b[0m\x1b[40m  \x1b[0m\x1b[40m  \x1b[0m\x1b[40m  \x1b[0m\x1b[47m  \x1b[0m\x1b[40m  \x1b[0m\x1b[47m  \x1b[0m\x1b[40m  \x1b[0m\x1b[40m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\n\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[40m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[40m  \x1b[0m\x1b[47m  \x1b[0m\x1b[40m  \x1b[0m\x1b[47m  \x1b[0m\x1b[40m  \x1b[0m\x1b[40m  \x1b[0m\x1b[47m  \x1b[0m\x1b[40m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[40m  \x1b[0m\x1b[47m  \x1b[0m\x1b[40m  \x1b[0m\x1b[40m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\n\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[40m  \x1b[0m\x1b[40m  \x1b[0m\x1b[40m  \x1b[0m\x1b[40m  \x1b[0m\x1b[40m  \x1b[0m\x1b[40m  \x1b[0m\x1b[40m  \x1b[0m\x1b[47m  \x1b[0m\x1b[40m  \x1b[0m\x1b[47m  \x1b[0m\x1b[40m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[40m  \x1b[0m\x1b[40m  \x1b[0m\x1b[47m  \x1b[0m\x1b[40m  \x1b[0m\x1b[40m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\n\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\n\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\n\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\n\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m\x1b[47m  \x1b[0m"

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
		assert.Contains(t, out, "entry is not in the password store", out)
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
		_, err = ts.run("config core.showsafecontent false")
		assert.NoError(t, err)

		out, err := ts.run("show fixed/secret")
		assert.NoError(t, err)
		assert.Equal(t, "moar", out)

		out, err = ts.run("show fixed/twoliner")
		assert.NoError(t, err)
		assert.Equal(t, "first line\nsecond line", out)

		out, err = ts.run("show --qr fixed/secret")
		assert.NoError(t, err)
		assert.Equal(t, goldenQr, out)
	})

	t.Run("show w/o autoclip", func(t *testing.T) {
		_, err = ts.run("config core.autoclip false")
		assert.NoError(t, err)
		_, err = ts.run("show fixed/secret")
		assert.NoError(t, err)
	})

	t.Run("show with safecontent", func(t *testing.T) {
		_, err = ts.run("config core.showsafecontent true")
		require.NoError(t, err, "set core.showsafecontent = true")

		out, err := ts.run("config core.showsafecontent")
		require.NoError(t, err)
		assert.Contains(t, out, "true", "verify core.showsafecontent = true")

		out, err = ts.run("show fixed/secret")
		assert.Error(t, err)
		assert.Contains(t, out, "safecontent", "output should contain a safecontent warning")

		out, err = ts.run("show fixed/twoliner")
		require.NoError(t, err)
		assert.NotContains(t, out, "password: ***")
		assert.Contains(t, out, "second line")
		assert.NotContains(t, out, "first line", "safecontent = true should remove the first (password) line")
	})

	t.Run("force showing full secret", func(t *testing.T) {
		_, err = ts.run("config core.showsafecontent true")
		assert.NoError(t, err)

		out, err := ts.run("show -u fixed/secret")
		assert.NoError(t, err)
		assert.Equal(t, "moar", out)

		out, err = ts.run("show -o fixed/secret")
		assert.NoError(t, err)
		assert.Equal(t, "moar", out)

		out, err = ts.run("show -u fixed/twoliner")
		assert.NoError(t, err)
		assert.Equal(t, "first line\nsecond line", out)

		out, err = ts.run("show -o fixed/twoliner")
		assert.NoError(t, err)
		assert.Equal(t, "first line", out)

		out, err = ts.run("show -c fixed/twoliner")
		assert.NoError(t, err)
		assert.NotContains(t, out, "***")
		assert.NotContains(t, out, "safecontent=true")
		assert.NotContains(t, out, "first line")
		assert.NotContains(t, out, "second line")

		out, err = ts.run("show -C fixed/twoliner")
		assert.NoError(t, err)
		assert.Contains(t, out, "second line")
		assert.NotContains(t, out, "first line")
	})

	t.Run("Regression test for #1574 and #1575", func(t *testing.T) {
		t.Setenv("GOPASS_CHARACTER_SET", "a")

		_, err = ts.run("config core.showsafecontent true")
		assert.NoError(t, err)

		_, err := ts.run("generate fo2 5")
		assert.NoError(t, err)

		out, err := ts.run("show fo2")
		assert.Error(t, err)
		assert.NotContains(t, out, "password: *****")
		assert.NotContains(t, out, "aaaaa")
		assert.Contains(t, out, "safecontent=true")

		out, err = ts.run("show -u fo2")
		assert.NoError(t, err)
		assert.Equal(t, "aaaaa", out)

		_, err = ts.run("generate fo6 5")
		assert.NoError(t, err)

		out, err = ts.run("show fo6")
		assert.Error(t, err)
		assert.NotContains(t, out, "password: ***")
		assert.NotContains(t, out, "aaaaa")
		assert.Contains(t, out, "safecontent=true")

		out, err = ts.run("show -u fo6")
		assert.NoError(t, err)
		assert.Equal(t, "aaaaa", out)
		assert.NotContains(t, out, "\n\n")
	})
}
