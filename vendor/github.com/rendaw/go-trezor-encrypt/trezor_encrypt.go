// Encrypt and decrypt things with Trezor
package trezor_encrypt

import (
	"encoding/binary"
	"fmt"
	"os"

	"github.com/rendaw/go-trezor"
	"github.com/rendaw/go-trezor/messages"

	"github.com/golang/protobuf/proto"
	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
)

const space = 4

func GetAddressN() []uint32 {
	return []uint32{10, 0}
}

//Show a confirmation dialog
func DoCheck(message string, okayText string, cancelText string) (bool, error) {
	okay := false

	gtk.Init(nil)
	{
		provider, err := gtk.CssProviderNew()
		provider.LoadFromData(
			".top { padding: 0.5em; border: none; } ")
		screen, err := gdk.ScreenGetDefault()
		if err != nil {
			return false, err
		}
		gtk.AddProviderForScreen(screen, provider, gtk.STYLE_PROVIDER_PRIORITY_APPLICATION)
	}

	cancel := func() {
		gtk.MainQuit()
	}

	done := func() {
		okay = true
		gtk.MainQuit()
	}

	main, err := gtk.BoxNew(gtk.ORIENTATION_VERTICAL, space)
	if err != nil {
		return false, err
	}

	{
		label, err := gtk.LabelNew(message)
		if err != nil {
			return false, err
		}
		label.SetLineWrap(true)
		label.SetMaxWidthChars(50)
		main.PackStart(label, false, false, 0)
	}

	{
		actions, err := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, space)
		if err != nil {
			return false, err
		}
		{
			button, err := gtk.ButtonNew()
			if err != nil {
				return false, err
			}
			button.SetLabel(okayText)
			actions.PackStart(button, false, false, 0)
			button.Connect("clicked", done)
		}
		{
			button, err := gtk.ButtonNew()
			if err != nil {
				return false, err
			}
			button.SetLabel(cancelText)
			actions.PackStart(button, false, false, 0)
			button.Connect("clicked", cancel)
		}
		actions.SetHAlign(gtk.ALIGN_CENTER)
		main.PackStart(actions, false, false, 0)
	}

	{
		win, err := gtk.WindowNew(gtk.WINDOW_TOPLEVEL)
		if err != nil {
			return false, err
		}
		win.SetTitle("Trezor - Confirm")
		win.Connect("destroy", cancel)
		win.Connect("key-press-event", func(win *gtk.Window, ev *gdk.Event) {
			keyEvent := &gdk.EventKey{ev}
			if keyEvent.KeyVal() == gdk.KEY_Return {
				done()
				return
			}
			if keyEvent.KeyVal() == gdk.KEY_Escape {
				cancel()
				return
			}
		})

		frame, err := gtk.FrameNew("")
		if err != nil {
			return false, err
		}
		style, err := frame.GetStyleContext()
		if err != nil {
			return false, err
		}
		style.AddClass("top")
		frame.SetShadowType(gtk.SHADOW_NONE)
		frame.Add(main)

		win.Add(frame)
		win.ShowAll()
	}

	gtk.Main()

	return okay, nil
}

