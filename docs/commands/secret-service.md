# `secret-service` command

**Linux only.** The `secret-service` command runs a daemon that implements the
[freedesktop.org Secret Service API](https://specifications.freedesktop.org/secret-service/latest/)
(`org.freedesktop.secrets`) on top of your gopass store. Desktop applications
that speak this API — Firefox, Chromium, VS Code, Electron apps,
NetworkManager, `secret-tool`, … — then store their secrets in the same
GPG-encrypted, git-backed store as the gopass CLI, instead of GNOME Keyring or
KDE Wallet.

This is a pure-Go implementation; no CGo and no extra runtime dependencies are
required. See [ADR A-11](../adr/A-11-secret-service.md) for the design and its
current limitations.

## Usage

```
gopass secret-service serve [--replace] [--prefix=secret-service] [--notify-on-access]
gopass secret-service status
gopass secret-service install
gopass secret-service uninstall
```

### `serve`

Starts the daemon and blocks until interrupted (e.g. by `Ctrl+C`).

| Flag                 | Description                                                              |
|----------------------|--------------------------------------------------------------------------|
| `--replace`          | Replace any existing `org.freedesktop.secrets` provider (e.g. GNOME Keyring). |
| `--prefix`           | gopass path prefix under which secrets are stored. Default `secret-service`. |
| `--notify-on-access` | Send a desktop notification whenever a secret is read.                    |

Secrets are stored below the configured prefix, for example:

```
~/.password-store/
└── secret-service/
    ├── _aliases.gpg      # alias -> collection map
    └── default/
        ├── _meta.gpg     # collection metadata
        └── i<hex>.gpg    # secret items
```

Because the items live in gopass, they can be inspected and managed with the
regular commands, e.g. `gopass show secret-service/default/i<hex>`.

The spec's volatile `session` collection is *not* stored in the password store.
It is served from memory (backed by the Linux kernel keyring) and disappears
with the daemon, so transient credentials never become GPG files or git commits.

### `status`

Reports whether some process currently owns the `org.freedesktop.secrets` bus
name.

### `install` / `uninstall`

`install` writes a systemd user unit and a D-Bus session activation file for the
current user:

* `~/.config/systemd/user/gopass-secret-service.service`
* `~/.local/share/dbus-1/services/org.freedesktop.secrets.service`

(the XDG base directories are honored). Afterwards run:

```sh
systemctl --user daemon-reload
systemctl --user enable --now gopass-secret-service
```

`uninstall` removes both files again.

## Replacing GNOME Keyring

Many desktop environments start GNOME Keyring, which grabs
`org.freedesktop.secrets` at login. To use gopass instead, either start the
daemon with `--replace`, or disable GNOME Keyring's secret service component:

```sh
cp /etc/xdg/autostart/gnome-keyring-secrets.desktop ~/.config/autostart/
echo "Hidden=true" >> ~/.config/autostart/gnome-keyring-secrets.desktop
```

## Known issues

* **Startup may hang when `gpg-agent` uses `pinentry-gnome3`.** If pinentry
  consults libsecret for a cached passphrase while the daemon is still starting,
  the two can deadlock. Add `no-allow-external-cache` to `~/.gnupg/gpg-agent.conf`
  and run `gpgconf --kill gpg-agent`, or use a pinentry that does not use
  libsecret.
* **Lock state is in-memory only.** Locking a collection makes `GetSecret`
  return `IsLocked`, but it does not evict the passphrase cached by `gpg-agent`.
* **Out-of-process changes need a restart.** Item metadata is cached in memory
  to avoid decrypting the store on every lookup. Secrets added or changed by the
  gopass CLI while the daemon runs are still listed (the store listing is not
  cached), but their metadata is only re-read after the daemon rewrites or
  forgets the item.
* **Linux only.** The `install`/`uninstall`/`serve` subcommands exist on other
  platforms but fail with a clear message.
