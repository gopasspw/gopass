//go:build (arm || arm64 || amd64 || 386) && (linux || windows || (cgo && darwin) || freebsd || netbsd)

package otp

import (
	"context"
	"strings"

	"github.com/gopasspw/gopass/internal/out"
	"github.com/kbinani/screenshot"
	"github.com/makiuchi-d/gozxing"
	"github.com/makiuchi-d/gozxing/qrcode"
)

// ParseScreen will attempt to parse all available screen and will look for otpauth QR codes. It returns the first one
// it has found.
func ParseScreen(ctx context.Context) (string, error) {
	for i := range screenshot.NumActiveDisplays() {
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
		if qr := result.GetText(); strings.HasPrefix(qr, "otpauth://") {
			out.OKf(ctx, "Found an otpauth:// QR code on screen n°%d (%v) for %s", i, img.Bounds(),
				// otpauth:// is 10 char, we display label information, but not the parameters containing the secret
				qr[10:10+strings.Index(qr[10:], "?")])

			return qr, nil
		}
		out.Warningf(ctx, "Not an otpauth:// QR code, please make sure to only have your OTP qrcode displayed.")
	}

	return "", nil
}
