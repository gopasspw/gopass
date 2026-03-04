# Use case: GPaste Clipboard management system

## Summary

On Linux one might want to use the [GPaste](https://github.com/Keruspe/GPaste) clipboard manager. Using the `GOPASS_CLIPBOARD_COPY_CMD` and `GOPASS_CLIPBOARD_CLEAR_CMD` environment variables one can instruct gopass to use the GPaste client directly. This hides passwords when viewed in the manager and removes them by name after the timeout.

## Usage

### Helper scripts

Both environment variables expect a path to an executable or the name of an executable in the `PATH` environment variable. The executables receive the name of the password as the first argument and the password (copy) or its checksum (clear) in `STDIN`. To use the GPaste client one has to use helper scripts like this:

`~/.local/scripts/gopass_clipboard_copy_cmd.sh`
```sh
#!/bin/sh

# gpaste-client will use the password in /dev/stdin
gpaste-client add-password "$1"
```

`~/.local/scripts/gopass_clipboard_clear_cmd.sh`
```sh
#!/bin/sh

gpaste-client delete-password "$1"
```

Make sure both are executable: `chmod +x ~/.local/scripts/gopass_clipboard_{copy,clear}_cmd.sh`

### Setting the environment variables

#### Shell

You can set the environment variables in the `.profile` file of your shell, for example:

`~/.bash_profile`
```sh
# [...]
export GOPASS_CLIPBOARD_COPY_CMD="$HOME/.local/scripts/gopass_clipboard_copy_cmd.sh"
export GOPASS_CLIPBOARD_CLEAR_CMD="$HOME/.local/scripts/gopass_clipboard_clear_cmd.sh"
# [...]
```

#### Graphical environment

If you are using X11 you can set the above in `~/.xprofile`.

On Wayland one may use systemd user environment variables:

`~/.config/environment.d/gopass.conf`
```sh
GOPASS_CLIPBOARD_COPY_CMD="$HOME/.local/scripts/gopass_clipboard_copy_cmd.sh"
GOPASS_CLIPBOARD_CLEAR_CMD="$HOME/.local/scripts/gopass_clipboard_clear_cmd.sh"
```

A reboot might be required.