package debug

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

// can not import out.Secret.
type testSecret string

func (t testSecret) SafeStr() string {
	return "(elided)"
}

type testShort string

func (t testShort) Str() string {
	return "shorter"
}

func TestDebug(t *testing.T) {
	td := t.TempDir()
	t.Cleanup(func() {
		initDebug()
	})

	fn := filepath.Join(td, "gopass.log")
	t.Setenv("GOPASS_DEBUG_LOG", fn)

	// it's been already initialized, need to re-init
	assert.True(t, initDebug())

	Log("foo")
	Log("%s", testSecret("secret"))
	Log("%s", testShort("toolong"))

	buf, err := os.ReadFile(fn)
	require.NoError(t, err)

	logStr := string(buf)
	assert.Contains(t, logStr, "foo")
	assert.NotContains(t, logStr, "secret")
	assert.NotContains(t, logStr, "toolong")
	assert.Contains(t, logStr, "shorter")
}

func TestDebugSecret(t *testing.T) {
	td := t.TempDir()
	t.Cleanup(func() {
		initDebug()
	})

	fn := filepath.Join(td, "gopass.log")
	t.Setenv("GOPASS_DEBUG_LOG", fn)
	t.Setenv("GOPASS_DEBUG_LOG_SECRETS", "true")

	// it's been already initialized, need to re-init
	assert.True(t, initDebug())

	assert.True(t, logSecrets)

	Log("foo")
	Log("%s", testSecret("secret"))

	buf, err := os.ReadFile(fn)
	require.NoError(t, err)

	logStr := string(buf)
	assert.Contains(t, logStr, "foo")
	assert.Contains(t, logStr, "secret")
}

func TestDebugFilter(t *testing.T) {
	td := t.TempDir()
	t.Cleanup(func() {
		initDebug()
	})

	fn := filepath.Join(td, "gopass.log")
	t.Setenv("GOPASS_DEBUG_LOG", fn)
	t.Setenv("GOPASS_DEBUG_FUNCS", "TestDebugFilter")
	t.Setenv("GOPASS_DEBUG_FILES", "debug_test.go")

	buf := &bytes.Buffer{}
	Stderr = buf
	defer func() {
		Stderr = os.Stderr
	}()

	// it's been already initialized, need to re-init
	assert.True(t, initDebug())

	Log("foo")
	Log("%s", testSecret("secret"))

	fbuf, err := os.ReadFile(fn)
	require.NoError(t, err)

	logStr := string(fbuf)
	assert.Contains(t, logStr, "foo")

	stderrStr := buf.String()
	assert.Contains(t, stderrStr, "TestDebugFilter")
}
