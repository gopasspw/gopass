package qrcon

import "testing"

func TestQRCode(t *testing.T) {
	_, err := QRCode("http://www.justwatch.com/")
	if err != nil {
		t.Fatalf("Failed to generate QR Code")
	}
}
