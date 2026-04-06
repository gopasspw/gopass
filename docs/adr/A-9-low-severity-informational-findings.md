# A-9: Low Severity and Informational Findings

**Status:** accepted — risks documented; mitigations noted where applicable  
**Source:** SECURITY_AUDIT_REPORT.md § L-1 through L-6

---

## L-1: GPG Ciphertext Logged to Debug Output

**Location:** `internal/backend/crypto/gpg/cli/encrypt.go`

Encrypted ciphertext is hex-dumped to the debug log via `io.MultiWriter`:

```go
hexLogger := hex.Dumper(debug.LogWriter)
cmd.Stdout = io.MultiWriter(buf, hexLogger)
```

While ciphertext is not plaintext, persisting it to a debug log could assist
offline brute-force or analysis of the encryption scheme.

**Mitigation:** The debug log is only written when `GOPASS_DEBUG_LOG` is
explicitly set. Ciphertext should be logged at a higher verbosity level
(e.g. `debug.V(2)`) rather than unconditionally, so that standard debug output
does not include bulk ciphertext.

---

## L-2: `GOPASS_GPG_BINARY` Allows Binary Override Without Validation

The `GOPASS_GPG_BINARY` environment variable allows the user to override the
path to the GPG binary. An attacker with control over the process environment
could point it to a malicious program.

**Assessment:** If an attacker controls the environment of the gopass process,
they already have equivalently powerful attack vectors (e.g. replacing
`GOPATH`, `PATH`, or setting `LD_PRELOAD`). The incremental risk from this
specific variable is low.

**Mitigation:** No change required. The variable is documented as a
developer/debugging aid and requires environment access that implies broader
privilege.

---

## L-3: Updater Relies on CA-Signed TLS Without Certificate Pinning

**Location:** `internal/updater/download.go`

The built-in updater connects to `github.com` over TLS 1.3 (min version
enforced) but does not pin the server certificate or its public key. A CA
that issues a fraudulent certificate for `github.com` could perform a
man-in-the-middle attack against the update download.

**Assessment:** The risk is mitigated by defence-in-depth: the downloaded
binary's SHA-256 checksum is verified against a checksum file, and that
checksum file's GPG signature is verified against a **hardcoded** project
public key embedded in the binary. An attacker would therefore need to
compromise both a trusted CA **and** the project's GPG signing key, which
is an implausible combination.

Certificate pinning would add implementation complexity and create an
operational burden (the pin must be updated with every certificate renewal)
without meaningfully reducing the realistic threat surface.

**Mitigation:** No change required. The current GPG-signed checksum
verification provides sufficient defence-in-depth.

---

## L-4: `text/template` Panic on Nil Interface in Custom Functions

If a custom template function receives a nil interface value and attempts a
method call on it, the Go runtime will panic. The `process` command currently
passes `nil` as the secret payload; the template engine handles `{{.Field}}`
access on nil gracefully via reflection, but custom Go functions called from
within templates are not similarly protected.

**Assessment:** Exploitable only if a user deliberately authors a template that
triggers the nil path — this is a denial-of-service against the user's own
session rather than a privilege-escalation vector.

**Mitigation:** Custom template functions should guard against nil receivers
using explicit nil checks. New template functions added in the future should be
reviewed for nil-safety.

---

## L-5: `DetectCrypto` Returns `nil, nil`

**Location:** `internal/backend/registry.go`

`DetectCrypto()` can return `(nil, nil)` when no matching backend is detected.
The function has a `// TODO` comment acknowledging this. Callers that do not
check for a nil crypto backend before using it will panic with a nil-pointer
dereference.

**Assessment:** All current call sites do check the error + nil before use.
The risk is that a future caller forgets.

**Mitigation:** The function should be updated to return an explicit
`fmt.Errorf("no crypto backend detected for %s", path)` error instead of
`nil, nil`. This is a small refactor that eliminates the ambiguous return and
allows callers to use standard Go error-checking patterns without needing to
special-case nil.

---

## L-6: HTTP Proxy Honoured for Update Downloads

**Location:** `internal/updater/download.go`

```go
Proxy: http.ProxyFromEnvironment,
```

The updater respects `HTTP_PROXY` / `HTTPS_PROXY` / `NO_PROXY`. A malicious
proxy could intercept or block the update download.

**Assessment:** The GPG-signed checksum verification (see L-3) means a proxy
cannot substitute a malicious binary without also forging the project GPG
signature. Proxy support is a legitimate operational requirement in corporate
environments where direct internet access may be disallowed.

**Mitigation:** No change required. Proxy support is correct and expected
behaviour in enterprise environments; GPG signature verification prevents
payload substitution.
