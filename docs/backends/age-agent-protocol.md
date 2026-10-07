# The gopass age agent protocol

This document specifies the protocol spoken by the gopass age agent over
its unix socket. It is the reference for client implementors, including
third-party agents and tools that talk to a running gopass.

The protocol is a line-oriented, strictly request-response text protocol:
the client sends one command per line, the agent answers with exactly one
response line and never speaks on its own.

## Transport and framing

- Transport is a unix domain socket (see [socket location](#socket-location-and-permissions)).
- One request line yields exactly one response line, terminated by `\n`.
- A connection may carry any number of commands sequentially; it ends on
  `quit`, on EOF, or when the client closes it. The official gopass client
  currently opens a fresh connection per command (with one exception: it
  fetches `status` and `hello` over a single connection), but clients are
  free to reuse connections.
- The maximum request line length is agent-defined and advertised via the
  `maxline` token (16 MiB in the official agent). Clients MUST NOT exceed
  the advertised limit; legacy agents, which advertise nothing, accepted
  only 64 KiB. Response lines never exceed the scale of their payload
  (decrypt plaintext is never larger than its ciphertext).
- Binary payloads (ciphertext, plaintext) are encoded with standard,
  padded base64 (`base64.StdEncoding`) on a single line.

## Socket location and permissions

The socket path is resolved from the environment, in this order:

| Condition                           | Socket path                                                  |
| ----------------------------------- | ------------------------------------------------------------ |
| `GOPASS_HOMEDIR` set                | `$GOPASS_HOMEDIR/.run/gopass-age-agent.sock` (all platforms) |
| `XDG_RUNTIME_DIR` set (non-windows) | `$XDG_RUNTIME_DIR/gopass/gopass-age-agent.sock`              |
| neither set (non-windows)           | `~/.run/gopass/gopass-age-agent.sock`                        |
| neither set (windows)               | `%LOCALAPPDATA%\gopass\gopass-age-agent.sock`                |

The agent creates the socket directory with `0700` and the socket itself
with `0600`, owned by the running user. Clients MUST verify owner and
permissions before connecting (see the
[security documentation](../security.md)).

## Commands

Commands are the first whitespace-separated token of a line; everything
after the first space forms the argument list. Commands are
case-sensitive.

### `hello [client version]`

Capability negotiation. The response payload is the space-separated
capability token list; see [capability negotiation](#capability-negotiation)
below. The optional argument is a self-reported client version string
(e.g. `gopass/1.19.0`) and carries no semantics — agents ignore it beyond
diagnostics.

### `ping`

Liveness probe. Answers `OK`. Takes no arguments.

### `status`

Reports the lock state. Answers `OK` when unlocked and `OK locked` when
locked. Takes no arguments.

### `identities <id> [<id> ...]`

Replaces the agent's in-memory identity set with the given age identity
strings (as produced by gopass: `AGE-SECRET-KEY-1...` private keys, plugin
identities and gopass's `identity|recipient` composite form). Identities
are separated by **spaces, not newlines** — a newline-separated send
misframes every identity after the first, which the agent then rejects as
an unknown command.

- Requires at least one argument, otherwise `ERR missing identities`.
- Unparseable identities yield `ERR failed to parse identities: ...` and
  leave the previous identity set untouched.
- Success answers `OK`.
- Sending identities while locked is allowed: they are stored, but
  decryption stays refused until `unlock`.

### `decrypt <base64 ciphertext>`

Decrypts the ciphertext with the currently loaded identities. The
passphrase-protected keyring and all key material stay inside the agent
process; only the plaintext is returned.

- Requires exactly one argument, otherwise `ERR missing ciphertext`.
- Answers `OK <base64 plaintext>` on success.
- Errors: `ERR failed to decode ciphertext: ...` (not valid base64),
  `ERR agent is locked` (see [state model](#state-model)), and
  `ERR failed to decrypt: ...` (no matching identity or damaged input).

### `ssh-identity <base64 SSH private key>`

Appends an unencrypted OpenSSH private key (openssh-key-v1 format, base64)
to the agent's identity set. Unlike `identities` this **adds to** the
loaded identities instead of replacing them.

- Requires exactly one argument, otherwise `ERR missing SSH identity`.
- Errors: `ERR failed to decode SSH identity: ...` (not valid base64),
  `ERR failed to parse SSH identity: ...`, `ERR cannot add external SSH
identities to a keyring session` (a source-bound session is active), and
  `ERR agent is stopped`.
- Success answers `OK`.

### `session <subcommand> [...]`

Optional source-bound session protocol, used by protected (hardware- or
passphrase-backed) keyrings. A session binds the loaded identities to a
client-supplied `source` string; while a session is active, the plain
`decrypt` command is refused (`ERR session decryption requires source`)
and `ssh-identity` cannot append identities. Subcommands:

- `session begin` — starts a session negotiation, answers
  `OK <unlock token>`.
- `session load <source> <timeout> <base64 identities>` — replaces the
  identity set bound to `source`; the timeout is in seconds.
- `session load-generation <source> <timeout> <token> <base64 identities>`
  — like `load`, but the token from a preceding `begin` must still be
  valid; any lock or reload invalidates it.
- `session decrypt <source> <base64 ciphertext>` — decrypts bound to
  `source`; answering with a mismatched source clears the session.

Locking (including the auto-lock timer) invalidates the session and its
token. Clients that do not use protected keyrings can ignore this command
entirely.

### `lock`

Locks the agent: clears every cached identity from memory and sets the
locked flag. Answers `OK`. Takes no arguments.

### `unlock`

Clears the locked flag only. It does **not** restore identities — they
are wiped by `lock` — so a client must re-send `identities` before
decrypting again. Answers `OK`. Takes no arguments.

### `set-timeout <seconds>`

Sets the inactivity auto-lock timeout in seconds (see
[state model](#state-model)). `0` disables the timer.

- Requires exactly one integer argument, otherwise
  `ERR missing timeout` / `ERR failed to parse timeout: ...`.
- Answers `OK`.

### `quit`

Shuts the agent down. The agent answers `OK`, closes the connection,
removes the socket file and exits. Takes no arguments.

## State model

The agent keeps three pieces of state:

- an identity set, replaced wholesale by every successful `identities`
  command and cleared by `lock`;
- a locked flag, set by `lock` (including the auto-lock timer) and
  cleared by `unlock`. While locked, `decrypt` is refused; `identities`
  still works;
- an optional idle timer, armed by `set-timeout`. Every `decrypt` resets
  it; when it fires, the agent locks itself exactly as if `lock` had been
  sent;
- an optional `source` binding established by the `session` subcommands:
  while set, only `session decrypt` can decrypt (see `session`).

A typical client session is: `hello` (optional) → `identities` →
`decrypt`\* → `lock`/`quit`.

## Capability negotiation

```text
> hello gopass/1.19.0
< OK decrypt identities lock ping quit session set-timeout ssh-identity status unlock maxline=16777216 version=1.19.0
```

- Tokens are space-separated on one response line, either bare (`decrypt`)
  or `key=value` (`maxline=16777216`). Clients MUST ignore tokens they do
  not recognize. When a token appears more than once, the first
  occurrence wins; redefinitions (including a bare token repeating a
  `key=value` one) MUST be ignored.
- _Provisional, pending resolution of the [open questions in
  #3624](https://github.com/gopasspw/gopass/issues/3624):_ the exact token
  grammar (`key[=value]`) and whether the client version argument is part
  of `hello`. If these change, new encodings are introduced additively per
  the immutability rule below.
- Forks and third-party implementations can add namespaced tokens
  (`something@example.com`), following the SSH agent extension convention
  of RFC 9987 §5.8.
- A pre-`hello` (legacy) agent answers `hello` with `ERR`; clients detect
  this by the error's existence alone (never by matching the error text)
  and keep the pre-hello behaviour.
- The `version=` token is diagnostics-only and must never be used to gate
  behaviour.

## Error handling

Every failure is a single line of the form `ERR <message>`. The message
text is a human-readable diagnostic and is **not** part of the protocol:
clients MUST decide by the presence of `ERR`, never by matching the
message text, which may change between releases. One exception exists
today (`agent is locked`, which the gopass client matches for a specific
user prompt) and is tracked as an open question in #3624.

## Protocol invariants

- An unrecognized command MUST yield a single-line `ERR`, MUST NOT change
  any state and MUST NOT terminate the connection.
- Command and token semantics are immutable: changed semantics are
  introduced under new names, existing ones are never redefined. In-place
  changes are only acceptable to fix behaviour that was already broken.
- Absence of a capability token means "assume legacy behaviour".

## Example session

```text
> hello
< OK decrypt identities lock ping quit session set-timeout ssh-identity status unlock maxline=16777216 version=1.19.0
> identities AGE-SECRET-KEY-1EXAMPLE...
< OK
> decrypt <base64 ciphertext>
< OK <base64 plaintext>
> status
< OK
> lock
< OK
> decrypt <base64 ciphertext>
< ERR agent is locked
> quit
< OK
```
