# A-11: Integrate `org.freedesktop.secrets` D-Bus Service

**Status:** implemented
**Source:** [GitHub Issue #3434](https://github.com/gopasspw/gopass/issues/3434)

---

## Background

The [freedesktop.org Secret Service specification](https://specifications.freedesktop.org/secret-service/latest/)
defines a D-Bus API that desktop applications (Firefox, Chrome, VS Code, Electron apps,
NetworkManager, `secret-tool`, …) use to store and retrieve secrets. On most systems it is served
by GNOME Keyring or KDE Wallet. Users who keep their passwords in gopass currently maintain two
separate secret stores.

The request is to add a `gopass secret-service` daemon subcommand that implements this D-Bus API on
top of the existing gopass store, so that GUI and CLI applications write their secrets into the
same GPG-encrypted, git-backed password store.

---

## Prior Art

Two reference implementations inform the design:

| Project | Language | License | Notes |
|---------|----------|---------|-------|
| [nikicat/gopass-secret-service](https://github.com/nikicat/gopass-secret-service) | Go (91 %) | MIT | Standalone daemon built on the gopass Go API. **Most relevant.** |
| [grimsteel/pass-secret-service](https://github.com/grimsteel/pass-secret-service) | Rust | GPL-3.0 | Standalone daemon for `pass`. Pure-Rust D-Bus. |

`nikicat/gopass-secret-service` implements the full spec in Go and consumes the **gopass Go API**
(`github.com/gopasspw/gopass/pkg/gopass/api`) directly — it does **not** shell out to the `gopass`
CLI. It additionally routes the spec's volatile `session` collection to an in-kernel keyring and
caches item metadata so that `SearchItems` does not decrypt the whole store on every lookup. Both
lessons are folded into the design below. Since the integrated version re-uses the same API, the
practical difference is **packaging** — a subcommand of the existing `gopass` binary versus a
separate daemon binary — not the storage access path. Users who prefer the standalone daemon remain
free to run it.

---

## Decision

Implement `gopass secret-service` as:

1. A new **Linux-only** subcommand of the main `gopass` binary.
2. A long-running **daemon** that acquires `org.freedesktop.secrets` on the D-Bus session bus and
   serves the full Secret Service interface.
3. Uses `github.com/godbus/dbus/v5` (already a direct dependency in `go.mod` at v5.2.2, pure Go,
   BSD-2 licensed).
4. Crypto uses the standard library (`crypto/aes`, `crypto/cipher`, `crypto/rand`, `crypto/sha256`,
   `math/big`) plus `golang.org/x/crypto/hkdf` (already a direct dependency at v0.55.0) for the DH
   key schedule — no CGo, no new external dependencies.
5. Secrets are stored under a configurable gopass path prefix (default: `secret-service`).
6. The spec's volatile `session` collection is served from memory (backed by the Linux kernel
   keyring) and is never written to the store.
7. Item metadata is cached in memory and invalidated on every local write, so `SearchItems` does
   not trigger one GPG decryption per item per lookup.

---

## Feasibility Summary

* **Pure-Go, zero-CGo**: `godbus/dbus/v5` (v5.2.2) is already a direct dependency. Crypto uses
  `crypto/aes`, `crypto/cipher`, `crypto/rand`, `crypto/sha256` and `math/big` from the standard
  library, plus `golang.org/x/crypto/hkdf` (v0.55.0, already a direct dependency) for the DH key
  schedule.
* **No new external dependencies**: only the standard library plus the existing `godbus/dbus/v5`,
  `golang.org/x/crypto` and `golang.org/x/sys` dependencies are needed. (`golang.org/x/sys` serves
  the kernel-keyring backing store for the `session` collection.) Item IDs are generated with
  `crypto/rand`, so no UUID dependency is pulled in.
* **License-compatible**: `godbus/dbus/v5` is BSD-2, `golang.org/x/*` is BSD-3 — both MIT
  compatible per `.license-lint.yml`.
* **Linux-only build tag**: The entire feature is gated with `//go:build linux` (same pattern as
  `internal/notify/notify_dbus.go` and `pkg/clipboard/unclip_linux.go`).
* **Architectural risk**: gopass is a short-lived CLI tool; this feature requires a persistent
  daemon process. This is handled by a blocking `gopass secret-service serve` subcommand (the user
  manages the lifecycle via systemd or similar).

---

## Implementation Status

The feature is implemented in `internal/secretservice` (all files carry a
`//go:build linux` constraint):

| Area | Status |
|------|--------|
| Pure-Go crypto sessions (`plain`, `dh-ietf1024-…`) | implemented, unit-tested (`internal/secretservice/crypto`) |
| D-Bus service, collections, items, sessions, aliases, search | implemented |
| Item metadata cache for `SearchItems` | implemented |
| CLI `gopass secret-service serve` / `status` / `install` / `uninstall` + non-Linux stub | implemented |
| Volatile `session` collection (kernel keyring) | implemented |
| Alias object paths (`/org/freedesktop/secrets/aliases/*`) | implemented |
| `CreateItem(replace=true)` | implemented |
| Item `Locked` property and signal coverage | implemented |
| Prompt objects | not needed — every operation completes immediately with `"/"` |
| systemd user unit / D-Bus activation `install` | implemented (`contrib/secret-service/`) |
| Integration tests against a real bus / libsecret client | implemented (`tests/secret_service_test.go`, `tests/secret_service_dh_test.go`, `tests/secret_service_libsecret_test.go`) |

Verification: `TestSecretServiceSecretTool` drives the daemon with `secret-tool`
(a real libsecret client, which selects the DH transport) and round-trips a
secret in both directions; `TestSecretServiceRoundTrip` and
`TestSecretServiceDHRoundTrip` cover both transport algorithms against a private
`dbus-daemon`; `TestSecretServiceSessionCollection` asserts that session secrets
never reach the store.

### Alias object paths

libsecret does **not** call `ReadAlias` to resolve a collection alias. It derives
`/org/freedesktop/secrets/aliases/<alias>` locally and then calls `CreateItem`
and friends on that path directly. The service therefore exports a `Collection`
object at every alias path in addition to the canonical
`/org/freedesktop/secrets/collection/<name>` path, and emits
`ItemCreated`/`ItemChanged`/`ItemDeleted` from all of them. Without this,
`secret-tool store` fails with `UnknownInterface: Object does not implement the
interface 'org.freedesktop.Secret.Collection'`.

### Item IDs and D-Bus object paths

A D-Bus object path element must match `[A-Za-z0-9_]`. Items generated by the
service (`i` + lowercase hex) are safe, but an item may also be created out of
band by the gopass CLI (e.g. `gopass insert secret-service/default/cli-inserted`),
producing a name with hyphens that cannot be used as a path element. Such names
are hex-encoded with an `x` prefix (`ItemDBusID`), and the mapping is reversed
exactly by `ItemNameFromDBusID`. Names that are already safe and do not start
with `x` are used verbatim, so generated IDs keep their historic layout and the
two encodings can never collide. Collection names that are not path-safe are
skipped when listing (`isPathSafe`).

Item properties, sessions, aliases and the volatile `session` collection are all
implemented. The remaining known limitations are:

1. **Prompt objects.** Every operation completes immediately, so all methods
   return the null prompt path (`"/"`). No `Prompt` objects are exported.
2. **Lock state is in-memory only.** Locking does not evict the passphrase
   cached by `gpg-agent`.
3. **Metadata cache invalidation is local.** Changes made by another gopass
   process require a daemon restart to be reflected in item metadata.
4. **No `install --gnome-keyring` automation.** The `install` subcommand writes
   the unit and activation files but does not disable GNOME Keyring; the user
   must do that manually (see the command documentation).

---

## D-Bus Interface Mapping

All objects live under the well-known service name `org.freedesktop.secrets`.

| Interface | Object path | Implementation type |
|-----------|-------------|---------------------|
| `org.freedesktop.Secret.Service` | `/org/freedesktop/secrets` | `service.Service` |
| `org.freedesktop.Secret.Collection` | `/org/freedesktop/secrets/collection/{name}` | `service.Collection` |
| `org.freedesktop.Secret.Item` | `/org/freedesktop/secrets/collection/{name}/{id}` | `service.Item` |
| `org.freedesktop.Secret.Session` | `/org/freedesktop/secrets/session/{id}` | `service.Session` |
| `org.freedesktop.Secret.Prompt` | `/org/freedesktop/secrets/prompt/{id}` | `service.Prompt` |

---

## Storage Layout in gopass

Secrets are stored under a configurable prefix (`secret-service` by default):

```
~/.password-store/
└── secret-service/
    ├── _aliases.age          # Map of alias → collection name (JSON)
    ├── default/
    │   ├── _meta.age         # Collection metadata (label, created, modified)
    │   └── i<hex>.age        # Secret items
    └── work/
        ├── _meta.age
        └── i<hex>.age
```

Item IDs are random values rendered as `i` followed by lowercase hex (`fmt.Sprintf("i%x", …)`),
**not** hyphenated UUIDs. A D-Bus object path element must match `[A-Za-z0-9_]` and may not contain
hyphens, and the item ID becomes the last element of the item's object path
(`/org/freedesktop/secrets/collection/{name}/{id}`). The reference implementation made the same
choice.

Items with names created out of band by the CLI (for example
`gopass insert secret-service/default/cli-inserted`) contain characters that are
not valid in an object path element. Those names are hex-encoded with an `x`
prefix; see [Item IDs and D-Bus object paths](#item-ids-and-d-bus-object-paths).

Each item secret file uses the standard gopass multi-line format:

```
the-secret-value
---
_ss_label: My GitHub Token
_ss_created: 2026-01-15T10:30:00Z
_ss_modified: 2026-01-15T10:30:00Z
_ss_content_type: text/plain
username: user@example.com
service: github.com
```

The first line is the secret value; `_ss_*` keys are internal metadata; all
other key/value pairs are user-visible item attributes (used for lookup by
`SearchItems`).

---

## The `session` collection

The spec reserves a well-known collection alias, `session`
(`SECRET_COLLECTION_SESSION`), for secrets that must live only for the duration of the login
session and never be persisted. Clients use it for transient credentials.

With a gopass backend this matters more than usual: if transient secrets fell through to the
persistent store, every one of them would become a GPG-encrypted file **and a git commit** — the
opposite of what the client asked for. The `session` collection is therefore **not** mapped to a
gopass subpath. It is served from a volatile, in-process store backed by the Linux kernel keyring
(the daemon's process keyring), so payloads stay in kernel memory and vanish when the daemon exits.

> **Status:** implemented in `internal/secretservice/volatile.go`. If the
> `keyctl` syscalls are unavailable (for example in a sandbox that blocks them),
> the daemon logs a debug message and falls back to an in-memory store.

Consequences for the implementation:

* `ReadAlias("session")` returns the well-known session collection path. The session collection is
  always present and does not appear in `Service.Collections`.
* No `CollectionCreated` / `CollectionDeleted` signals are emitted for it, and it cannot be deleted.
* The kernel-keyring backing store must run on a single dedicated, `runtime.LockOSThread`-pinned OS
  thread: Linux stores the process/session/thread keyring references in the per-task `struct cred`,
  so keyring syscalls must not migrate between threads. The syscalls come from
  `golang.org/x/sys/unix` (already a dependency) — still no CGo.
* On non-Linux platforms the whole feature is compiled out; there is no fallback keyring.

---

## Crypto Sessions

The spec defines two `OpenSession` algorithms:

| Algorithm | Description | Implementation |
|-----------|-------------|----------------|
| `plain` | No transport encryption | Return empty bytes; secret value passed as-is |
| `dh-ietf1024-sha256-aes128-cbc-pkcs7` | DH key exchange + AES-128-CBC | stdlib `crypto/aes`, `crypto/sha256`, `math/big` + `golang.org/x/crypto/hkdf` |

The `plain` algorithm is secure because D-Bus session bus traffic is carried over a local UNIX
socket with kernel-enforced access control. Implementing `dh-ietf1024-sha256-aes128-cbc-pkcs7` is
required for compatibility with `libsecret`-based applications.

DH parameters: [RFC 3526](https://www.rfc-editor.org/rfc/rfc3526) 1024-bit MODP group 2
(the reference implementation uses the RFC 2409 group 2 prime; the two are identical).
The server:
1. Generates a DH ephemeral key pair on the MODP-1024 group.
2. Receives the client's public key in `OpenSession`.
3. Computes `shared = clientPub^serverPriv mod p`.
4. Left-pads `shared` to 128 bytes (a known pitfall — see grimsteel commit c781717).
5. Derives the AES key with **HKDF-SHA256** over the padded shared secret, using a NULL salt and
   empty info, and takes the first 16 bytes: `hkdf.New(sha256.New, padded, nil, nil)`. This is
   *not* a bare `SHA256(shared)[0:16]`; libsecret clients derive the same key the same way and
   would reject a non-HKDF key.
6. Returns the server's public key, itself left-padded to 128 bytes.
7. Each secret returned has its own random 16-byte IV prepended to the ciphertext.

---

## Package Structure

All new code lives under `internal/secretservice/` (Linux-only files) and integrates
via a new `secretservice_linux.go` action handler shim.

```
internal/secretservice/
├── doc.go                    # Package doc + interface names, Secret struct
├── service.go                # org.freedesktop.Secret.Service implementation
├── collection.go             # org.freedesktop.Secret.Collection implementation
├── item.go                   # org.freedesktop.Secret.Item implementation
├── session.go                # Session lifecycle + crypto dispatch
├── store.go                  # Adapter between Secret Service and gopass API
├── volatile.go               # Volatile `session` collection (kernel keyring)
├── units.go                  # systemd unit / D-Bus activation install helpers
├── paths.go                  # Object path <-> store path mapping, ID encoding
├── introspection.go          # Static introspection XML
├── errors.go                 # D-Bus error definitions (org.freedesktop.Secret.Error.*)
├── crypto/
│   ├── crypto.go             # Session interface + factory
│   ├── plain.go              # "plain" algorithm
│   └── dh.go                 # "dh-ietf1024-sha256-aes128-cbc-pkcs7"
└── *_test.go                 # Unit tests (in-memory gopass.Store, no GPG needed)
```

CLI integration:

```
internal/action/
├── secretservice_linux.go    # secret-service command tree (serve/status/install/uninstall)
└── secretservice_other.go    # Stub for non-Linux platforms that prints "linux only"
```

Systemd / D-Bus activation files (installed by `gopass secret-service install`):

```
contrib/secret-service/
├── org.freedesktop.secrets.service   # D-Bus session activation (Exec=gopass secret-service serve)
└── gopass-secret-service.service     # systemd user unit
```

---

## Implementation Phases

The phases below were the implementation plan; all of them are now implemented
(see [Implementation Status](#implementation-status)). They are retained as a
description of how each part of the spec maps onto the code.

---

### Phase 1 — Foundation: types, errors, session, crypto

**Goal**: acquire the `org.freedesktop.secrets` bus name and negotiate a session.
No collections or items yet; stub implementations may be used.

**Files to create**:
- `internal/secretservice/doc.go`
- `internal/secretservice/types.go`
- `internal/secretservice/errors.go`
- `internal/secretservice/crypto/crypto.go`
- `internal/secretservice/crypto/plain.go`
- `internal/secretservice/crypto/dh.go`
- `internal/secretservice/session.go`
- `internal/secretservice/service.go` (skeleton: bus name, `OpenSession`, `CloseSession`)
- `internal/secretservice/service_test.go`

**`types.go`** — key D-Bus types:
```go
//go:build linux

package secretservice

import "github.com/godbus/dbus/v5"

const (
    ServiceName = "org.freedesktop.secrets"
    ServicePath = dbus.ObjectPath("/org/freedesktop/secrets")
    ServiceIface = "org.freedesktop.Secret.Service"
    CollectionIface = "org.freedesktop.Secret.Collection"
    ItemIface = "org.freedesktop.Secret.Item"
    SessionIface = "org.freedesktop.Secret.Session"
    PromptIface = "org.freedesktop.Secret.Prompt"

    CollectionPathPrefix = "/org/freedesktop/secrets/collection/"
    SessionPathPrefix    = "/org/freedesktop/secrets/session/"
    PromptPathPrefix     = "/org/freedesktop/secrets/prompt/"
)

// Secret is the D-Bus Secret struct as defined in the spec.
// It is the wire format for passing secrets across D-Bus.
type Secret struct {
    Session     dbus.ObjectPath // session used for transport encryption
    Parameters  []byte          // IV (empty for "plain" algorithm)
    Value       []byte          // encrypted (or plaintext) secret value
    ContentType string          // MIME type, e.g. "text/plain; charset=utf-8"
}
```

**`errors.go`** — see spec §11 for all error names. Key ones:
```go
//go:build linux

package secretservice

import "github.com/godbus/dbus/v5"

var (
    ErrNoSession       = dbus.NewError("org.freedesktop.Secret.Error.NoSession", nil)
    ErrNoSuchObject    = dbus.NewError("org.freedesktop.Secret.Error.NoSuchObject", nil)
    ErrIsLocked        = dbus.NewError("org.freedesktop.Secret.Error.IsLocked", nil)
    ErrAlreadyExists   = dbus.NewError("org.freedesktop.Secret.Error.AlreadyExists", nil)
    ErrNotSupported    = dbus.NewError("org.freedesktop.Secret.Error.NotSupported", nil)
)
```

**`crypto/crypto.go`** — session interface:
```go
//go:build linux

package crypto

// CryptoSession is an established client session.
type CryptoSession interface {
    // Decrypt decrypts a Secret.Value using Parameters as IV.
    Decrypt(params, ciphertext []byte) ([]byte, error)
    // Encrypt encrypts plaintext and returns (params/IV, ciphertext).
    Encrypt(plaintext []byte) (params, ciphertext []byte, err error)
}

// NewSession negotiates a new CryptoSession.
// algorithm is one of "plain" or "dh-ietf1024-sha256-aes128-cbc-pkcs7".
// clientInput is the client public key (empty for "plain").
// Returns the CryptoSession and server output (empty for "plain", server pub key for DH).
func NewSession(algorithm string, clientInput []byte) (CryptoSession, []byte, error) { ... }
```

**`crypto/plain.go`**:
```go
//go:build linux

package crypto

type plainSession struct{}

func (plainSession) Decrypt(_, ciphertext []byte) ([]byte, error) { return ciphertext, nil }
func (plainSession) Encrypt(plaintext []byte) ([]byte, []byte, error) { return nil, plaintext, nil }
```

**`crypto/dh.go`** implementation outline:
```go
//go:build linux

package crypto

import (
    "crypto/aes"
    "crypto/cipher"
    "crypto/rand"
    "crypto/sha256"
    "math/big"

    "golang.org/x/crypto/hkdf"
)

// RFC 3526 MODP 1024-bit group 2
var (
    dhPrime, _ = new(big.Int).SetString("FFFFFFFFFFFFFFFFC90FDAA2...", 16) // full 1024-bit prime
    dhGen      = big.NewInt(2)
)

type dhSession struct{ aesKey []byte }

// NewDHSession computes the shared secret and derives the AES key.
// clientPubBytes is the client's public key.
// Returns the dhSession and the server's public key bytes.
func NewDHSession(clientPubBytes []byte) (*dhSession, []byte, error) {
    // 1. Generate server private key (random 128-byte big.Int)
    // 2. Compute serverPub = g^serverPriv mod p
    // 3. Compute shared = clientPub^serverPriv mod p
    // 4. Left-pad shared to 128 bytes (IMPORTANT: see grimsteel/pass-secret-service#24)
    // 5. aesKey = HKDF-SHA256(paddedShared, salt=nil, info=nil)[0:16]
    // 6. Return serverPub left-padded to 128 bytes
    ...
}

func (s *dhSession) Decrypt(params, ciphertext []byte) ([]byte, error) {
    // AES-128-CBC with IV=params, key=s.aesKey, PKCS7 unpadding
}

func (s *dhSession) Encrypt(plaintext []byte) ([]byte, []byte, error) {
    // random 16-byte IV, AES-128-CBC with PKCS7 padding
}
```

**`session.go`**:
```go
//go:build linux

package secretservice

import (
    "fmt"
    "sync"
    "github.com/godbus/dbus/v5"
    "github.com/gopasspw/gopass/internal/secretservice/crypto"
)

type session struct {
    id     string
    path   dbus.ObjectPath
    crypto crypto.CryptoSession
}

type sessionManager struct {
    mu       sync.RWMutex
    sessions map[string]*session
}

func (sm *sessionManager) Open(algorithm string, input []byte) (*session, []byte, error) { ... }
func (sm *sessionManager) Get(path dbus.ObjectPath) (*session, error) { ... }
func (sm *sessionManager) Close(path dbus.ObjectPath) error { ... }
```

**`service.go`** skeleton:
```go
//go:build linux

package secretservice

import (
    "context"
    "github.com/godbus/dbus/v5"
)

// Service implements org.freedesktop.Secret.Service.
type Service struct {
    conn     *dbus.Conn
    sessions *sessionManager
    // collections added in Phase 2
}

// New creates and starts the service.
// It acquires the org.freedesktop.secrets bus name.
func New(ctx context.Context) (*Service, error) {
    conn, err := dbus.SessionBus()
    ...
    reply, err := conn.RequestName(ServiceName, dbus.NameFlagDoNotQueue)
    ...
    svc := &Service{ conn: conn, sessions: &sessionManager{} }
    conn.Export(svc, ServicePath, ServiceIface)
    conn.Export(introspect.NewIntrospectable(svc), ServicePath, "org.freedesktop.DBus.Introspectable")
    return svc, nil
}

// OpenSession implements org.freedesktop.Secret.Service.OpenSession
func (s *Service) OpenSession(algorithm string, input dbus.Variant) (dbus.Variant, dbus.ObjectPath, *dbus.Error) { ... }

// CloseSession (called on Session object) is delegated to sessionManager.
```

**Testing approach for Phase 1**:
- Use `dbus.SessionBusPrivate()` with `conn.Auth(nil)` + `conn.Hello()` to create a private
  peer-to-peer connection for unit tests, no real session bus needed.
- Test `OpenSession("plain", ...)` returns an empty variant and a valid session path.
- Test `OpenSession("dh-ietf1024-sha256-aes128-cbc-pkcs7", clientPub)` and verify a round-trip
  encrypt/decrypt.

---

### Phase 2 — Collections: CRUD and property management

**Goal**: implement `org.freedesktop.Secret.Collection`, map collections to
gopass subpaths under the `secret-service/` prefix. Add `CreateCollection`,
`DeleteCollection`, and supporting `SearchItems` / `ReadAlias` / `SetAlias` on
the Service.

**Files to create/modify**:
- `internal/secretservice/collection.go` (new)
- `internal/secretservice/store.go` (new — gopass API adapter)
- `internal/secretservice/service.go` (extend: CreateCollection, SearchItems, ReadAlias, SetAlias, GetSecrets)

**`store.go`** — adapter between Secret Service and gopass API.
Do NOT invoke the `gopass` CLI. Use `pkg/gopass/api` directly:
```go
//go:build linux

package secretservice

import (
    "context"
    "github.com/gopasspw/gopass/pkg/gopass/api"
)

// Store wraps the gopass API and provides Secret-Service-specific operations.
type Store struct {
    gp     *api.Gopass
    prefix string // default: "secret-service"
}

func NewStore(ctx context.Context, prefix string) (*Store, error) {
    gp, err := api.New(ctx)
    ...
}

// CollectionPath returns the gopass path for a collection.
func (s *Store) CollectionPath(name string) string {
    return s.prefix + "/" + name
}

// ItemPath returns the gopass path for an item.
func (s *Store) ItemPath(collection, id string) string {
    return s.prefix + "/" + collection + "/i" + id
}

// MetaPath returns the gopass path for a collection's metadata.
func (s *Store) MetaPath(collection string) string {
    return s.prefix + "/" + collection + "/_meta"
}

// AliasPath returns the gopass path for the aliases map.
func (s *Store) AliasPath() string {
    return s.prefix + "/_aliases"
}
```

**Collection metadata** is stored as a gopass secret at `MetaPath`:
```
<empty first line>
---
label: Personal
created: 2026-01-15T10:30:00Z
modified: 2026-01-15T10:30:00Z
locked: false
```

**Collection aliases** are stored at `AliasPath` in JSON on the first line:
```json
{"default":"default","login":"default"}
```

**`collection.go`** key methods to implement:
```go
// CreateItem, SearchItems, Delete
// Properties: Items (ao), Label (s), Locked (b), Created (t), Modified (t)
```

D-Bus property access via `org.freedesktop.DBus.Properties.Get/Set/GetAll`.
Use `godbus/dbus/v5`'s `prop` subpackage for property management.

**Signal emission** (required by spec):
- `Collection.ItemCreated(item: o)`
- `Collection.ItemDeleted(item: o)`
- `Collection.ItemChanged(item: o)`
- `Service.CollectionCreated(collection: o)`
- `Service.CollectionDeleted(collection: o)`
- `Service.CollectionChanged(collection: o)`

**Testing approach for Phase 2**:
- Mock `Store` using an in-memory map (no real gopass/GPG needed for tests).
- Test `CreateCollection` → verify gopass path created and D-Bus object exported.
- Test `ReadAlias("default")` before and after creating collections.

---

### Phase 3 — Items: GetSecret, SetSecret, Delete, Search

**Goal**: implement `org.freedesktop.Secret.Item`. Full item CRUD with attribute-based search.

**Files to create/modify**:
- `internal/secretservice/item.go` (new)
- `internal/secretservice/store.go` (extend: item read/write/delete/search)
- `internal/secretservice/collection.go` (extend: CreateItem, SearchItems)

**Item storage format** (in gopass secret first-line + YAML):
```
secret-value-here
---
_ss_label: GitHub Token
_ss_created: 2026-05-01T12:00:00Z
_ss_modified: 2026-05-01T12:00:00Z
_ss_content_type: text/plain; charset=utf-8
username: octocat
server: github.com
```

Rules:
- Keys prefixed with `_ss_` are reserved for internal use.
- All other key/value pairs are item attributes (arbitrary strings, per spec).
- The secret value is the **first line** of the gopass secret (the standard gopass password field).

**`item.go`** key methods:
```go
// GetSecret(session: o) → secret: Secret
// SetSecret(secret: Secret) → nothing
// Delete() → prompt: o (return "/" as prompt path for immediate completion)
// Properties: Locked (b), Attributes (a{ss}), Label (s), Created (t), Modified (t)
```

**Store item search**:
```go
// SearchItems(attrs map[string]string) ([]dbus.ObjectPath, error)
// Lists all items in a collection and filters them by attrs.
```

A naive implementation loads (and therefore **decrypts**) every item's attributes for each search.
With a GPG backend that is one decryption per item per lookup, and libsecret clients issue
`SearchItems` for nearly every operation; with a cold `gpg-agent` cache a single lookup can even
trigger pinentry. The design therefore keeps an **in-memory metadata cache** keyed by gopass path:

* The cache holds only decrypted *metadata* (label, timestamps, content type and user attributes) —
  never the secret payload (`Password()` / `Body()`).
* It is populated lazily on the first read of an item and refreshed opportunistically whenever an
  item is decrypted for another reason.
* It is invalidated on every local mutation (create/update/delete, and prefix-wide when a whole
  collection is removed). There is no TTL; out-of-process changes require a daemon restart.
* The secret value itself is still re-decrypted on demand in `GetSecret` and never cached.

**Locking**: In this implementation lock state is **in-memory only**.
When a collection is "locked", `GetSecret` returns `ErrIsLocked`.
The underlying GPG file is always accessible if the GPG agent has a cached key.

**Testing approach for Phase 3**:
- Create item, verify it appears in `Items` property of collection.
- `SearchItems` with matching/non-matching attributes.
- Round-trip: `SetSecret(plain)` → `GetSecret(plain)` → verify value.
- Round-trip: `SetSecret(dh)` → `GetSecret(dh)` → verify value.
- Delete item → verify gone from `Items` and gopass store.

---

### Phase 4 — Prompt, Lock/Unlock, GetSecrets (batch)

**Goal**: complete the spec. Implement `Prompt` objects, `Service.Lock`,
`Service.Unlock`, `Service.GetSecrets`.

**Files to create/modify**:
- `internal/secretservice/prompt.go` (new)
- `internal/secretservice/service.go` (extend: Lock, Unlock, GetSecrets)

**Prompts**: The spec uses `Prompt` objects for operations that may require user
interaction. For Lock/Unlock in this implementation, the prompt completes
immediately (no actual user interaction needed because GPG-agent handles
passphrase caching). The pattern:

```go
// Unlock(objects []dbus.ObjectPath) → (unlocked []dbus.ObjectPath, prompt dbus.ObjectPath)
// Returns prompt "/" (null prompt) when all objects are already unlocked.
// Returns a real prompt path when any object needs unlocking; the prompt
// signals Completed(dismissed bool, result dbus.Variant) when done.

type Prompt struct {
    path   dbus.ObjectPath
    conn   *dbus.Conn
    action func() ([]dbus.ObjectPath, error)
}

// Prompt.Dismiss() aborts the operation.
// For an immediate-complete prompt: export the object, fire Completed signal in a goroutine.
```

**`Service.GetSecrets`**:
```go
// GetSecrets(items []dbus.ObjectPath, session dbus.ObjectPath) → map[dbus.ObjectPath]Secret
// Batch retrieval. For each path, resolve the item and call its GetSecret logic.
```

**Testing approach for Phase 4**:
- Lock collection, verify `GetSecret` returns `ErrIsLocked`.
- Unlock collection → prompt completes → verify `GetSecret` succeeds.
- `GetSecrets` with mixed locked/unlocked items.
- Prompt dismiss returns correct `dismissed=true`.

---

### Phase 5 — CLI integration: `gopass secret-service`

**Goal**: wire everything into the gopass CLI as a new subcommand.

**Files to create/modify**:
- `internal/action/secretservice_linux.go` (new)
- `internal/action/secretservice_other.go` (new — non-Linux stub)
- `internal/action/commands.go` (add entry point, with build constraints)
- `contrib/secret-service/org.freedesktop.secrets.service` (new — D-Bus activation)
- `contrib/secret-service/gopass-secret-service.service` (new — systemd user unit)
- `docs/commands/secret-service.md` (new — user documentation; expand separately)

**CLI design**:
```
gopass secret-service serve [--replace] [--prefix=secret-service] [--notify-on-access]
gopass secret-service install    # installs systemd unit + D-Bus activation file
gopass secret-service uninstall  # removes systemd unit + D-Bus activation file
gopass secret-service status     # checks whether the service is running
```

**`secretservice_linux.go`** action handler:
```go
//go:build linux

package action

import (
    "context"
    "github.com/gopasspw/gopass/internal/secretservice"
    "github.com/urfave/cli/v3"
)

func (s *Action) SecretService(ctx context.Context, cmd *cli.Command) error {
    prefix := cmd.String("prefix")
    replace := cmd.Bool("replace")
    svc, err := secretservice.New(ctx, secretservice.Config{
        Prefix:  prefix,
        Replace: replace,
    })
    if err != nil {
        return err
    }
    return svc.Serve(ctx) // blocks until ctx is cancelled or fatal error
}
```

**`secretservice_other.go`** stub (for Windows/macOS):
```go
//go:build !linux

package action

import (
    "context"
    "fmt"
    "github.com/urfave/cli/v3"
)

func (s *Action) SecretService(ctx context.Context, cmd *cli.Command) error {
    return fmt.Errorf("secret-service is only supported on Linux")
}
```

**`commands.go`** entry** — add to `GetCommands()`:
```go
{
    Name:  "secret-service",
    Usage: "Run a D-Bus Secret Service daemon backed by gopass",
    Description: "Implements the org.freedesktop.secrets D-Bus API so that " +
        "desktop applications store their secrets in the gopass password store.",
    Before: s.IsInitialized,
    Commands: []*cli.Command{
        {
            Name:   "serve",
            Usage:  "Start the Secret Service daemon",
            Action: s.SecretService,
            Flags: []cli.Flag{
                &cli.BoolFlag{
                    Name:  "replace",
                    Usage: "Replace any existing Secret Service provider (e.g. GNOME Keyring)",
                },
                &cli.StringFlag{
                    Name:  "prefix",
                    Usage: "gopass path prefix for secret-service secrets",
                    Value: "secret-service",
                },
                &cli.BoolFlag{
                    Name:  "notify-on-access",
                    Usage: "Send a desktop notification when a secret is read",
                },
            },
        },
        {
            Name:   "install",
            Usage:  "Install systemd user service and D-Bus activation files",
            Action: s.SecretServiceInstall,
        },
        {
            Name:   "uninstall",
            Usage:  "Remove systemd user service and D-Bus activation files",
            Action: s.SecretServiceUninstall,
        },
        {
            Name:   "status",
            Usage:  "Check whether the Secret Service daemon is running",
            Action: s.SecretServiceStatus,
        },
    },
},
```

**D-Bus activation file** (`org.freedesktop.secrets.service`):
```ini
[D-BUS Service]
Name=org.freedesktop.secrets
Exec=/usr/bin/gopass secret-service serve
```

**systemd user unit** (`gopass-secret-service.service`):
```ini
[Unit]
Description=gopass Secret Service D-Bus daemon
After=graphical-session.target
PartOf=graphical-session.target

[Service]
Type=dbus
BusName=org.freedesktop.secrets
ExecStart=/usr/bin/gopass secret-service serve
Restart=on-failure

[Install]
WantedBy=graphical-session.target
```

**GNOME Keyring conflict resolution**: `gopass secret-service serve --replace` passes
`dbus.NameFlagReplaceExisting` to `conn.RequestName(...)`. Users must also disable GNOME
Keyring's secret service component:
```sh
cp /etc/xdg/autostart/gnome-keyring-secrets.desktop ~/.config/autostart/
echo "Hidden=true" >> ~/.config/autostart/gnome-keyring-secrets.desktop
```
The `install` subcommand should offer to do this automatically.

---

### Phase 6 — Tests and Documentation (expand in a separate prompt)

**Goal**: integration tests, user documentation, and `make test-integration` compatibility.

**Files to create/modify**:
- `internal/secretservice/*_test.go` — expand unit tests
- `tests/secret_service_test.go` — new integration test using `gptest`
- `docs/commands/secret-service.md` — user-facing documentation

**Integration test approach**:
```go
// tests/secret_service_test.go
// Uses gptest.NewGUnitTester to set up a real gopass store.
// Starts the service on a private D-Bus connection (dbus.SessionBusPrivate).
// Uses secret-tool (if available) or direct godbus calls to store/retrieve secrets.
// Verifies:
// - secret created via D-Bus appears in gopass (gopass show secret-service/default/i<hex>)
// - secret created via gopass insert is visible via D-Bus GetSecret
// - attributes are searched correctly by SearchItems
```

**Known caveats to document**:
1. **GPG circular dependency**: If `pinentry-gnome3` tries to check libsecret for cached
   passphrases at startup, a deadlock occurs. Solution: add `no-allow-external-cache` to
   `~/.gnupg/gpg-agent.conf`. Document this prominently (see nikicat/gopass-secret-service#troubleshooting).
2. **Lock state**: Lock/Unlock are in-memory only. Locking a collection does not evict GPG
   agent's cached passphrase.
3. **Linux only**: Ensure the subcommand entry in `commands.go` still compiles on all platforms
   (use the `secretservice_other.go` stub pattern).
4. **Session bus required**: `DBUS_SESSION_BUS_ADDRESS` must be set. The daemon should fail
   gracefully with a clear error if it is not.

---

## Key Implementation Notes for LLM Agents

The following are common pitfalls to avoid:

1. **DH shared-secret left-padding and HKDF**: `shared = clientPub^serverPriv mod p` may produce a
   `big.Int` whose byte representation is shorter than 128 bytes. It must be left-padded with zero
   bytes to exactly 128 bytes before key derivation. The AES key is then
   `HKDF-SHA256(padded, salt=nil, info=nil)[0:16]`, not a bare `SHA256(padded)[0:16]`. The server's
   public key must likewise be left-padded to 128 bytes. Failing to do either breaks compatibility
   with libsecret clients. See commit c781717 in grimsteel/pass-secret-service.

2. **godbus export pattern**: To export a Go struct as a D-Bus object, use
   `conn.Export(obj, path, iface)`. The exported methods must have the exact signature
   `func(args...) (returns..., *dbus.Error)`. The D-Bus method names are the Go method names
   exactly. Use `introspect.NewIntrospectable` to expose introspection.

3. **D-Bus properties**: Use the `prop` subpackage from `godbus/dbus/v5/prop` for the
   `org.freedesktop.DBus.Properties` interface. Mandatory for libsecret compatibility.

4. **Null prompts**: When an operation completes immediately (no user interaction needed),
   return `dbus.ObjectPath("/")` as the prompt path per spec §6.

5. **Object lifecycle**: When an item/collection is deleted, call `conn.Export(nil, path, iface)`
   to unexport the D-Bus object.

6. **`Replace` flag**: `conn.RequestName(ServiceName, dbus.NameFlagReplaceExisting)` asks the
   bus to evict the current holder. Without this flag, the second `gopass secret-service serve`
   invocation quietly fails to acquire the name.

7. **`_meta` and `_aliases` naming**: These use underscore prefix to avoid collisions with
   user-created secrets. In the `SearchItems` listing loop, always skip paths ending in
   `/_meta` and `/_aliases`.

8. **Build tags**: Every file in `internal/secretservice/` must start with `//go:build linux`.
   The action shims in `internal/action/` use the `_linux.go` / `_other.go` filename convention
   (Go's implicit build tag from filename suffix) which is equivalent, and consistent with
   `notify_dbus.go` and `unclip_linux.go` in the existing codebase.

9. **gopass API vs CLI**: Use `pkg/gopass/api.New(ctx)` (the public Go API), NOT `exec.Command("gopass", ...)`.
   The API is already used by gopass integrations and is stable enough for this purpose.

10. **Error wrapping**: D-Bus methods must return `*dbus.Error`, not `error`. Map internal errors
    to the appropriate `org.freedesktop.Secret.Error.*` D-Bus errors defined in `errors.go`.

---

## References

- [Secret Service Spec (latest)](https://specifications.freedesktop.org/secret-service/latest/)
- [godbus/dbus/v5 docs](https://pkg.go.dev/github.com/godbus/dbus/v5)
- [nikicat/gopass-secret-service](https://github.com/nikicat/gopass-secret-service) — Go reference impl (MIT)
- [grimsteel/pass-secret-service](https://github.com/grimsteel/pass-secret-service) — Rust reference impl (GPL-3.0, for reference only, not to be copied)
- [RFC 3526 — MODP DH groups](https://www.rfc-editor.org/rfc/rfc3526) — 1024-bit group 2
- [RFC 5869 — HKDF](https://www.rfc-editor.org/rfc/rfc5869) — key derivation for `dh-ietf1024-…`
- [golang.org/x/crypto/hkdf](https://pkg.go.dev/golang.org/x/crypto/hkdf) — HKDF implementation
- [keyctl(2)](https://man7.org/linux/man-pages/man2/keyctl.2.html) — kernel keyring for the `session` collection
- [pkg/gopass/api/api.go](../../pkg/gopass/api/api.go) — gopass public API
- [internal/notify/notify_dbus.go](../../internal/notify/notify_dbus.go) — existing godbus usage pattern
- [pkg/clipboard/unclip_linux.go](../../pkg/clipboard/unclip_linux.go) — existing Linux-only D-Bus pattern
