# A-10: Code Quality Findings

**Status:** open — improvements tracked here for future work  
**Source:** SECURITY_AUDIT_REPORT.md § Q-1 through Q-4

---

## Q-1: Path Traversal Protection Inconsistency

**Problem:** Path sanitisation is performed differently at each entry point:

| Location | Approach |
|----------|----------|
| `leaf.Store.Set()` | Rejects names containing `//` |
| `fs` storage layer | `filepath.Clean` + bounds check (added in C-1 fix) |
| Binary action | `isInStore()` with `filepath.Abs()` |
| `create` wizard | `CleanFilename()` |
| `gopass show` | No explicit check |

The inconsistency creates a maintenance risk: a new entry point that forgets
to apply the correct check would silently reintroduce a path traversal
vulnerability.

**Recommendation:** Introduce a single exported function:

```go
// internal/backend/storage/fs/validate.go
func ValidateSecretName(storePath, name string) error {
    resolved := filepath.Join(storePath, filepath.Clean(name))
    if !strings.HasPrefix(resolved, storePath+string(filepath.Separator)) {
        return fmt.Errorf("path traversal detected: %q escapes store root", name)
    }
    return nil
}
```

Replace all divergent ad-hoc checks with calls to this function. This is a
medium-effort refactor but significantly reduces the attack surface for future
regressions.

---

## Q-2: Platform-Specific Behaviour Differences

**Problem:** Several security-relevant behaviours differ across platforms in
ways that are not always intentional or documented:

| Behaviour | Linux/macOS | Windows |
|-----------|-------------|---------|
| Path traversal (`../`) | Rejected by C-1 fix | Rejected by C-1 fix (test expectation updated) |
| Tempfile ramdisk | `/dev/shm` or macOS ramdisk | Falls back to OS temp dir — not a ramdisk |
| Clipboard clear | Kills predecessor `unclip` processes | No predecessor killing |
| Editor parsing | `shellquote.Split` + `LookPath` | Direct argument passing, no `LookPath` |

The editor parsing inconsistency on Windows is notable: the H-3 fix added
`exec.LookPath` validation on non-Windows platforms only (consistent with the
existing `runtime.GOOS != "windows"` guard), so Windows users do not get the
benefit of early binary validation.

**Recommendation:**
- Extend the `LookPath` check to Windows (the `exec.LookPath` function works
  on Windows; the guard is in the `shellquote.Split` branch which is
  Windows-only excluded).
- Document the tempfile ramdisk limitation on Windows so users understand that
  plaintext may temporarily appear in the OS temp directory on that platform.

---

## Q-3: Error Handling Inconsistencies

**Problem:** Several recurring patterns make errors invisible or misleading:

1. **`_ = os.Setenv(...)`** — environment variable set failures are silently
   ignored. A failed `Setenv` could mean secrets are not exported to the
   subprocess in `gopass env`, which would produce a confusing user experience
   without any diagnostic.

2. **`fsck.go` permission repair** — permission change failures are logged as
   warnings but not returned as errors, so `gopass fsck --fix` exits 0 even if
   it could not actually fix the issues it found.

3. **`DetectCrypto` returning `nil, nil`** — documented in L-5; should return
   an explicit error.

4. **Template functions returning `err.Error()` as a string** — partially
   addressed in H-1 (error strings are now generic), but the underlying pattern
   (swallowing the error and embedding it in the output) should be replaced with
   proper error propagation.

**Recommendation:** Audit all `_ = ...` error discards and replace them with
at minimum a `debug.Log` call. Review `fsck.go` to ensure errors from repair
operations are aggregated and returned.

---

## Q-4: Dead Code

**Problem:** Several code paths are permanently unreachable but remain in the
repository, increasing maintenance burden and creating confusion about intent:

1. **Hook system** (`internal/hook/hook.go#L47–L91`) — unreachable below the
   hardcoded `if true { return nil }` guard. Documented in A-7; should be kept
   until GH-2546 is resolved, but the intent must be clearly stated.

2. **GitHub recipient support in age backend** — the `github:` key prefix was
   removed as a supported feature, but the age backend still contains handling
   code that logs a warning when it is encountered. If the feature is truly
   gone, this code should be removed.

3. **Passage identity loading** — deprecated format support for the `passage`
   fork's identity file layout. If this format is no longer supported, the
   parsing code should be removed to reduce the attack surface of the identity
   loading path.

**Recommendation:** Each piece of dead code should be explicitly categorised:
- **Keep with comment** (hook system — pending GH-2546)
- **Remove** (github: prefix warning, passage identity loading) if those
  features are confirmed gone

A `go vet` + `staticcheck` run in CI would catch some categories of dead code
automatically and should be part of the standard `make codequality` target.
