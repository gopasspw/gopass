# gopass Security & Code Quality Audit Report

**Date:** April 4, 2026
**Scope:** Full source code analysis of the gopass project (commit on `master`)
**Methodology:** Manual static analysis of all security-critical code paths

---

## Table of Contents

1. [Executive Summary](#executive-summary)
2. [Critical Findings](#critical-findings)
3. [High Severity Findings](#high-severity-findings)
4. [Medium Severity Findings](#medium-severity-findings)
5. [Low Severity / Informational](#low-severity--informational)
6. [Code Quality & Inconsistencies](#code-quality--inconsistencies)
7. [Positive Security Observations](#positive-security-observations)
8. [Recommendations Summary](#recommendations-summary)

---

## Executive Summary

gopass is a well-engineered password manager with strong fundamentals: PIE binaries, no CGo dependencies, proper use of `exec.Command` argument arrays (avoiding shell injection), signed updates with TLS 1.3, and a clean pluggable backend architecture. The project demonstrates security-first thinking in many areas.

However, the audit identified several issues ranging from a path traversal gap on Windows to password leakage through environment variables, template-based secret extraction, and code bugs. The most critical actionable items are the missing path traversal validation in the storage layer, the `env` command leaking secrets via process environment, and two counters in the `grep` command that are never incremented.

---

## Critical Findings

### C-1: Path Traversal — Missing Bounds Check in Storage Layer

Status: fixed

**Location:** [internal/backend/storage/fs/store.go](internal/backend/storage/fs/store.go#L43)

All storage operations use `filepath.Join(s.path, filepath.Clean(name))` but **never validate that the resulting path remains under `s.path`**. While `filepath.Clean` removes redundant separators, it does *not* strip leading `../` sequences:

```go
filepath.Clean("../../../etc/passwd")  // returns "../../../etc/passwd"
filepath.Join("/store/root", "../../../etc/passwd")  // returns "/etc/passwd"
```

The leaf store's `Set()` only rejects names containing `//` ([internal/store/leaf/write.go#L19](internal/store/leaf/write.go#L19)), which does not catch `..` sequences.

**Evidence from tests:** The existing test in [internal/store/leaf/write_test.go#L25-L29](internal/store/leaf/write_test.go#L25-L29) shows this is a known issue:
```go
if runtime.GOOS != "windows" {
    require.Error(t, s.Set(ctx, "../../../../../etc/passwd", sec))
} else {
    require.NoError(t, s.Set(ctx, "../../../../../etc/passwd", sec))
}
```
On non-Windows, the write fails only because filesystem permissions prevent writing to `/etc/passwd.age` — not because the path is explicitly rejected. On Windows, the test **expects success**, confirming path traversal is possible.

**Impact:** An attacker who can control secret names (e.g., via a malicious git repository that is cloned/synced) could read or write files outside the password store.

**Recommendation:** Add an explicit check after path construction:
```go
resolved := filepath.Join(s.path, filepath.Clean(name))
if !strings.HasPrefix(resolved, s.path+string(filepath.Separator)) {
    return fmt.Errorf("path traversal detected: %q escapes store root", name)
}
```

---

### C-2: `env` Command Leaks Passwords via Process Environment

Status: fixed

**Location:** [internal/action/env.go#L68-L76](internal/action/env.go#L68-L76)

The `gopass env` command injects decrypted passwords into the subprocess environment:

```go
env = append(env, fmt.Sprintf("%s=%s", envKey, sec.Password()))
cmd.Env = append(os.Environ(), env...)
```

Process environment variables are observable by:
- Any process that can read `/proc/<pid>/environ` on Linux
- Tools like `ps eww` on macOS/Linux
- Child process introspection on all platforms

While this is the documented behavior, it is inherently unsafe for a password manager and contradicts security best practices.

**Recommendation:** Document the risk prominently. Consider adding a `--stdin` mode that pipes secrets via stdin instead, or at minimum clear the environment variable before the subprocess exits. Other tools (e.g., `aws-vault`) use short-lived temporary credentials for this pattern, which is not applicable to static passwords.

---

## High Severity Findings

### H-1: Template Engine Allows Unrestricted Secret Access

Status: fixed

**Location:** [internal/tpl/funcs.go#L195-L270](internal/tpl/funcs.go#L195-L270), [internal/action/process.go](internal/action/process.go)

The template engine exposes `get`, `getpw`, `getval`, and `getvals` functions that can read **any** secret in the entire store without restriction. The `gopass process` command reads an arbitrary file and executes it as a template:

```go
buf, err := os.ReadFile(file)               // User-supplied file
obuf, err := tpl.Execute(ctx, string(buf), file, nil, s.Store)  // Full store access
```

An attacker who tricks a user into running `gopass process malicious.txt` can extract every secret:
```
{{getpw "admin/root-password"}}
{{get "production/database-credentials"}}
```

Additionally, error messages from failed `get()` calls are returned as template output ([funcs.go#L207](internal/tpl/funcs.go#L207)):
```go
if err != nil {
    return err.Error(), nil  // Leaks decryption errors, paths, backend details
}
```

**Impact:** Complete secret store extraction if a user processes an untrusted template file. Information disclosure via error messages.

**Recommendation:**
- Return generic error messages from template functions instead of `err.Error()`
- Document the `process` command's security implications prominently
- Consider adding a `--allow-paths` flag to restrict which secrets templates can access

### H-2: Symlink Following in Store Walk Can Escape Store Boundary

Status: fixed

**Location:** [internal/backend/storage/fs/walk.go#L27-L40](internal/backend/storage/fs/walk.go#L27-L40)

The `walkSymlinks()` function follows directory symlinks without validating that the target remains within the store:

```go
destPath, err := filepath.EvalSymlinks(path)  // Follows symlinks to any location
if destInfo.IsDir() {
    return walk(destPath, path, walkFn)  // Recursively walks outside the store
}
```

There is no depth limit or cycle detection beyond what `filepath.EvalSymlinks` provides. A symlink pointing to `/` or `/home` would cause the store list operation to enumerate the entire filesystem.

**Impact:** Information disclosure of filesystem structure. Potential denial of service via symlink loops or large directory trees.

**Recommendation:** After resolving a symlink, validate the target is under the store root. Add a maximum depth counter to prevent symlink loops.

### H-3: Editor Command Parsing via Shell Quoting

Status: fixed

**Location:** [internal/editor/editor.go#L62-L69](internal/editor/editor.go#L62-L69)

The editor command (from `$EDITOR` env var or `edit.editor` config) is parsed using `shellquote.Split()`:

```go
cmdArgs, err := shellquote.Split(editor)
editor = cmdArgs[0]
args = append(args, cmdArgs[1:]...)
```

While `shellquote.Split` is safer than `sh -c`, a malicious `EDITOR` value can still execute unintended programs. The `$EDITOR` environment variable is user-controlled, and the `edit.editor` config key could be modified in a shared/synced configuration.

**Impact:** Arbitrary command execution if `EDITOR` or config is attacker-controlled.

**Recommendation:** After parsing, validate that the resolved editor binary exists at an expected path. Consider a whitelist of known editors.

---

## Medium Severity Findings

### M-1: `grep` Command — Match and Error Counters Never Incremented

**Location:** [internal/action/grep.go#L42-L57](internal/action/grep.go#L42-L57)

The `matches` and `errors` counters are declared but never incremented:

```go
var matches int
var errors int
for _, v := range haystack {
    sec, err := s.Store.Get(ctx, v)
    if err != nil {
        out.Errorf(ctx, "failed to decrypt %s: %v", v, err)
        // MISSING: errors++
        continue
    }
    if matchFn(string(sec.Bytes())) {
        out.Printf(ctx, "%s matches", color.BlueString(v))
        // MISSING: matches++
    }
}
out.Printf(ctx, "\nScanned %d secrets. %d matches, %d errors", len(haystack), matches, errors)
```

The summary line always reports "0 matches, 0 errors" regardless of actual results.

**Impact:** Misleading output. Users cannot verify grep reliability.

### M-2: `text/template` Used Instead of `html/template`

**Location:** [internal/tpl/template.go#L10](internal/tpl/template.go#L10), [internal/tpl/funcs.go#L9](internal/tpl/funcs.go#L9)

Go's `text/template` package allows arbitrary method calls on any value passed to the template. While the current `payload` struct only contains string fields (limiting the attack surface), using `text/template` in a security-sensitive context is a known anti-pattern. If the payload struct ever gains a method or field that returns a more complex type, template injection could escalate.

**Recommendation:** Evaluate whether `html/template` (which auto-escapes) could be used, or add a strict allowlist of template actions.

### M-3: Age SSH Key Cache Has No Thread Safety

Status: fixed

**Location:** `internal/backend/crypto/age/ssh.go` — global `sshCache` variable

The `sshCache` global is read and written without synchronization. If multiple goroutines call `getSSHIdentities()` concurrently, a data race can occur.

**Impact:** Minor — worst case is loading SSH keys twice or returning a partial cache. However, data races are undefined behavior in Go.

**Recommendation:** Use `sync.Once` for cache initialization.

### M-4: No Minimum Password Length Enforcement

**Location:** [internal/config/config.go](internal/config/config.go)

The `GOPASS_PW_DEFAULT_LENGTH` environment variable and `generate.length` config key accept any value >= 1. There is no minimum enforced (e.g., 8 or 12 characters).

**Impact:** Users could accidentally generate very weak passwords.

**Recommendation:** Enforce a minimum of 8 characters or display a warning when generating passwords shorter than 12 characters.

### M-5: Hook System Disabled But Code Present (CVE-2023-24055)

**Location:** [internal/hook/hook.go#L43-L46](internal/hook/hook.go#L43-L46)

The hook system is correctly disabled with a hardcoded `return nil`:
```go
if true {
    // TODO(GH-2546) disabled until further discussion, cf. CVE-2023-24055
    return nil
}
```

However, the vulnerable code remains below the early return. If this is ever re-enabled without fixing the `shellquote.Split` parsing, it would allow command injection via config values.

**Recommendation:** Either remove the dead code entirely or implement safe hook execution (direct `exec.Command` with argument arrays, not shell parsing) before re-enabling.

### M-6: Shred Operation Not Effective on Modern Storage

**Location:** [pkg/fsutil/fsutil.go#L152-L216](pkg/fsutil/fsutil.go#L152-L216)

The `Shred()` function overwrites files with random data before deletion. On modern SSDs with wear leveling, journaling filesystems (ext4, APFS, NTFS), and copy-on-write filesystems (ZFS, Btrfs), this provides no security guarantee — the original data may persist in reallocated blocks or journal entries.

**Impact:** False sense of security. Users may believe binary files are securely deleted when they are not.

**Recommendation:** Document this limitation. On macOS/Linux, consider using platform-specific APIs or the ramdisk-based tempfile system for sensitive temporary data (which is already done for the editor).

---

## Low Severity / Informational

### L-1: GPG Ciphertext Logged to Debug Output

**Location:** `internal/backend/crypto/gpg/cli/encrypt.go`

Encrypted ciphertext is hex-dumped to debug output via `io.MultiWriter`:
```go
hexLogger := hex.Dumper(debug.LogWriter)
cmd.Stdout = io.MultiWriter(buf, hexLogger)
```

While ciphertext is not plaintext, logging it could assist offline attacks.

**Recommendation:** Only log ciphertext at verbose debug level (V(2) or higher).

### L-2: `GOPASS_GPG_BINARY` Allows Binary Override Without Validation

The `GOPASS_GPG_BINARY` environment variable allows overriding the GPG binary path. An attacker with environment control could point this to a malicious binary.

**Impact:** Low — if an attacker controls the environment, they likely have more direct attack vectors.

### L-3: Updater Relies on CA-Signed TLS Without Certificate Pinning

**Location:** [internal/updater/download.go](internal/updater/download.go)

While TLS 1.3 is enforced, there is no certificate pinning for `github.com`. A compromised CA could issue a fraudulent certificate.

**Impact:** Low — the GPG signature verification of the download provides defense-in-depth. An attacker would need both a forged certificate AND the project's GPG signing key.

### L-4: `text/template` Panic on Nil Access

If a template references `{{.Content}}` and `content` is `nil` (which happens in the `process` command where `nil` is passed), the template engine handles it gracefully. However, accessing methods on nil interfaces in custom template functions could panic.

### L-5: DetectCrypto Returns nil, nil

**Location:** [internal/backend/registry.go](internal/backend/registry.go)

`DetectCrypto()` can return `nil, nil` when no backend is detected, with a TODO comment. Callers must handle this case or face nil pointer dereferences.

### L-6: HTTP Proxy Honored for Update Downloads

**Location:** [internal/updater/download.go](internal/updater/download.go)

```go
Proxy: http.ProxyFromEnvironment,
```

The updater respects `HTTP_PROXY`/`HTTPS_PROXY` environment variables. A malicious proxy could intercept the download, but this is mitigated by GPG signature verification.

---

## Code Quality & Inconsistencies

### Q-1: Path Traversal Protection Inconsistency

The leaf store's `Set()` checks for `//` but not `..`. The fs storage layer uses `filepath.Clean` but not bounds checking. The binary action uses `isInStore()` with proper `filepath.Abs()` validation. The `create` wizard uses `CleanFilename()`. There is no single, consistent path validation function used across all entry points.

**Recommendation:** Create a single `ValidateSecretName(storePath, name string) error` function that validates all secret names at the boundary and use it consistently.

### Q-2: Platform-Specific Behavior Differences

| Behavior | Linux/macOS | Windows |
|----------|-------------|---------|
| Path traversal (`../`) | Fails (filesystem permissions) | Succeeds (test expects no error) |
| Tempfile ramdisk | `/dev/shm` or macOS ramdisk | No-op (falls back to OS temp) |
| Clipboard clear | Kills predecessor `unclip` processes | No predecessor killing |
| Editor parsing | `shellquote.Split` | Direct argument passing |

The Windows path traversal difference is the most concerning inconsistency.

### Q-3: Error Handling Inconsistencies

- `fsck.go`: Permission repair failures are only warned, not returned as errors
- Template functions: Return `err.Error()` as a string value instead of propagating the error
- `DetectCrypto`: Returns `nil, nil` instead of an explicit error
- Multiple places use `_ = os.Setenv(...)` ignoring the error

### Q-4: Dead Code

- The hook system below the `return nil` guard ([internal/hook/hook.go#L47-L91](internal/hook/hook.go#L47-L91))
- GitHub recipient support warning in age backend (feature removed but code handles `github:` prefix)
- Passage identity loading (deprecated format support)

### Q-5: Missing `errors++` and `matches++` in Grep

As detailed in M-1, both counters in `grep.go` are declared but never incremented — a simple but impactful oversight.

---

## Positive Security Observations

The following areas demonstrate excellent security engineering:

1. **Binary hardening:** PIE binaries, stripped symbols, trimpath, `CGO_ENABLED=0`, `netgo` tag
2. **No shell injection:** Consistent use of `exec.Command` with string slice arguments throughout all backends
3. **Update verification:** GPG-signed checksums with hardcoded public key + TLS 1.3 minimum
4. **Age agent socket security:** Checks socket permissions (`0o600`) and ownership (UID) before connecting
5. **Clipboard auto-clear:** Argon2id checksum verification before clearing, detached unclip process
6. **Debug log protection:** `SafeStr()` interface masks secrets in logs; `GOPASS_DEBUG_LOG_SECRETS` must be explicitly enabled
7. **Tempfile security:** Ramdisk on macOS, `/dev/shm` on Linux, `0o600` permissions
8. **Recipient validation:** GPG backend validates key usability (expiration, trust, encryption capability) before encrypting
9. **OpenBSD pledge:** `protect.Pledge("stdio rpath wpath cpath tty proc exec fattr")` restricts syscalls
10. **Proper secret masking:** The `out.Secret` type implements `SafeStr()` returning `"(elided)"`

---

## Recommendations Summary

| Priority | ID | Issue | Effort |
|----------|----|-------|--------|
| **Critical** | C-1 | Add path bounds check in storage layer | Small |
| **Critical** | C-2 | Document/mitigate `env` command password leakage | Medium |
| **High** | H-1 | Restrict template secret access; fix error leakage | Medium |
| **High** | H-2 | Add bounds check to symlink walk | Small |
| **High** | H-3 | Validate editor binary path | Small |
| **Medium** | M-1 | Fix `grep` match/error counters | Trivial |
| **Medium** | M-3 | Add `sync.Once` to SSH cache | Trivial |
| **Medium** | M-4 | Enforce minimum password length | Small |
| **Medium** | M-5 | Remove or fix dead hook code | Small |
| **Medium** | M-6 | Document shred limitations on modern storage | Trivial |
| **Low** | Q-1 | Unify path validation into single function | Medium |
| **Low** | Q-2 | Fix Windows path traversal in tests | Small |

---

*This report was produced through manual static analysis of the gopass source code. No dynamic testing, fuzzing, or dependency vulnerability scanning (beyond code review) was performed. The findings should be validated and prioritized by the maintainers.*
