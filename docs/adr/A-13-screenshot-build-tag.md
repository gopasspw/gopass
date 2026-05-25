# A-13: `noscreenshot` Build Tag for OTP Screen-Capture Feature

**Status:** accepted  
**Source:** [GitHub Issue #3415](https://github.com/gopasspw/gopass/issues/3415)

---

## Background

`pkg/otp/screenshot_supported.go` imports `github.com/kbinani/screenshot` to capture display
contents when the user runs `gopass otp --snip`.  This allows gopass to locate an OTP QR code
that is visible on screen and store the decoded `otpauth://` URL directly into a secret.

`screenshot` is a notable dependency for a password manager because it grants the binary the
ability to read the full screen.  Users auditing binary capabilities or operating in
policy-restricted environments (e.g., enterprise security reviews, MDM policies) may need to
understand this surface or opt out of it entirely at compile time.

### Audit findings

* **Which function?**  `pkg/otp.ParseScreen` (implemented in `screenshot_supported.go`) is the
  sole caller of the `kbinani/screenshot` API.
* **Caller?**  `internal/action.(*otpHandler).OTP` in `internal/action/otp.go`, guarded by
  `if snip { … }`.
* **User-visible trigger?**  The `--snip` / `-s` flag on `gopass otp`.  Screen capture
  **cannot** be triggered without the user explicitly passing this flag.
* **Affected packages?**  `pkg/passkey` does **not** use `screenshot`.  The issue description
  was slightly inaccurate on this point.
* **Platform scope?**  Only compiled on `(arm|arm64|amd64|386) && (linux|windows|(cgo &&
  darwin)|freebsd|netbsd)`.  Other platforms already receive a no-op stub.

---

## Decision

Implement the `noscreenshot` build tag (Option A below).  It is low-risk, additive, and
allows enterprise / policy-aware users to produce a binary without the screen-capture surface
while keeping the default experience unchanged.

---

## Options considered

### A — Negative opt-out build tag `noscreenshot` ✅ (chosen)

Add `&& !noscreenshot` to the existing build constraint in `screenshot_supported.go` and
`|| noscreenshot` to the fallback `screenshot_others.go`.

```
go build -tags noscreenshot .
```

**Pros:**
* No change to the default build; existing users and CI pipelines are unaffected.
* Simple: one tag, two build-constraint lines.
* Canonical Go pattern for feature opt-out (mirrors `nomsgpack`, `noasm`, etc.).

**Cons:**
* The tag name must be documented; there is no automated reminder to keep stubs in sync.

### B — Separate module / package `pkg/otp/screenshot`

Move the screenshot logic into a sub-package and make `ParseScreen` a function variable
that callers inject.

**Pros:** clean separation.  
**Cons:** more invasive refactor for a minor gain; deferred.

### C — Runtime config flag `otp.screenshot: false`

A config option to disable the feature at runtime without recompiling.

**Pros:** no special build required.  
**Cons:** the library is still linked; does not address the "linked capability" concern.

---

## Consequences

* `gopass otp --snip` continues to work for all users who build without the tag.
* Builds with `-tags noscreenshot` will return `"not supported on your platform"` for
  `--snip`, consistent with unsupported-platform behaviour.
* `github.com/kbinani/screenshot` is absent from the linked binary when the tag is set.
* `docs/commands/otp.md` documents the tag and the screen-capture scope.
