# `otp` command

The `otp` command generates TOTP tokens from an OTP URL (`otpauth://`).
The command tries to parse the password and the totp fields as an OTP URL.

Note: HTOP is currently not supported.

## Modes of operation

* Generate the current TOTP token from a valid OTP URL

## Flags

Flag | Aliases | Description
---- | ------- | -----------
`--clip` | `-c` | Copy the time-based token into the clipboard.
`--qr` | `-q` | Write QR code to file.
`--password` | `-o` | Only display the token. For use in scripts.