// Get a pin from a user.  You shouldn't need to use this directly - the Encrypt* functions will invoke it themselves.
func DoPin(message string) (bool, string, error) {
	var dontFlash bool
	{
		flashEnv, found := os.LookupEnv("TREZOR_ENCRYPT_NOFLASH")
		if found {
			dontFlash = flashEnv == "1"
		} else {
			dontFlash = false
		}
	}

	value := ""
	_done := false

	buttons := make(map[int32]*gtk.Button)
	clearId := int32(1)

	gtk.Init(nil)
	{
		provider, err := gtk.CssProviderNew()
		provider.LoadFromData(
			".pinkey { border: 1.5px solid black; border-radius: 0.3em; background: transparent; } " +
				".flash { background: #3F7FBF; } " +
				".top { padding: 0.5em; border: none; } ")
		screen, err := gdk.ScreenGetDefault()
		if err != nil {
			return false, "", err
		}
		gtk.AddProviderForScreen(screen, provider, gtk.STYLE_PROVIDER_PRIORITY_APPLICATION)
	}
	win, err := gtk.WindowNew(gtk.WINDOW_TOPLEVEL)
	if err != nil {
		return false, "", err
	}

	flash := func(buttonId int32) {
		if dontFlash {
			return
		}
		style, err := buttons[buttonId].GetStyleContext()
		if err != nil {
			return
		}
		style.AddClass("flash")
		glib.TimeoutAdd(100, func() bool {
			style, err := buttons[buttonId].GetStyleContext()
			if err != nil {
				return false
			}
			style.RemoveClass("flash")
			return false
		})
	}

	cancel := func() {
		win.Hide()
		gtk.MainQuit()
	}

	done := func() {
		_done = true
		win.Hide()
		gtk.MainQuit()
	}

	clear := func() {
		value = ""
		flash(clearId)
	}

	entry := func(letter int32) {
		value = value + string(letter)
		flash(letter)
	}

	main, err := gtk.BoxNew(gtk.ORIENTATION_VERTICAL, space)
	if err != nil {
		return false, "", err
	}

	{
		label, err := gtk.LabelNew(message)
		if err != nil {
			return false, "", err
		}
		label.SetLineWrap(true)
		label.SetMaxWidthChars(50)
		main.PackStart(label, false, false, 0)
	}

	{
		grid, err := gtk.GridNew()
		grid.SetColumnSpacing(5)
		grid.SetRowSpacing(5)
		if err != nil {
			return false, "", err
		}
		for i, v := range "789456123" {
			button, err := gtk.ButtonNew()
			if err != nil {
				return false, "", err
			}
			buttons[v] = button
			v2 := v
			button.Connect("clicked", func() {
				entry(v2)
			})
			style, err := button.GetStyleContext()
			if err != nil {
				return false, "", err
			}
			style.AddClass("pinkey")
			col := i % 3
			row := int(i / 3)
			grid.Attach(button, col, row, 1, 1)
		}
		grid.SetHAlign(gtk.ALIGN_CENTER)
		main.PackStart(grid, true, false, 0)
	}

	{
		actions, err := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, space)
		if err != nil {
			return false, "", err
		}
		{
			button, err := gtk.ButtonNew()
			if err != nil {
				return false, "", err
			}
			button.SetLabel("Done")
			actions.PackStart(button, false, false, 0)
			button.Connect("clicked", done)
		}
		{
			button, err := gtk.ButtonNew()
			if err != nil {
				return false, "", err
			}
			button.SetLabel("Clear")
			actions.PackStart(button, false, false, 0)
			button.Connect("clicked", clear)
			buttons[clearId] = button
		}
		{
			button, err := gtk.ButtonNew()
			if err != nil {
				return false, "", err
			}
			button.SetLabel("Cancel")
			actions.PackStart(button, false, false, 0)
			button.Connect("clicked", cancel)
		}
		actions.SetHAlign(gtk.ALIGN_CENTER)
		main.PackStart(actions, false, false, 0)
	}

	{
		win.SetTitle("Trezor PIN Entry")
		win.Connect("destroy", cancel)
		win.Connect("key-press-event", func(win *gtk.Window, ev *gdk.Event) {
			keyEvent := &gdk.EventKey{ev}
			if keyEvent.KeyVal() == gdk.KEY_Return {
				done()
				return
			}
			if keyEvent.KeyVal() == gdk.KEY_BackSpace {
				clear()
				return
			}
			if keyEvent.KeyVal() == gdk.KEY_Escape {
				cancel()
				return
			}
			extraKeysets := []string{"xcvsdfwer", "m,.jkluio"}
			for _, keyset := range append([]string{"123456789"}, extraKeysets...) {
				for i, k := range keyset {
					if k == gdk.KeyvalToUnicode(keyEvent.KeyVal()) {
						entry(int32('1' + i))
						return
					}
				}
			}
		})

		frame, err := gtk.FrameNew("")
		if err != nil {
			return false, "", err
		}
		style, err := frame.GetStyleContext()
		if err != nil {
			return false, "", err
		}
		style.AddClass("top")
		frame.SetShadowType(gtk.SHADOW_NONE)
		frame.Add(main)

		win.Add(frame)
		win.ShowAll()
	}
	gtk.Main()

	return _done, value, nil
}

