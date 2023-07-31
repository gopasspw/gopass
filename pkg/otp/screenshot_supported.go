//go:build linux || windows || darwin || freebsd || netbsd
// +build linux windows darwin freebsd netbsd

package otp

import (
	"context"
	"strings"

	"github.com/AnomalRoil/screenshot"
	"github.com/gopasspw/gopass/internal/out"
	"github.com/makiuchi-d/gozxing"
	"github.com/makiuchi-d/gozxing/qrcode"
)

// ParseScreen will attempt to parse all available screen and will look for otpauth QR codes. It returns the first one
// it has found.
func ParseScreen(ctx context.Context) (string, error) {
	var qr string
	for i := 0; i < screenshot.NumActiveDisplays(); i++ {
		out.Noticef(ctx, "Scanning screen n°%d", i)

		img, err := screenshot.CaptureDisplay(i)
		if err != nil {
			return "", err
		}

		out.OKf(ctx, "Area scanned on screen n°%d: %v", i, img.Bounds())
		bmp, err := gozxing.NewBinaryBitmapFromImage(img)
		if err != nil {
			return "", err
		}
		qrReader := qrcode.NewQRCodeReader()
		result, err := qrReader.Decode(bmp, nil)
		if err != nil {
			out.Warningf(ctx, "No QR code found while parsing screen n°%d.", i)

			continue
		}

		out.Noticef(ctx, "Found a qrcode, checking.")
		if strings.HasPrefix(result.GetText(), "otpauth://") {
			qr = result.GetText()
			out.OKf(ctx, "Found an otpauth:// QR code on screen n°%d (%v)", i, img.Bounds())

			break
		}
		out.Warningf(ctx, "Not an otpauth:// QR code, please make sure to only have your OTP qrcode displayed.")
	}

	return qr, nil
}
