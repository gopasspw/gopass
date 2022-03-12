package debug

import (
	"os"
	"testing"
)

func BenchmarkLogging(b *testing.B) {
	b.Setenv("GOPASS_DEBUG", "true")

	initDebug()

	for i := 0; i < b.N; i++ {
		Log("string")
	}
}

func BenchmarkNoLogging(b *testing.B) {
	_ = os.Unsetenv("GOPASS_DEBUG")

	initDebug()

	for i := 0; i < b.N; i++ {
		Log("string")
	}
}