// Encrypt or decrypt the specified bytes with the Trezor device.  The prompt message is shown to the user on their
// computer.  The key is the message displayed on the Trezor pin screen and is factored into the encryption, so you
// must use the same key when decrypting. value is the data to encrypt/decrypt.
//
// For compatibility with other Trezor software - the data to encrypt is prepended with its length (4 bytes) and padded
// to a length multiple of 16.  Additionally, the BIP32 key derivation path is 10, 0.
func EncryptWithDevice(device trezor.Transport, encrypt bool, prompt string, key string, value []byte) ([]byte, error) {
	var padded []byte
	if encrypt {
		padded = make([]byte, uint32((4+len(value))/16+1)*16)
		binary.BigEndian.PutUint32(padded[:], uint32(len(value)))
		copy(padded[4:], value)
	} else {
		padded = value
	}
	t := true
	err := device.Write(&messages.CipherKeyValue{
		AddressN:     GetAddressN(),
		Key:          &key,
		Value:        padded[:],
		Encrypt:      &encrypt,
		AskOnEncrypt: &t,
		AskOnDecrypt: &t,
		Iv:           []byte{},
	})
	if err != nil {
		return nil, err
	}
	for {
		messageType, message, err := device.Read()
		if err != nil {
			return nil, err
		}
		if false {

		} else if messageType == messages.MessageType_MessageType_Failure {
			var failure messages.Failure
			proto.Unmarshal(message, &failure)
			return nil, fmt.Errorf("[%d] %s", failure.Code, *failure.Message)
		} else if messageType == messages.MessageType_MessageType_PinMatrixRequest {
			done, value, err := DoPin(prompt)
			if err != nil {
				return nil, err
			}
			if done {
				err = device.Write(&messages.PinMatrixAck{Pin: &value})
				if err != nil {
					return nil, err
				}
			} else {
				err = device.Write(&messages.Cancel{})
				if err != nil {
					return nil, err
				}
				return nil, nil
			}
		} else if messageType == messages.MessageType_MessageType_ButtonRequest {
			err = device.Write(&messages.ButtonAck{})
			if err != nil {
				return nil, err
			}
		} else if messageType == messages.MessageType_MessageType_CipheredKeyValue {
			var ciphered messages.CipheredKeyValue
			proto.Unmarshal(message, &ciphered)
			var clipped []byte
			if encrypt {
				clipped = ciphered.Value
			} else {
				length := binary.BigEndian.Uint32(ciphered.Value[0:4])
				clipped = ciphered.Value[4 : 4+length]
			}
			return clipped, nil
		} else {
			return nil, fmt.Errorf("Unexpected response from Trezor [%d]", messageType)
		}
	}
}

// See the documentation for EncryptWithDevice.  This helper method automatically uses the first detected Trezor device.
func Encrypt(encrypt bool, prompt string, key string, value []byte) ([]byte, error) {
	var devices []*trezor.HidTransport
	for {
		var err error
		devices, err = trezor.Enumerate()
		if err != nil {
			return nil, err
		}
		if len(devices) == 0 {
			okay, err := DoCheck("Couldn't find any Trezor devices", "Retry", "Cancel")
			if err != nil {
				return nil, err
			}
			if !okay {
				return nil, fmt.Errorf("Canceled by user")
			}
		} else {
			break
		}
	}
	var result []byte
	device := devices[0]
	TrezorDo(device, func(features messages.Features) error {
		var err error
		result, err = EncryptWithDevice(device, encrypt, prompt, key, value)
		return err
	})
	return result, nil
}

func GetPublicKey(device trezor.Transport, prompt string) ([]byte, error) {
	curve := "secp256k1" // from bitcoin
	coin := "Bitcoin"
	err := device.Write(&messages.GetPublicKey{
		AddressN:       nil,
		EcdsaCurveName: &curve,
		CoinName:       &coin,
	})
	if err != nil {
		return nil, err
	}
	for {
		messageType, message, err := device.Read()
		if err != nil {
			return nil, err
		}
		if false {
		} else if messageType == messages.MessageType_MessageType_Failure {
			var failure messages.Failure
			proto.Unmarshal(message, &failure)
			return nil, fmt.Errorf("[%d] %s", failure.Code, *failure.Message)
		} else if messageType == messages.MessageType_MessageType_PinMatrixRequest {
			done, value, err := DoPin(prompt)
			if err != nil {
				return nil, err
			}
			if done {
				err = device.Write(&messages.PinMatrixAck{Pin: &value})
				if err != nil {
					return nil, err
				}
			} else {
				err = device.Write(&messages.Cancel{})
				if err != nil {
					return nil, err
				}
				return nil, nil
			}
		} else if messageType == messages.MessageType_MessageType_ButtonRequest {
			err = device.Write(&messages.ButtonAck{})
			if err != nil {
				return nil, err
			}
		} else if messageType == messages.MessageType_MessageType_PublicKey {
			var key messages.PublicKey
			err = proto.Unmarshal(message, &key)
			if err != nil {
				return nil, err
			}
			return key.Node.PublicKey, nil
		} else {
			return nil, fmt.Errorf("Unexpected response from Trezor [%d]", messageType)
		}
	}
}

func TrezorDo(device trezor.Transport, do func(features messages.Features) error) error {
	err := device.Open()
	if err != nil {
		return err
	}
	err = device.Write(&messages.Initialize{})
	if err != nil {
		return err
	}
	messageType, message, err := device.Read()
	if err != nil {
		_ = device.Close()
		return fmt.Errorf("Failed to query features of Trezor device %s: %s", device.String(), err)
	}
	if messageType != messages.MessageType_MessageType_Features {
		_ = device.Close()
		return fmt.Errorf("Failed to query features of Trezor device %s: received response type %s", device.String(), messageType)
	}
	var features messages.Features
	err = proto.Unmarshal(message, &features)
	if err != nil {
		return fmt.Errorf("Failed to read features of Trezor device at %s: %s", device.String(), err)
	}
	err = do(features)
	if err != nil {
		device.Close()
		return err
	}
	err = device.Close()
	if err != nil {
		return err
	}
	return nil
}
