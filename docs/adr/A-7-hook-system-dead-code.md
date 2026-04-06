# A-7: Hook System Dead Code and CVE-2023-24055

**Status:** deferred — hooks remain disabled; safe re-enablement path documented  
**Source:** SECURITY_AUDIT_REPORT.md § M-5

---

## Background

gopass ships a hook system in `internal/hook/hook.go` that was designed to
allow users to run custom commands at lifecycle events (e.g. pre-commit,
post-decrypt). The system was **disabled** by inserting a hardcoded early
return:

```go
if true {
    // TODO(GH-2546) disabled until further discussion, cf. CVE-2023-24055
    return nil
}
```

The code below this guard remains in the repository. It parses hook command
strings using `shellquote.Split()` and then executes the result, which is
the same pattern that was the root cause of CVE-2023-24055 in the KeePass
ecosystem (untrusted config values leading to arbitrary command execution via
shell-style parsing).

---

## Why It Was Disabled

If an attacker can write to the gopass configuration file (e.g. via a
malicious git repository that includes a synced config) they could set a hook
to any command value and have gopass execute it on the next relevant lifecycle
event. The `shellquote.Split` approach would allow argument injection even
without shell metacharacters.

---

## Decision

The hooks remain disabled. The dead code **should not** be removed yet because
the feature is under active discussion (GH-2546). Removing it prematurely
would force a larger re-implementation effort when the discussion concludes.

The dead code is harmless as long as the early-return guard stays in place.

---

## Requirements for Safe Re-enablement

Before hooks can be re-enabled the implementation must satisfy all of the
following:

1. **No shell parsing of hook values.** Hook commands must be stored as
   structured config (e.g. an array of strings for binary + arguments) rather
   than a single shell-quoted string. `shellquote.Split` must not be used.

2. **Binary existence check.** The resolved hook binary must pass an
   `exec.LookPath` check before execution (consistent with the fix applied
   in H-3 for the editor command).

3. **Explicit user consent.** Hooks should require explicit opt-in through a
   first-class config key (e.g. `hooks.enabled = true`) that is **not** synced
   via git-managed config by default, so that a shared/cloned repository
   cannot silently activate hooks on a new machine.

4. **Audit trail.** Each hook execution should be logged at info level so users
   can observe what commands are being run on their behalf.

---

## Implementation Sketch

```go
// safe hook execution — no shell parsing
hookBin, err := exec.LookPath(hookConfig.Command[0])
if err != nil {
    return fmt.Errorf("hook binary %q not found: %w", hookConfig.Command[0], err)
}
cmd := exec.CommandContext(ctx, hookBin, hookConfig.Command[1:]...)
```

When GH-2546 reaches a resolution, this ADR should be updated with the chosen
design and then closed.
