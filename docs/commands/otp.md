# `otp` command

The `otp` command generates TOTP tokens from an OTP URL (`otpauth://`).
The command tries to parse the password and the totp fields as an OTP URL.

Note: HTOP is currently not supported.

Note: If `show.safecontent` is enabled, OTP URLs are hidden from the `show` command.

## Modes of operation

* Generate the current TOTP token from a valid OTP URL
* Snip the screen to add a TOTP QR code as an OTP field to an entry.

## Flags

| Flag         | Aliases | Description                                                              |
|--------------|---------|--------------------------------------------------------------------------|
| `--clip`     | `-c`    | Copy the time-based token into the clipboard.                            |
| `--qr`       | `-q`    | Write QR code to file.                                                   |
| `--chained`  | `-p`    | chain the token to the password                                          |
| `--password` | `-o`    | Only display the token. For use in scripts.                              |
| `--snip`     | `-s`    | Try and find a QR code in the screen content to add as OTP to the entry. |
