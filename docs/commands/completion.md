# `completion` commands

The `completion` commands print a shell completion script for the requested
shell. Source the output in your shell to enable auto completion for `gopass`.

## Synopsis

```bash
gopass completion bash
gopass completion zsh
gopass completion fish
gopass completion openbsdksh
```

## Subcommands

| Subcommand   | Description                            |
|--------------|----------------------------------------|
| `bash`       | Print the completion script for bash.  |
| `zsh`        | Print the completion script for zsh.   |
| `fish`       | Print the completion script for fish.  |
| `openbsdksh` | Print the completion script for OpenBSD's ksh. |

## Details

Redirect the output directly to the location your shell expects. Do not pipe it
through other commands, as that may corrupt the script.

Note: The completion scripts checked into the repository root
(`bash.completion`, `zsh.completion`, `fish.completion`) are generated with
`make completion`. Regenerate them after changing the command or flag
definitions.

See [docs/setup.md](../setup.md) for per-shell installation instructions.

## Exit codes

| Code | Meaning                       |
|-----:|-------------------------------|
| 0    | Completion script printed     |

See [docs/exit-codes.md](../exit-codes.md) for the full table.
