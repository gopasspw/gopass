package termwiz

import (
	"fmt"

	runewidth "github.com/mattn/go-runewidth"
	"github.com/nsf/termbox-go"
)

//func main() {
//	ls := []string{
//		"foo/bar",
//		"foo/baz",
//		"foo/zab",
//	}
//
//	act, sel := GetSelection(ls)
//	fmt.Printf("User selected: %s %d\n", act, sel)
//}

func tbprint(x, y int, fg, bg termbox.Attribute, msg string) {
	for _, c := range msg {
		termbox.SetCell(x, y, c, fg, bg)
		x += runewidth.RuneWidth(c)
	}
}

func GetSelection(choices []string) (string, int) {
	if err := termbox.Init(); err != nil {
		panic(err)
	}
	defer termbox.Close()
	termbox.SetInputMode(termbox.InputEsc)
	const coldef = termbox.ColorDefault
	termbox.Clear(coldef, coldef)

	cur := 0
	for {
		tbprint(0, 0, coldef, coldef, "Please select:")
		for i, c := range choices {
			mark := " "
			if cur == i {
				mark = ">"
			}
			tbprint(0, 1+i, coldef, coldef, fmt.Sprintf("%s %s\n", mark, c))
		}
		tbprint(0, len(choices)+1, coldef, coldef, "<↑/↓> to change the selection, <→> to show, <←> to copy, <ESC> to quit")
		termbox.Flush()
		switch ev := termbox.PollEvent(); ev.Type {
		case termbox.EventKey:
			switch ev.Key {
			case termbox.KeyEsc:
				return "aborted", cur
			case termbox.KeyArrowLeft:
				return "copy", cur
			case termbox.KeyArrowRight, termbox.KeyEnter:
				return "show", cur
			case termbox.KeyArrowDown, termbox.KeyTab:
				cur++
				if cur >= len(choices) {
					cur = 0
				}
				continue
			case termbox.KeyArrowUp:
				cur--
				if cur < 0 {
					cur = len(choices) - 1
				}
			default:
				if ev.Ch != 0 {
					switch ev.Ch {
					case 'h':
						return "copy", cur
					case 'j':
						cur++
						if cur >= len(choices) {
							cur = 0
						}
					case 'k':
						cur--
						if cur < 0 {
							cur = len(choices) - 1
						}
					case 'l':
						return "show", cur
					}
				}
			}
		}
	}
	return "unreachable", cur
}
