//go:build !windows
// +build !windows

package main

import (
	"os/signal"
	"syscall"
)

func init() {
	// workaround for https://github.com/golang/go/issues/37942
	signal.Ignore(syscall.SIGURG)
}
