package cui

import (
	"context"
	"fmt"
	"runtime"

	"github.com/fatih/color"
	"github.com/jroimartin/gocui"
	"github.com/justwatchcom/gopass/utils/ctxutil"
)

type selection struct {
	prompt    string
	usage     string
	choices   []string
	action    string
	selection int
}

func (s *selection) layout(g *gocui.Gui) error {
	maxx, maxy := g.Size()
	v, err := g.SetView("header", 0, 0, max(80, len(s.prompt)), 3)
	if err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Frame = true
	}

	v.Clear()
	s.renderHeader(v, maxx)

	v, err = g.SetView("list", 0, 3, maxx-1, maxy-3)
	if err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Frame = false
		v.Highlight = true
		v.SelBgColor = gocui.ColorWhite
		v.SelFgColor = gocui.ColorBlack

		if _, err := g.SetCurrentView("list"); err != nil {
			return err
		}
	}

	v.Clear()
	s.render(v, maxx)

	v, err = g.SetView("footer", 0, maxy-3, max(80, len(s.usage)), maxy-1)
	if err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Frame = true
	}

	v.Clear()
	s.renderFooter(v, maxx)

	return nil
}

func (s *selection) renderHeader(v *gocui.View, maxx int) {
	fmt.Fprintf(v, "%s\n", color.GreenString("gopass"))
	fmt.Fprintf(v, s.prompt)
}

func (s *selection) render(v *gocui.View, maxx int) {
	for _, item := range s.choices {
		fmt.Fprintf(v, "%s\n", item)
	}
}

func (s *selection) renderFooter(v *gocui.View, maxx int) {
	fmt.Fprintf(v, "\u001b[1m%s\u001b[0m\n", s.usage)
}

func (s *selection) keybindings(g *gocui.Gui) error {
	for _, kb := range []struct {
		name   string
		key    interface{}
		action func(*gocui.Gui, *gocui.View) error
	}{
		{"", 'q', s.quit},
		{"", gocui.KeyEsc, s.quit},
		{"list", 'h', s.copy},
		{"list", gocui.KeyArrowLeft, s.copy},
		{"list", 'j', s.cursorDown},
		{"list", gocui.KeyArrowDown, s.cursorDown},
		{"list", 'k', s.cursorUp},
		{"list", gocui.KeyArrowUp, s.cursorUp},
		{"list", 'l', s.show},
		{"list", gocui.KeyPgup, s.pageUp},
		{"list", gocui.KeyPgdn, s.pageDown},
		{"list", gocui.KeyArrowRight, s.show},
		{"list", gocui.KeyEnter, s.def},
		{"list", 's', s.sync},
	} {
		if err := g.SetKeybinding(kb.name, kb.key, gocui.ModNone, kb.action); err != nil {
			return err
		}
	}
	return nil
}

func (s *selection) quit(g *gocui.Gui, v *gocui.View) error {
	return gocui.ErrQuit
}

func (s *selection) copy(g *gocui.Gui, v *gocui.View) error {
	s.selection = getSelectedLine(v)
	s.action = "copy"
	return gocui.ErrQuit
}

func (s *selection) cursorDown(g *gocui.Gui, v *gocui.View) error {
	y := getSelectedLine(v)
	if y < len(s.choices)-1 {
		v.MoveCursor(0, 1, false)
	}
	return nil
}

func (s *selection) cursorUp(g *gocui.Gui, v *gocui.View) error {
	v.MoveCursor(0, -1, false)
	return nil
}

func (s *selection) pageDown(g *gocui.Gui, v *gocui.View) error {
	y := getSelectedLine(v)
	if y < len(s.choices)-10 {
		v.MoveCursor(0, 10, false)
	}
	return nil
}

func (s *selection) pageUp(g *gocui.Gui, v *gocui.View) error {
	v.MoveCursor(0, -10, false)
	return nil
}

func (s *selection) show(g *gocui.Gui, v *gocui.View) error {
	s.selection = getSelectedLine(v)
	s.action = "show"
	return gocui.ErrQuit
}

func (s *selection) def(g *gocui.Gui, v *gocui.View) error {
	s.selection = getSelectedLine(v)
	s.action = "default"
	return gocui.ErrQuit
}

func (s *selection) sync(g *gocui.Gui, v *gocui.View) error {
	s.action = "sync"
	return gocui.ErrQuit
}

// GetSelection show a navigateable multiple-choice list to the user
// and returns the selected entry along with the action
func GetSelection(ctx context.Context, prompt, usage string, choices []string) (string, int) {
	if ctxutil.IsAlwaysYes(ctx) || !ctxutil.IsInteractive(ctx) {
		return "impossible", 0
	}
	g, err := gocui.NewGui(gocui.OutputNormal)
	if err != nil {
		panic(err)
	}
	defer g.Close()

	g.InputEsc = true
	g.Mouse = false
	g.Cursor = false
	if runtime.GOOS == "windows" {
		g.ASCII = true
	}

	s := selection{
		prompt:    prompt,
		usage:     usage,
		choices:   choices,
		action:    "",
		selection: 0,
	}

	g.SetManagerFunc(s.layout)

	if err := s.keybindings(g); err != nil {
		panic(err)
	}

	if err := g.MainLoop(); err != nil {
		if err != gocui.ErrQuit {
			return "aborted", s.selection
		}
	}
	return s.action, s.selection
}

func getSelectedLine(v *gocui.View) int {
	_, y := v.Cursor()
	_, oy := v.Origin()

	return y + oy
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
