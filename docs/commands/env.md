# `env` command

> **Security warning:** Any mode that injects secrets as environment variables
> (`default` and `--exec`) exposes those values to every process that can read
> `/proc/<pid>/environ` on Linux or `ps eww` on macOS for the entire lifetime of
> the subprocess. If secret exposure via the process environment is a concern,
> use `--stdin` (single secret) or `--file` (ramdisk-backed temp file) instead.

The `env` command runs a binary as a subprocess with a pre-populated environment.
The environment of the subprocess is populated with a set of environment variables corresponding
to the secret subtree specified on the command line.

## Synopsis

```
$ gopass env [options] secret-or-prefix command [args...]
```

## Flags

| Flag | Description |
|------|-------------|
| `--keep-case` / `-kc` | Do not uppercase the environment variable name (default: names are uppercased) |
| `--stdin` | Pipe the secret's password to the subprocess's **stdin** instead of injecting it into the environment |
| `--file` | Write each secret to a ramdisk temporary file and export `KEY_FILE=/path/to/file` instead of `KEY=value` |
| `--exec` | Replace the current gopass process with the subprocess via `exec(3)` (Linux/macOS only; not supported on Windows) |

`--stdin`, `--file`, and `--exec` are mutually exclusive.

## Modes

### Default (env injection)

```
$ gopass env db/prod psql -U admin mydb
```

Each secret key under the given prefix is exported as an uppercased environment variable
(`DB_PASSWORD=secret`). The subprocess runs as a **child process** of gopass.

> **Security caveat:** The injected variables are visible in `/proc/<pid>/environ` on Linux
> and via `ps eww` on macOS for as long as the subprocess is running. Any local process
> with read access to `/proc` (including other user processes on a shared system) can
> observe these values. Prefer `--stdin` or `--file` for sensitive credentials.

### `--stdin`

```
$ gopass env --stdin db/password gpg --passphrase-fd 0 --decrypt file.gpg
```

The secret's password is written to the subprocess's **stdin**. No environment variable is
set, so the secret is never visible in `/proc/<pid>/environ` or `ps` output.

> **Caveats:**
> - Only works with a **single** secret. Passing a directory/prefix is not supported.
> - The subprocess must be designed to read credentials from stdin (e.g. via
>   `--passphrase-fd 0`, `--password-stdin`, or similar flags). Programs that do
>   not read from stdin at all will hang waiting for input.
> - The password is written to stdin **without** a trailing newline. Most programs
>   (e.g. `gpg --passphrase-fd 0`) accept this, but a small number of programs
>   require a newline-terminated passphrase. Wrap with `printf '%s\n'` or a shell
>   heredoc in those cases.
> - The subprocess's own stdin is replaced by the secret. If the subprocess also
>   needs interactive stdin input from the user, this mode is not suitable.

### `--file`

```
$ gopass env --file db/prod psql -U admin mydb
```

Each secret is written to a ramdisk-backed temporary file (on Linux via `/dev/shm`, on macOS
via a RAM disk). The environment variable is set to `KEY_FILE=/path/to/tmpfile` following the
`*_FILE` convention used by Docker Compose and HashiCorp Vault. All temporary files are
removed automatically when the subprocess exits.

> **Caveats:**
> - On **Windows** there is no ramdisk support; temp files fall back to the regular OS
>   temporary directory, which resides on a persistent disk. The secret may be
>   recoverable from disk after deletion.
> - Temp files are deleted (not shredded) on exit. On SSDs with wear leveling,
>   journaling filesystems, or copy-on-write filesystems (ZFS, Btrfs, APFS) the
>   original data may persist in reallocated blocks or journal entries.
> - The `KEY_FILE` variable itself is visible in `/proc/<pid>/environ`, though it
>   exposes only the file path, not the secret value.

### `--exec`

```
$ gopass env --exec db/prod psql -U admin mydb
```

Uses `exec(3)` (via `syscall.Exec`) to **replace** the current gopass process with the
subprocess. Because gopass disappears from the process table entirely, there is no lingering
parent process whose `/proc/<pid>/environ` can be observed.

> **Caveats:**
> - **Not supported on Windows.**
> - The injected variables are still present in the subprocess's own `/proc/<pid>/environ`.
>   `--exec` eliminates the *gopass parent* from the process table but does not prevent
>   the subprocess from exposing the variables.
> - Because the gopass process is replaced, any deferred cleanup (e.g. temp files from
>   a previous `--file` call in the same invocation) will **not** run after the subprocess
>   exits.

## Choosing a mode

| Scenario | Recommended mode |
|----------|------------------|
| Single secret for a program that reads stdin | `--stdin` |
| Multiple secrets or program does not support stdin | `--file` |
| Program requires env variables and secrets are low-sensitivity | default or `--exec` |
| Must avoid a lingering gopass process in `ps` output | `--exec` |

## Security summary

| Mode | Secret in env? | Visible in `/proc`? | Ramdisk? |
|------|---------------|---------------------|----------|
| Default | Yes (`KEY=value`) | Yes (subprocess PID) | No |
| `--exec` | Yes (`KEY=value`) | Yes (subprocess PID) | No |
| `--file` | No (path only) | File path only | Yes (Linux/macOS) |
| `--stdin` | No | No | N/A |

