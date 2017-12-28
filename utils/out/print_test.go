package out

import (
	"bytes"
	"context"
	"os"
	"testing"
)

func TestPrint(t *testing.T) {
	ctx := context.Background()
	buf := &bytes.Buffer{}
	Stdout = buf
	defer func() {
		Stdout = os.Stdout
	}()

	Print(ctx, "%s = %d", "foo", 42)
	if buf.String() != "foo = 42\n" {
		t.Errorf("Wrong output: %s", buf.String())
	}
	buf.Reset()

	Print(WithNewline(ctx, false), "%s = %d", "foo", 42)
	if buf.String() != "foo = 42" {
		t.Errorf("Wrong output: %s", buf.String())
	}
	buf.Reset()
}

func TestDebug(t *testing.T) {
	ctx := context.Background()
	buf := &bytes.Buffer{}
	Stdout = buf
	defer func() {
		Stdout = os.Stdout
	}()

	Debug(ctx, "foobar")
	if buf.String() != "" {
		t.Errorf("Got output: %s", buf.String())
	}
}
