# Gopass Hooks

`gopass` exposes some hook-able events during it's invocation lifecycle. This allows uses to inject additional functionality of perform addition logging.

## Hook API

All hooks are subject to the follwing constraints:

* Hooks do not inherit `STDIN` or `STDOUT` from the parent process.
* Hooks do inherit `STDERR` from the parent process and may use it to print anything they want.
* Hooks always run from the `password-store` directory.
* Hooks are run with the `GOPASS_HOOK=1` in their environment and with `GOPASS_CONFIG_DIR` set to the configuration directory the original `gopass` command was started.
* An exit from a hook (or execution failure) cases the entire `gopass` command to fail.
* Hooks have at most one minute to complete.

## Reentrancy

`gopass` hooks are non-reentrant by default.

For example take this setup:

```text
[rm]
  post-hook: ~/.config/gopass/hooks/post-rm.sh
```

```shell
# ~/.config/gopass/hooks/post-rm.sh
gopass rm some-other-entry
```

and finally

```shell
$ gopass rm foo
```

In this scenario users should expect `post-rm.sh` to be executed exactly once on `gopass rm foo`.

But in fact it would be run twice: Once on `gopass rm foo` and once on `gopass rm some-other-entry`, i.e. hooks would reenter themselves when they try to use `gopass` internally.

Since most users would find this confusing `gopass` do not do this by default. However if you really need to allow reentrant hooks you currently have one workaround:

* You can `unset` the `GOPASS_HOOK` environment variable in your hook before running `gopass` internally.
