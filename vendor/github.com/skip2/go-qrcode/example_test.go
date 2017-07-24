// go-qrcode
// Copyright 2014 Tom Harwood

package qrcode

import (
	"fmt"
	"os"
)

func ExampleEncode() {
	var png []byte
	png, err := Encode("https://example.org", Medium, 256)

	if err != nil {
		fmt.Printf("Error: %s", err.Error())
	} else {
		fmt.Printf("PNG is %d bytes long", len(png))
	}
}

func ExampleWriteFile() {
	filename := "example.png"

	err := WriteFile("https://example.org", Medium, 256, filename)

	if err != nil {
		err = os.Remove(filename)
	}

	if err != nil {
		fmt.Printf("Error: %s", err.Error())
	}
}
