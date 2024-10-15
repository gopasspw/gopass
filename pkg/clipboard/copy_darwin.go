//go:build darwin
// +build darwin

package clipboard

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"

	"github.com/atotto/clipboard"
	"github.com/gopasspw/gopass/pkg/debug"
)

func copyToClipboard(ctx context.Context, content []byte) error {
	password := string(content)

	// first try to copy via osascript, if that fails fallback to pbcopy
	if err := copyViaOsascript(ctx, password); err != nil {
		debug.Log("failed to copy via osascript: %s", err)

		return clipboard.WriteAll(password)
	}

	return nil
}

func copyViaOsascript(ctx context.Context, password string) error {
	args := []string{
		// The Foundation library has the Objective C symbols for pasteboard
		"-e", `use framework "Foundation"`,
		// Need to use scripting additions for access to "do shell script"
		"-e", "use scripting additions",
		// type = a reference to the ObjC constant NSPasteboardTypeString
		// which is needed to indentify clioboard contents as text
		"-e", "set type to current application's NSPasteboardTypeString",
		// pb = a reference to the system's pasteboard
		"-e", "set pb to current application's NSPasteboard's generalPasteboard()",
		// Must clear contents before adding a new item to pasteboard
		"-e", "pb's clearContents()",
		// Set the flag ConcealedType so clipboard history managers don't record the password.
		// The first argument can by anything, but an empty string will do fine.
		"-e", `pb's setString:"" forType:"org.nspasteboard.ConcealedType"`,
		// AppleScript cannot read from stdin, so pipe fd#3 to stdin of cat and read the output.
		// This output is put in the clipboard, setting type = string type
		"-e", `pb's setString:(do shell script "cat 0<&3") forType:type`,
	}
	cmd := exec.CommandContext(ctx, "osascript", args...)
	r, w, err := os.Pipe()
	if err != nil {
		return err
	}

	// This connects the pipe to stdin of the osascript command, see the "do shell script"
	// part around line 46. The pipe is created and written to before the osascript command
	// is run so we shouldn't need to worry about partial writes (I hope!).
	//
	// TODO: We might be able to use `cmd.Stdin = strings.NewReader(password)` instead
	cmd.ExtraFiles = []*os.File{r} // Receiving end of pipes is connected to fd#3
	go func() {
		defer w.Close()                    //nolint:errcheck
		_, _ = io.WriteString(w, password) // Write the password to fd#3
	}()

	out, err := cmd.Output()
	if err != nil {
		return err
	}

	// osascript should print true (return value of the last setString call) on success
	if string(out) != "true\n" {
		// Fallback to using attoto's pbcopy
		return fmt.Errorf("osascript failed to set password: %s", string(out))
	}
	debug.Log("copied via osascript")

	return nil
}
