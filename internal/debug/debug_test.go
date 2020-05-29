package debug

import (
	"os"
	"testing"
)

func BenchmarkLogging(b *testing.B) {
	os.Setenv("GOPASS_DEBUG", "true")
	defer func() { os.Unsetenv("GOPASS_DEBUG") }()
	initDebug()

	for i := 0; i < b.N; i++ {
		Log("string")
	}
}

func BenchmarkNoLogging(b *testing.B) {
	os.Unsetenv("GOPASS_DEBUG")
	initDebug()

	for i := 0; i < b.N; i++ {
		Log("string")
	}
}
