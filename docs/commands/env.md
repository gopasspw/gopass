# `env` command

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
(`DB_PASSWORD=secret`). The subprocess runs as a **child process** of gopass. The secret
values are visible in `/proc/<pid>/environ` on Linux and via `ps eww` on macOS while the
subprocess is running.

### `--stdin`

```
$ gopass env --stdin db/password gpg --passphrase-fd 0 --decrypt file.gpg
```

The secret's password is written to the subprocess's **stdin**. No environment variable is
set. Only works with a single secret (not a prefix/directory).

### `--file`

```
$ gopass env --file db/prod psql -U admin mydb
```

Each secret is written to a ramdisk-backed temporary file (on Linux via `/dev/shm`, on macOS
via a RAM disk). The environment variable is set to `KEY_FILE=/path/to/tmpfile` following the
`*_FILE` convention used by Docker Compose and HashiCorp Vault. All temporary files are
removed automatically when the subprocess exits.

### `--exec`

```
$ gopass env --exec db/prod psql -U admin mydb
```

Uses `exec(3)` (via `syscall.Exec`) to **replace** the current gopass process with the
subprocess. Because gopass disappears from the process table entirely, there is no lingering
parent process whose `/proc/<pid>/environ` can be observed. The subprocess inherits the
current environment plus the injected secret variables.

> **Note:** `--exec` is not supported on Windows.

## Security note

The default mode and `--exec` both expose secret values as environment variables, which are
readable via `/proc/<pid>/environ` on Linux and `ps eww` on macOS. If this is a concern,
prefer `--stdin` (for single-secret use-cases) or `--file` (which stores secrets off the
main environment on a ramdisk).

