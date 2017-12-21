/*
 * goprogressbar
 *     Copyright (c) 2016-2017, Christian Muehlhaeuser <muesli@gmail.com>
 *
 *   For license see LICENSE
 */

package goprogressbar

import (
	"fmt"
	"io"
	"math"
	"os"
	"strings"
	"syscall"
	"time"

	"golang.org/x/crypto/ssh/terminal"
)

var (
	// Stdout defines where output gets printed to
	Stdout io.Writer = os.Stdout
	// BarFormat defines the bar design
	BarFormat = "[#>-]"
)

// ProgressBar is a helper for printing a progress bar
type ProgressBar struct {
	// Text displayed on the very left
	Text string
	// Text prepending the bar
	PrependText string
	// Max value (100%)
	Total int64
	// Current progress value
	Current int64
	// Desired bar width
	Width uint

	// If a PrependTextFunc is set, the PrependText will be automatically
	// generated on every print
	PrependTextFunc func(p *ProgressBar) string

	lastPrintTime time.Time
}

// MultiProgressBar is a helper for printing multiple progress bars
type MultiProgressBar struct {
	ProgressBars []*ProgressBar

	lastPrintTime time.Time
}

// percentage returns the percentage bound between 0.0 and 1.0
func (p *ProgressBar) percentage() float64 {
	pct := float64(p.Current) / float64(p.Total)
	if p.Total == 0 {
		if p.Current == 0 {
			// When both Total and Current are 0, show a full progressbar
			pct = 1
		} else {
			pct = 0
		}
	}

	// percentage is bound between 0 and 1
	return math.Min(1, math.Max(0, pct))
}

// UpdateRequired returns true when this progressbar wants an update regardless
// of fps limitation
func (p *ProgressBar) UpdateRequired() bool {
	return p.Current == 0 || p.Current == p.Total
}

// LazyPrint writes the progress bar to stdout if a significant update occurred
func (p *ProgressBar) LazyPrint() {
	now := time.Now()
	if p.UpdateRequired() || now.Sub(p.lastPrintTime) > time.Second/25 {
		// Max out at 25fps
		p.lastPrintTime = now
		p.Print()
	}
}

// Clear deletes everything on the current terminal line, hence removing a printed progressbar
func (p *ProgressBar) Clear() {
	clearCurrentLine()
}

// Print writes the progress bar to stdout
func (p *ProgressBar) Print() {
	if p.PrependTextFunc != nil {
		p.PrependText = p.PrependTextFunc(p)
	}
	pct := p.percentage()
	clearCurrentLine()

	pcts := fmt.Sprintf("%.2f%%", pct*100)
	for len(pcts) < 7 {
		pcts = " " + pcts
	}

	tiWidth, _, _ := terminal.GetSize(int(syscall.Stdin))
	if tiWidth < 0 {
		// we're not running inside a real terminal (e.g. CI)
		// we assume a width of 80
		tiWidth = 80
	}
	barWidth := uint(math.Min(float64(p.Width), float64(tiWidth)/2.0))

	size := int(barWidth) - len(pcts) - 4
	fill := int(math.Max(2, math.Floor((float64(size)*pct)+.5)))
	if size < 16 {
		barWidth = 0
	}

	text := p.Text
	maxTextWidth := int(tiWidth) - 3 - int(barWidth) - len(p.PrependText)
	if maxTextWidth < 0 {
		maxTextWidth = 0
	}
	if len(p.Text) > maxTextWidth {
		if len(p.Text)-maxTextWidth+3 < len(p.Text) {
			text = "..." + p.Text[len(p.Text)-maxTextWidth+3:]
		} else {
			text = ""
		}
	}

	// Print text
	s := fmt.Sprintf("%s%s  %s ",
		text,
		strings.Repeat(" ", maxTextWidth-len(text)),
		p.PrependText)
	fmt.Fprint(Stdout, s)

	if barWidth > 0 {
		progChar := BarFormat[2]
		if p.Current == p.Total {
			progChar = BarFormat[1]
		}

		// Print progress bar
		fmt.Fprintf(Stdout, "%c%s%c%s%c %s",
			BarFormat[0],
			strings.Repeat(string(BarFormat[1]), fill-1),
			progChar,
			strings.Repeat(string(BarFormat[3]), size-fill),
			BarFormat[4],
			pcts)
	}
}

// AddProgressBar adds another progress bar to the multi struct
func (mp *MultiProgressBar) AddProgressBar(p *ProgressBar) {
	mp.ProgressBars = append(mp.ProgressBars, p)

	if len(mp.ProgressBars) > 1 {
		fmt.Println()
	}
	mp.Print()
}

// Print writes all progress bars to stdout
func (mp *MultiProgressBar) Print() {
	moveCursorUp(uint(len(mp.ProgressBars)))

	for _, p := range mp.ProgressBars {
		moveCursorDown(1)
		p.Print()
	}
}

// LazyPrint writes all progress bars to stdout if a significant update occurred
func (mp *MultiProgressBar) LazyPrint() {
	forced := false
	for _, p := range mp.ProgressBars {
		if p.UpdateRequired() {
			forced = true
			break
		}
	}

	now := time.Now()
	if !forced {
		forced = now.Sub(mp.lastPrintTime) > time.Second/25
	}

	if forced {
		// Max out at 25fps
		mp.lastPrintTime = now

		moveCursorUp(uint(len(mp.ProgressBars)))
		for _, p := range mp.ProgressBars {
			moveCursorDown(1)
			p.Print()
		}
	}
}
