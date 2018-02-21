# Encrypt with Trezor
### In Go

This package provides methods for encrypting and decrypting data with a Trezor.

When a PIN is necessay an on-screen PIN entry will be displayed.  You may enter your PIN using the buttons or via the keyboard using these grids

```
w e r      u i o      7 8 9
s d f      j k l      4 5 6
x c v      m , .      1 2 3
```

Press `enter` to submit the PIN, `escape` to cancel, or `backspace` to clear the PIN.

<br/>
<hr/>

```
start := []byte("wizard")
log.Printf("Decrypted: %s", hex.EncodeToString(start))
mid, err := trezor_encrypt.Encrypt(true, "Hello user encrypt this nonsense", "Example App", start)
if err != nil {
    log.Fatal(err)
    return
}
log.Printf("Encrypted: %s", hex.EncodeToString(mid))
end, err := trezor_encrypt.Encrypt(false, "Please decrypt it now", "Example App", mid)
if err != nil {
    log.Fatal(err)
    return
}
log.Printf("Decrypted: %s", hex.EncodeToString(end))
```

Some notes:

1. This will show a GTK keypad so encryption can be used in software without a terminal.  The `DISPLAY` environment variable is probably used.
2. There are plans to have a terminal keypad as well, if the TTY is defined in an environment variable.
3. You should take note of the Trezor device when encrypting and use that to locate the same Trezor device when decrypting. The best way to do this is a work in progress.
4. Disable the keypad flash by setting environment variable `TREZOR_KEYPAD_NOFLASH=1`

Documentation [here](https://godoc.org/github.com/rendaw/go-trezor-encrypt)