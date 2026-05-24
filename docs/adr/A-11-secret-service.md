# A-11: Integrate `org.freedesktop.secrets` D-Bus Service

**Status:** proposed  
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
| [nikicat/gopass-secret-service](https://github.com/nikicat/gopass-secret-service) | Go (90 %) | MIT | Standalone daemon that invokes `gopass` CLI. **Most relevant.** |
| [grimsteel/pass-secret-service](https://github.com/grimsteel/pass-secret-service) | Rust | GPL-3.0 | Standalone daemon for `pass`. Pure-Rust D-Bus. |

`nikicat/gopass-secret-service` implements the full spec in Go and re-uses `gopass` via its CLI.
The key difference for this ADR is that the integrated version will use the **gopass Go API**
(`github.com/gopasspw/gopass/pkg/gopass/api`) directly rather than shelling out.

---

## Decision

Implement `gopass secret-service` as:

1. A new **Linux-only** subcommand of the main `gopass` binary.
2. A long-running **daemon** that acquires `org.freedesktop.secrets` on the D-Bus session bus and
   serves the full Secret Service interface.
3. Uses `github.com/godbus/dbus/v5` (already in `go.mod` at v5.1.0, pure Go, BSD-2 licensed).
4. All crypto uses stdlib `crypto/...` packages — no CGo, no new external dependencies.
5. Secrets are stored under a configurable gopass path prefix (default: `secret-service`).

---

## Feasibility Summary

* **Pure-Go, zero-CGo**: `godbus/dbus/v5` is already a direct dependency; all required crypto
  (`crypto/aes`, `crypto/sha256`, `math/big` for DH) is in the stdlib.
* **No new external dependencies**: Only stdlib + existing `godbus` dependency needed.
* **License-compatible**: `godbus/dbus/v5` is BSD-2 (≡ MIT compatible per `.license-lint.yml`).
* **Linux-only build tag**: The entire feature is gated with `//go:build linux` (same pattern as
  `internal/notify/notify_dbus.go` and `pkg/clipboard/unclip_linux.go`).
* **Architectural risk**: gopass is a short-lived CLI tool; this feature requires a persistent
  daemon process. This is handled by a blocking `gopass secret-service serve` subcommand (the user
  manages the lifecycle via systemd or similar).

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
    │   └── i<uuid>.age       # Secret items
    └── work/
        ├── _meta.age
        └── i<uuid>.age
```

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

## Crypto Sessions

The spec defines two `OpenSession` algorithms:

| Algorithm | Description | Implementation |
|-----------|-------------|----------------|
| `plain` | No transport encryption | Return empty bytes; secret value passed as-is |
| `dh-ietf1024-sha256-aes128-cbc-pkcs7` | DH key exchange + AES-128-CBC | stdlib `crypto/aes`, `crypto/sha256`, `math/big` |

The `plain` algorithm is secure because D-Bus session bus traffic is carried over a local UNIX
socket with kernel-enforced access control. Implementing `dh-ietf1024-sha256-aes128-cbc-pkcs7` is
required for compatibility with `libsecret`-based applications.

DH parameters: [RFC 3526](https://www.rfc-editor.org/rfc/rfc3526) 1024-bit MODP group 2.
The server:
1. Generates a DH ephemeral key pair on the MODP-1024 group.
2. Receives the client's public key in `OpenSession`.
3. Computes `shared = clientPub^serverPriv mod p`.
4. Left-pads `shared` to 128 bytes (a known pitfall — see grimsteel commit c781717).
5. Derives AES key: `aes_key = SHA256(shared)[0:16]`.
6. Each secret returned has its own random 16-byte IV prepended to the ciphertext.

---

## Package Structure

All new code lives under `internal/secretservice/` (Linux-only files) and integrates
via a new `secretservice_linux.go` action handler shim.

```
internal/secretservice/
├── doc.go                    # Package doc
├── service.go                # org.freedesktop.Secret.Service implementation
├── collection.go             # org.freedesktop.Secret.Collection implementation
├── item.go                   # org.freedesktop.Secret.Item implementation
├── session.go                # Session lifecycle + crypto dispatch
├── prompt.go                 # Prompt objects (required by spec for async ops)
├── crypto/
│   ├── crypto.go             # Session interface + factory
│   ├── plain.go              # "plain" algorithm
│   └── dh.go                 # "dh-ietf1024-sha256-aes128-cbc-pkcs7"
├── store.go                  # Adapter between Secret Service and gopass API
├── errors.go                 # D-Bus error definitions (org.freedesktop.DBus.Error.*)
├── types.go                  # D-Bus type aliases (Secret struct, path constants)
└── service_test.go           # Unit tests (mock D-Bus + mock gopass API)
```

CLI integration:

```
internal/action/
├── secretservice_linux.go    # SecretService() handler, registers via GetCommands()
└── secretservice_other.go    # Stub for non-Linux platforms that prints "linux only"
```

Systemd / D-Bus activation files (installed by `gopass secret-service install`):

```
contrib/secret-service/
├── org.freedesktop.secrets.service   # D-Bus session activation (ExecStart=gopass secret-service serve)
└── gopass-secret-service.service     # systemd user unit
```

---

## Implementation Phases

This feature is too large for a single prompt. Each phase below is self-contained and can be
implemented and tested independently. Phases must be implemented in order because each phase
depends on the previous.

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
    // 5. aesKey = SHA256(paddedShared)[0:16]
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
// Lists all items in a collection, loads each item's attributes, filters by attrs.
// This is O(n) — acceptable given typical collection sizes.
```

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
// - secret created via D-Bus appears in gopass (gopass show secret-service/default/i<uuid>)
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

1. **DH shared-secret left-padding**: `shared = clientPub^serverPriv mod p` may produce a
   `big.Int` whose byte representation is shorter than 128 bytes. It must be left-padded with
   zero bytes to exactly 128 bytes before SHA256. Failing to do this breaks compatibility with
   libsecret clients. See commit c781717 in grimsteel/pass-secret-service.

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
- [pkg/gopass/api/api.go](../../pkg/gopass/api/api.go) — gopass public API
- [internal/notify/notify_dbus.go](../../internal/notify/notify_dbus.go) — existing godbus usage pattern
- [pkg/clipboard/unclip_linux.go](../../pkg/clipboard/unclip_linux.go) — existing Linux-only D-Bus pattern
