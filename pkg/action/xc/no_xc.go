// +build !xc

package xc

import (
	"github.com/urfave/cli/v2"
)

// GetCommands returns the cli commands provided by this module
func GetCommands() []*cli.Command { return nil }
