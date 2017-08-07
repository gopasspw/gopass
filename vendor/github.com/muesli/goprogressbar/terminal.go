/*
 * goprogressbar
 *     Copyright (c) 2016-2017, Christian Muehlhaeuser <muesli@gmail.com>
 *
 *   For license see LICENSE
 */

package goprogressbar

import "fmt"

func clearCurrentLine() {
	fmt.Fprintf(Stdout, "\033[2K\r")
}

func moveCursorUp(lines uint) {
	fmt.Fprintf(Stdout, "\033[%dA", lines)
}

func moveCursorDown(lines uint) {
	fmt.Fprintf(Stdout, "\033[%dB", lines)
}
