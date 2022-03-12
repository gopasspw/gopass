// Package qrcon implements a QR Code ANSI printer for displaying QR codes on
// the console.
package qrcon

import (
	"fmt"
	"image/color"
	"strings"

	"github.com/skip2/go-qrcode"
)

const (
	black = "\033[40m  \033[0m"
	white = "\033[47m  \033[0m"
)

// ErrUnknowColor is returned when the color is unknown.
var ErrUnknowColor = fmt.Errorf("unknown color")

// QRCode returns a string containing an ANSI encoded
// QR Code.
func QRCode(content string) (string, error) {
	q, err := qrcode.New(content, qrcode.Medium)
	if err != nil {
		return "", fmt.Errorf("failed to create qr code: %w", err)
	}

	var sb strings.Builder

	i := q.Image(0)
	b := i.Bounds()

	for x := b.Min.X; x < b.Max.X; x++ {
		for y := b.Min.Y; y < b.Max.Y; y++ {
			col := i.At(x, y)

			switch {
			case sameColor(col, q.ForegroundColor):
				_, _ = sb.WriteString(black)
			case sameColor(col, q.BackgroundColor):
				_, _ = sb.WriteString(white)
			default:
				return "", fmt.Errorf("error at (%d,%d): %+v: %w", x, y, col, ErrUnknowColor)
			}
		}

		_, _ = sb.WriteString("\n")
	}

	return sb.String(), nil
}

func sameColor(a color.Color, b color.Color) bool {
	r1, g1, b1, a1 := a.RGBA()
	r2, g2, b2, a2 := b.RGBA()

	if r1 == r2 && g1 == g2 && b1 == b2 && a1 == a2 {
		return true
	}

	return false
}
