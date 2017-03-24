package qrcon

import (
	"bytes"
	"fmt"
	"image/color"

	"github.com/skip2/go-qrcode"
)

const (
	black = "\033[40m  \033[0m"
	white = "\033[47m  \033[0m"
)

// QRCode returns a string containing an ANSI encoded
// QR Code
func QRCode(content string) (string, error) {
	q, err := qrcode.New(content, qrcode.Medium)
	if err != nil {
		return "", err
	}
	buf := bytes.Buffer{}
	i := q.Image(0)
	b := i.Bounds()
	for x := b.Min.X; x < b.Max.X; x++ {
		for y := b.Min.Y; y < b.Max.Y; y++ {
			col := i.At(x, y)
			if sameColor(col, q.ForegroundColor) {
				buf.WriteString(black)
			} else if sameColor(col, q.BackgroundColor) {
				buf.WriteString(white)
			} else {
				return "", fmt.Errorf("Unexpected color at (%d,%d): %+v", x, y, col)
			}
		}
		buf.WriteString("\n")
	}
	return buf.String(), nil
}

func sameColor(a color.Color, b color.Color) bool {
	r1, g1, b1, a1 := a.RGBA()
	r2, g2, b2, a2 := b.RGBA()
	if r1 == r2 && g1 == g2 && b1 == b2 && a1 == a2 {
		return true
	}
	return false
}
