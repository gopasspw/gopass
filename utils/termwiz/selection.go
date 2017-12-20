package termwiz

import (
	"context"
	"fmt"

	"github.com/gdamore/tcell/termbox"
	"github.com/justwatchcom/gopass/utils/ctxutil"
	runewidth "github.com/mattn/go-runewidth"
)

func tbprint(x, y int, fg, bg termbox.Attribute, msg string) {
	for _, c := range msg {
		termbox.SetCell(x, y, c, fg, bg)
		x += runewidth.RuneWidth(c)
	}
}

// GetSelection show a navigateable multiple-choice list to the user
// and returns the selected entry along with the action
func GetSelection(ctx context.Context, prompt, usage string, choices []string) (string, int) {
	if prompt != "" {
		prompt += " "
	}

	if err := termbox.Init(); err != nil {
		panic(err)
	}
	defer termbox.Close()

	const coldef = termbox.ColorDefault
	termbox.Clear(coldef, coldef)

	cur := 0
	for {
		// check for context cancelation
		select {
		case <-ctx.Done():
			return "aborted", cur
		default:
		}
		termbox.Clear(coldef, coldef)
		tbprint(0, 0, coldef, coldef, prompt+"Please select:")
		_, h := termbox.Size()
		offset := 0
		if len(choices)+2 > h && cur > h-3 {
			offset = cur
		}
		for i := offset; i < len(choices) && i-offset < h; i++ {
			c := choices[i]
			mark := " "
			if cur == i {
				mark = ">"
			}
			tbprint(0, 1+i-offset, coldef, coldef, fmt.Sprintf("%s %s", mark, c))
		}
		usageLine := usage
		if usageLine == "" {
			usageLine = "<↑/↓> to change the selection, <→> to show, <←> to copy, <s> to sync, <ESC> to quit"
		}
		if ctxutil.IsDebug(ctx) {
			usageLine += " - DEBUG: " + fmt.Sprintf("Offset: %d - Cur: %d - Choices: %d", offset, cur, len(choices))
		}
		tbprint(0, h-1, coldef, coldef, usageLine)
		_ = termbox.Flush()
		var act string
		if act, cur = tbpoll(cur, len(choices)); act != "" {
			return act, cur
		}
	}
}

func tbpoll(cur, max int) (string, int) {
	switch ev := termbox.PollEvent(); ev.Type {
	case termbox.EventKey:
		switch ev.Key {
		case termbox.KeyEsc:
			return "aborted", cur
		case termbox.KeyArrowLeft:
			return "copy", cur
		case termbox.KeyArrowRight:
			return "show", cur
		case termbox.KeyEnter:
			return "default", cur
		case termbox.KeyArrowDown, termbox.KeyTab:
			cur++
			if cur >= max {
				cur = 0
			}
			return "", cur
		case termbox.KeyArrowUp:
			cur--
			if cur < 0 {
				cur = max - 1
			}
			return "", cur
		default:
			if ev.Ch != 0 {
				return tbchars(ev.Ch, cur, max)
			}
		}
	}
	return "", cur
}

func tbchars(ch rune, cur, max int) (string, int) {
	switch ch {
	case 'h':
		return "copy", cur
	case 'j':
		cur++
		if cur >= max {
			cur = 0
		}
		return "", cur
	case 'k':
		cur--
		if cur < 0 {
			cur = max - 1
		}
		return "", cur
	case 'l':
		return "show", cur
	case 's':
		return "sync", cur
	default:
		return "", cur
	}
}
