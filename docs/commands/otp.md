# `otp` command

The `otp` command generates TOTP tokens from an OTP URL (`otpauth://`).
The command tries to parse the password and the totp fields as an OTP URI.

Note: HTOP is supported, but requires a `counter` field to keep track of it.

Note: If `show.safecontent` is enabled, OTP URIs are hidden from the `show` command,
see the [docs for show](show.md#parsing-and-secrets) to learn more about it.

## Screen capture dependency

The `--snip` mode requires the `github.com/kbinani/screenshot` library to capture the
contents of the display(s) so that gopass can locate and decode an OTP QR code.  Screen
capture is **only performed when the user explicitly passes `--snip`** (`-s`); it is never
triggered automatically.

This capability is compiled in by default on supported platforms
(`arm`, `arm64`, `amd64`, `386` on Linux, Windows, FreeBSD, and NetBSD; CGo Darwin).
Users who prefer to omit this surface (e.g. in enterprise or policy-restricted environments)
can build gopass without it:

```
go build -tags noscreenshot .
```

When built with `noscreenshot`, the `--snip` flag will return an error on all platforms
and the `github.com/kbinani/screenshot` package will not be linked into the binary.

## Modes of operation

* Generate the current TOTP token from a valid OTP URL
* Snip the screen to add a TOTP QR code as an OTP field to an entry.

## Flags

| Flag         | Aliases | Description                                                              |
|--------------|---------|--------------------------------------------------------------------------|
| `--clip`     | `-c`    | Copy the time-based token into the clipboard.                            |
| `--alsoclip` | `-C`    | Copy the time-based token into the clipboard and show it.                |
| `--qr`       | `-q`    | Write QR code to file.                                                   |
| `--chained`  | `-p`    | chain the token to the password                                          |
| `--password` | `-o`    | Only display the token. For use in scripts.                              |
| `--snip`     | `-s`    | Try and find a QR code in the screen content to add as OTP to the entry. |

## Supported formats

Your secret needs to either contain a `otpauth`, `hotp` or a `totp` field.
When using the OTP code directly you can simply add it to a secret using
`gopass insert your/entry totp`.

The `otp` command also tries to parse the body of your secret to try and find a line starting
by `otpauth://` in case you're not using the key-value format for your secret.

Finally, if your secret contains nothing but a password on the first line, the `otp` command
will try and use that password to generate an OTP code. This allows use-cases where you
store your password in a given entry and your OTP code in another dedicated entry.

The otpauth URIs are typically communicated through a QR code which can be read on Linux using
the `gopass otp -s your/entry` flag. It should also work if they are added using
`gopass insert your/entry otpauth`, but won't work if you add them under the `totp`
or `hotp` keys.

Steam OTP is supported, but requires using the `otpauth` URI input to specify the
encoder, e.g. `otpauth://totp/username%20steam:username?secret=qlt6vmy6svfx4bt4rpmisaiyol6hihca&period=30&digits=5&issuer=username%20steam&encoder=steam`.
