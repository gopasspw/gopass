package termio

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/gopasspw/gopass/pkg/debug"
)

// SortListResult is the outcome of a SortList interaction.
type SortListResult int

const (
	// SortListSaved means the user confirmed the new order.
	SortListSaved SortListResult = iota
	// SortListAborted means the user aborted without saving.
	SortListAborted
)

// sortListKeys describes the key bindings shown to the user.
const sortListKeys = "j/k or ↑/↓: select · u or Ctrl+u: move up · d or Ctrl+d: move down · s: save · q: abort"

// SortList presents an interactive, keyboard-driven list editor on the
// terminal. The user can move the cursor, reorder entries and finally save
// or abort. It returns the (possibly reordered) list and the result.
//
// This is intentionally built only on the primitives already available in
// this package (and golang.org/x/term for raw mode) to avoid pulling in any
// external TUI dependency.
func SortList(ctx context.Context, title string, items []string) ([]string, SortListResult, error) {
	if ctxutil.IsAlwaysYes(ctx) || !ctxutil.IsInteractive(ctx) {
		return items, SortListAborted, nil
	}

	if !ctxutil.IsTerminal(ctx) {
		// fall back to a non-interactive no-op
		return items, SortListAborted, nil
	}

	l := &sortList{
		title: title,
		items: slicesClone(items),
	}

	if err := l.run(ctx); err != nil {
		return items, SortListAborted, err
	}

	if l.aborted {
		return items, SortListAborted, nil
	}

	return l.items, SortListSaved, nil
}

type sortList struct {
	title   string
	items   []string
	cursor  int
	aborted bool
}

func (s *sortList) run(ctx context.Context) error {
	oldState, err := makeRaw()
	if err != nil {
		return fmt.Errorf("failed to enter raw mode: %w", err)
	}
	defer func() { _ = restore(oldState) }()

	s.render()

	for {
		select {
		case <-ctx.Done():
			s.aborted = true

			return nil
		default:
		}

		key, err := readKey()
		if err != nil {
			s.aborted = true

			return err
		}

		if s.handleKey(key) {
			return nil
		}

		s.render()
	}
}

// handleKey processes a single key press. It returns true when the
// interaction is finished (saved or aborted).
//
//nolint:cyclop
func (s *sortList) handleKey(key string) bool {
	switch key {
	case "ctrl+c", "esc", "q":
		s.aborted = true

		return true
	case "s":
		return true
	case "j", "down":
		if s.cursor < len(s.items)-1 {
			s.cursor++
		}
	case "k", "up":
		if s.cursor > 0 {
			s.cursor--
		}
	case "u", "ctrl+u":
		if s.cursor > 0 {
			s.swap(s.cursor-1, s.cursor)
			s.cursor--
		}
	case "d", "ctrl+d":
		if s.cursor < len(s.items)-1 {
			s.swap(s.cursor, s.cursor+1)
			s.cursor++
		}
	default:
		debug.Log("sortList: unhandled key %q", key)
	}

	return false
}

func (s *sortList) swap(i, j int) {
	s.items[i], s.items[j] = s.items[j], s.items[i]
}

func (s *sortList) render() {
	// clear the previous rendering: one header line, the key bindings,
	// one blank line, one line per item and one trailing blank line.
	lines := len(s.items) + 4
	fmt.Fprintf(Stderr, "\033[%dA\033[J", lines)

	fmt.Fprintf(Stderr, "%s\r\n", s.title)
	fmt.Fprintf(Stderr, "  %s\r\n", sortListKeys)
	fmt.Fprintln(Stderr)

	for i, item := range s.items {
		marker := " "
		if i == s.cursor {
			marker = ">"
		}

		fmt.Fprintf(Stderr, "%s %d. %s\r\n", marker, i+1, item)
	}

	fmt.Fprintln(Stderr)
}

func slicesClone(in []string) []string {
	out := make([]string, len(in))
	copy(out, in)

	return out
}

// readKey reads a single key press from stdin and returns a normalized
// description of it, e.g. "j", "up", "down", "ctrl+c" or "esc".
func readKey() (string, error) {
	buf := make([]byte, 3)
	n, err := os.Stdin.Read(buf)
	if err != nil || n == 0 {
		return "", fmt.Errorf("failed to read key: %w", err)
	}

	switch {
	case n == 1:
		switch buf[0] {
		case 3: // Ctrl+C
			return "ctrl+c", nil
		case 4: // Ctrl+D
			return "ctrl+d", nil
		case 21: // Ctrl+U
			return "ctrl+u", nil
		case 27: // ESC
			return "esc", nil
		case 13, 10: // Enter
			return "enter", nil
		default:
			return string(buf[0]), nil
		}
	case n == 3 && buf[0] == 27 && buf[1] == '[':
		switch buf[2] {
		case 'A':
			return "up", nil
		case 'B':
			return "down", nil
		default:
			return fmt.Sprintf("esc[%c]", buf[2]), nil
		}
	default:
		return strings.TrimSpace(string(buf[:n])), nil
	}
}
