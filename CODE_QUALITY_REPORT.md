# gopass Code Quality, UX & Implementation Audit Report

**Date:** April 4, 2026  
**Scope:** Code style, abstractions, latent bugs, documentation–implementation mismatches, CLI UX  
**Companion:** See `SECURITY_AUDIT_REPORT.md` for the security-focused findings

---

## Table of Contents

1. [Confirmed Bugs](#confirmed-bugs)
2. [Documentation vs Implementation Mismatches](#documentation-vs-implementation-mismatches)
3. [CLI / UX Issues](#cli--ux-issues)
4. [Architecture & Abstraction Concerns](#architecture--abstraction-concerns)
5. [Code Style Inconsistencies](#code-style-inconsistencies)
6. [Improvement Suggestions](#improvement-suggestions)

---

## Confirmed Bugs

All bugs below were verified by reading the source directly.

### B-1: `grep` Command — Match and Error Counters Never Incremented

**Location:** [internal/action/grep.go#L42-L57](internal/action/grep.go#L42-L57)

```go
var matches int
var errors int
for _, v := range haystack {
    sec, err := s.Store.Get(ctx, v)
    if err != nil {
        out.Errorf(ctx, "failed to decrypt %s: %v", v, err)
        // BUG: missing errors++
        continue
    }
    if matchFn(string(sec.Bytes())) {
        out.Printf(ctx, "%s matches", color.BlueString(v))
        // BUG: missing matches++
    }
}
out.Printf(ctx, "\nScanned %d secrets. %d matches, %d errors", len(haystack), matches, errors)
// Always prints "0 matches, 0 errors"
```

Both `matches` and `errors` are declared but never written. The summary line at the end is always wrong.

### B-2: `autoSync` — Timestamp Updated Only on Failure

**Location:** [internal/action/sync.go#L87-L93](internal/action/sync.go#L87-L93)

```go
if time.Since(ls) > syncInterval {
    err := s.sync(ctx, "", true)
    if err != nil {
        autosyncLastRun = time.Now()  // only set on ERROR
    }
    return err
}
```

The `autosyncLastRun` variable is set inside the `if err != nil` block, meaning it is only updated when sync **fails**. On success, the timestamp is never advanced. Combined with the debounce in `sync()` (which checks `time.Since(autosyncLastRun) < 10*time.Second`), this means:

- A successful sync never updates `autosyncLastRun`
- The `sync()` debounce guard at line 101 (`< 10 seconds`) fires correctly only after failures
- The `Reminder.LastSeen("autosync")` on line 71 acts as the real gate for interval-based scheduling

The logic still works because the reminder-based scheduling dominates, but the `autosyncLastRun` debounce is effectivly dead code for the success path. The intent was almost certainly to always update the timestamp.

### B-3: Create Wizard — Password Max Length Check Uses Wrong Variable

**Location:** [internal/create/wizard.go#L294](internal/create/wizard.go#L294)

```go
// Line 230 — CORRECT (for non-password fields):
if v.Max > 0 && len(sv) > v.Max {

// Line 294 — BUG (for password fields):
if v.Max > 0 && len(password) > v.Min {  // should be v.Max
    return fmt.Errorf("%s is too long (at most %d)", v.Name, v.Max)
}
```

The password maximum length check compares against `v.Min` instead of `v.Max`. Any password longer than `v.Min` is rejected as "too long", while passwords exceeding `v.Max` are accepted.

### B-4: `RemoveMount` — Identical Condition Checked Twice

**Location:** [internal/store/root/mount.go#L105-L114](internal/store/root/mount.go#L105-L114)

```go
func (r *Store) RemoveMount(ctx context.Context, alias string) error {
    if _, found := r.mounts[alias]; !found {
        out.Warningf(ctx, "%s is not mounted", alias)
    }
    if _, found := r.mounts[alias]; !found {       // duplicate check
        out.Warningf(ctx, "%s is not initialized", alias)  // unreachable different message
    }
    delete(r.mounts, alias)  // proceeds regardless
```

The same `!found` condition is checked twice with different messages. The state of `r.mounts[alias]` cannot change between the two checks. Additionally, if the mount doesn't exist, the function prints warnings but continues to `delete()` the non-existent key and modify config — silently succeeding when it should probably return an error.

### B-5: `convert` Command — Error Message Prints Wrong Variable

**Location:** [internal/action/convert.go#L41-L43](internal/action/convert.go#L41-L43)

```go
storage, err = backend.StorageRegistry.Backend(sv)
if err != nil {
    return exit.Error(exit.Usage, err,
        "unknown destination storage backend %q: %s", storage, err)
        //                                           ^^^^^^^ should be sv
}
```

When `Backend()` fails, `storage` holds the zero value of the backend type, not the user-supplied string `sv`. The error message shows a meaningless zero value instead of what the user actually typed. The crypto equivalent on line 54 correctly uses `sv`.

### B-6: `convert` (leaf) — "stroage" Typo in Error Message

**Location:** [internal/store/leaf/convert.go#L49](internal/store/leaf/convert.go#L49)

```go
return fmt.Errorf("failed to initialize new stroage backend %s: %w", storageBe.String(), err)
//                                             ^^^^^^^ "stroage" → "storage"
```

### B-7: `audit` Command — No Output Without `--full` or `--summary`

**Location:** [internal/action/audit.go#L88-L91](internal/action/audit.go#L88-L91)

```go
if !c.Bool("full") && !c.Bool("summary") {
    out.Warning(ctx, "No output format specified. Use `--full` or `--summary` to specify.")
}
```

When neither `--full` nor `--summary` is passed, the command prints a warning and returns `nil` (success) — the audit runs but the results are silently discarded. It should either default to `--summary` or return an error.

### B-8: Queue Package — Self-Documented as "Likely Broken"

**Location:** [internal/queue/background.go#L1-L7](internal/queue/background.go#L1-L7)

```go
// Package queue implements an experimental background queue for cleanup jobs.
// Beware: It's likely broken.
// We can easily close a channel which might later be written to.
```

The `Idle()` method uses `len(q.work) < 1` without synchronization (data race) and spawns a goroutine that never terminates (goroutine leak). The `Close()` method can panic if tasks are still being enqueued via `Add()` after the channel is closed.

---

## Documentation vs Implementation Mismatches

### D-1: `pwgen.xkcd-len` Type Wrong in Docs

**Location:** [docs/config.md#L138](docs/config.md#L138)

Documentation says the type is `bool`. Code uses `config.Int(ctx, "pwgen.xkcd-len")` in three places ([internal/action/generate.go#L328](internal/action/generate.go#L328), [internal/action/pwgen/pwgen.go#L60](internal/action/pwgen/pwgen.go#L60), [internal/create/wizard.go#L423](internal/create/wizard.go#L423)). Should be `int`.

### D-2: `show` Command — Undocumented Flags

[docs/commands/show.md](docs/commands/show.md) does not mention:
- `--safe` / `-s` — force safecontent protection (opposite of `--unsafe`)
- `--qrbody` — show body as QR code instead of password
- `--nosync` — disable auto-sync for this invocation
- `--alsoclip` / `-C` — copy password AND show everything

### D-3: `show` Command — Newline Behavior Unspecified

[docs/commands/show.md#L41](docs/commands/show.md#L41) contains a literal TODO: `"TODO: We need to specify the expectations around new lines."`

### D-4: `find` Command — Regex Support Undocumented

[internal/action/find.go](internal/action/find.go) supports `--regexp` for regex matching, but [docs/commands/find.md](docs/commands/find.md) doesn't mention this capability.

### D-5: `generate` Command — XKCD Flags Partially Documented

`--xkcdcapitalize` and `--xkcdnumbers` exist in the command registration but are not documented in [docs/commands/generate.md](docs/commands/generate.md).

### D-6: Secret Format Parser Order Undocumented

The secret parser in [pkg/gopass/secrets/secparse/](pkg/gopass/secrets/secparse/) tries MIME → YAML → AKV in sequence. This fallback chain is not documented in [docs/secrets.md](docs/secrets.md), which can lead to user confusion when secrets are parsed as the wrong format.

### D-7: Reference Syntax Undocumented

Secrets support `gopass://path/to/other/secret` references (resolved when `core.follow-references` is enabled). This isn't mentioned in user-facing documentation.

### D-8: ARCHITECTURE.md Has Placeholder

[ARCHITECTURE.md#L99](ARCHITECTURE.md#L99) contains: `"TODO: There is a lot to be said about this package, e.g. custom errors."`

### D-9: `show.safecontent` Deprecation Status Unclear

[docs/features.md](docs/features.md) says safecontent "is not perfect and might be removed in the future", yet [docs/config.md](docs/config.md) documents it as an active feature. Its status should be clarified.

---

## CLI / UX Issues

### U-1: `--force` / `-f` Means Different Things Across Commands

| Command | `--force` / `-f` meaning |
|---------|--------------------------|
| `show` | Ignore `safecontent`, display password (aliased to `--unsafe`) |
| `copy`, `move`, `delete` | Overwrite without prompting (skip confirmation) |
| `generate` | Overwrite existing password without confirmation |

The `show` command registers `--force` as an alias for `--unsafe`:
```go
Aliases: []string{"u", "force", "f"},
```
This means `-f` silently changes semantics when a user switches between commands. A user accustomed to `gopass rm -f` (force delete) might think `gopass show -f` forces some output mode — and accidentally disable safecontent.

**Recommendation:** Remove `force`/`f` as aliases for `unsafe` in `show`. They have independent, established semantics.

### U-2: Four Overlapping Secret Creation Commands

Users must choose between `create`, `insert`, `edit -c`, and `generate`, with no clear guidance:

| | `create` | `insert` | `edit` | `generate` |
|---|---|---|---|---|
| Creates new secret | ✓ | ✓ | ✓ (with `-c`) | ✓ |
| Wizard-guided | ✓ | | | |
| Opens editor | | with `-m` | ✓ | with `-e` |
| Auto-generates password | | | | ✓ |
| Append to existing | | ✓ (with `-a`) | | |

**Recommendation:** Add a "Getting Started" section to help text or output a suggestion when users run `gopass` for the first time. Consider making `create` the unified entry point that offers all four modes.

### U-3: `--alsoclip` Naming and Alias

The flag `--alsoclip` / `-C` (capital C) is non-intuitive:
- It combines `--clip` (copy) with showing the full secret
- The name doesn't follow any established convention
- Capital `-C` is unusual as a short flag

**Recommendation:** Consider `--clip-and-show` or make `--clip` accept an optional `--show` modifier.

### U-4: XKCD Generator Flags Are Inconsistent

```
--sep / -xs / --xkcdsep        (has two long forms)
--lang / -xl / --xkcdlang      (has two long forms)
--xkcdcapitalize               (no short form)
--xkcdnumbers                  (no short form)
```

Short aliases `-xs` and `-xl` don't follow the single-letter convention. Some XKCD flags have multiple long forms while others have none.

**Recommendation:** Standardize: `--xkcd-sep`, `--xkcd-lang`, `--xkcd-capitalize`, `--xkcd-numbers` with `-S`, `-L` short forms if needed.

### U-5: `audit` Default Output Mode

Running `gopass audit` without `--full` or `--summary` performs the full audit, then discards all output with only a warning. Since the audit can take significant time (decrypts every secret), this is a frustrating user experience.

**Recommendation:** Default to `--summary` when neither flag is specified.

### U-6: `--force-regen` / `-t` Alias Is Cryptic

In `generate`, the `-t` short alias for `--force-regen` has no mnemonic connection. Users are unlikely to discover or remember it.

### U-7: Deprecated `GOPASS_AUTOSYNC_INTERVAL` Not Clearly Removed

The code in [internal/action/sync.go#L35](internal/action/sync.go#L35) still processes `GOPASS_AUTOSYNC_INTERVAL` with a deprecation log message, but this env var is not mentioned in current docs. Users relying on it won't know to switch to `autosync.interval`.

---

## Architecture & Abstraction Concerns

### A-1: `Action` Struct Is a God Object

The `Action` struct in [internal/action/](internal/action/) has **100+ methods** spanning every CLI concern: secret CRUD, mount management, git operations, auditing, templates, recipient management, OTP, completion generation, REPL, binary operations, cloning, and more.

This makes the struct:
- Impossible to test in isolation (every test needs the full store infrastructure)
- Hard to navigate (methods spread across 30+ files with no grouping beyond filename)
- Resistant to refactoring (every new feature adds another method to the same struct)

**Recommendation:** Split into focused handler types grouped by domain (e.g., `SecretHandler`, `MountHandler`, `RecipientHandler`, `AuditHandler`). The `Action` struct can delegate to these.

### A-2: Context Keys Used as Configuration System

[pkg/ctxutil/ctxutil.go](pkg/ctxutil/ctxutil.go) defines **28+ context keys** used to pass configuration through the call stack. Each key requires ~4 helper functions (`WithX`, `HasX`, `IsX`, `GetX`). Some carry **callback functions** (`PasswordCallback`, `ImportFunc`, `ProgressCallback`), which is an anti-pattern for Go contexts.

Go contexts are designed for request-scoped values (deadlines, cancellation, request IDs), not for application configuration. This pattern:
- Defeats type safety (all values are `interface{}`)
- Makes dependencies invisible (no function signature shows what config it reads)
- Creates testing overhead (each test must build context chains)

**Recommendation:** Consolidate into a typed `ExecutionConfig` struct passed explicitly or stored once in context. Move callbacks to proper dependency injection via struct fields.

### A-3: `Storage` Interface Embeds VCS Operations

The `Storage` interface in [internal/backend/storage.go](internal/backend/storage.go) embeds the `rcs` interface, coupling basic file operations with git-specific methods (commit, push, pull). Backends like `fs` (no VCS) must implement stubs for all VCS methods.

**Recommendation:** Separate `Storage` (file ops) from `VersionControl` (VCS ops). Backends implement one or both.

### A-4: Queue Package Needs Replacement

The [internal/queue/background.go](internal/queue/background.go) package self-documents as "likely broken." The `Idle()` method leaks goroutines, channel operations have race conditions, and error handling is minimal. The `noop` fallback means disabled queuing silently drops tasks.

**Recommendation:** Replace with `errgroup`, `sync.WaitGroup`, or a proper worker pool. The buffered channel pattern is fine but needs correct lifecycle management.

### A-5: `CleanMountAlias` Over-Processing

[internal/store/root/mount.go#L173-L184](internal/store/root/mount.go#L173-L184) loops through prefix/suffix stripping with nested `TrimPrefix`/`TrimSuffix` calls:

```go
for strings.HasPrefix(alias, "/") || strings.HasPrefix(alias, "\\") {
    alias = strings.TrimPrefix(strings.TrimSuffix(alias, "/"), "/")
    alias = strings.TrimPrefix(strings.TrimSuffix(alias, "\\"), "\\")
}
```

This could be `strings.Trim(alias, "/\\")` — a single call handles all leading and trailing separators.

---

## Code Style Inconsistencies

### S-1: Error Message Capitalization

Error messages inconsistently capitalize the first word:

```go
out.Errorf(ctx, "failed to list store: %s", err)   // lowercase
out.Errorf(ctx, "Failed to remove mount: %s", err)  // uppercase
out.Errorf(ctx, "Failed to add file to tree: %s", err)  // uppercase
fmt.Errorf("can not delete a mount point...")        // lowercase
fmt.Errorf("Can not unmount %s: %s", ...)            // uppercase (grep finds both)
```

Go convention is lowercase error messages (per `go vet` / `errcheck`). The codebase uses both.

**Recommendation:** Standardize on lowercase for `fmt.Errorf` returns and capitalized for user-facing `out.Errorf` messages (since those are displayed directly to the user).

### S-2: "can not" vs "cannot"

The codebase uses both forms:
- `"can not delete"` in [internal/store/root/move.go](internal/store/root/move.go)
- `"cannot be reached"` in various docs
- `"can not be used"` in [AGENTS.md](AGENTS.md)

Standard English and Go conventions prefer "cannot" (one word).

### S-3: `sort.Strings()` vs `slices.Sorted()` vs `slices.Sort()`

The codebase uses three different sorting approaches depending on when the code was written:
- `sort.Strings(keys)` — older code
- `slices.Sorted(keys)` — newer code (in `pkg/set/sorted.go`)
- `sort.Slice(cmds, func...)` — custom comparators

This is cosmetic, but new code should prefer `slices.Sort`/`slices.Sorted` (Go 1.21+).

### S-4: `context.TODO()` in Production Code

Four files use `context.TODO()` in non-test code:
- [internal/backend/storage/fs/store.go#L242](internal/backend/storage/fs/store.go#L242) — in `String()` method
- [internal/backend/storage/gitfs/storage.go#L49](internal/backend/storage/gitfs/storage.go#L49) — in `String()` method
- [internal/backend/storage/fossilfs/storage.go#L47](internal/backend/storage/fossilfs/storage.go#L47) — in `String()` method
- [pkg/pinentry/cli/fallback.go#L51](pkg/pinentry/cli/fallback.go#L51) — in `GetPIN()`

The `String()` methods call `Version(context.TODO())` because the `fmt.Stringer` interface doesn't accept context. This is a design tension — the `Version()` method requires context for external commands.

**Recommendation:** Cache the version at construction time instead of computing it in `String()`.

### S-5: Stale TODOs

Several TODOs have been in the code for a long time:

| Location | TODO |
|----------|------|
| [internal/backend/crypto.go#L92](internal/backend/crypto.go#L92) | "should return ErrNotSupported, but need to fix some tests" |
| [internal/store/root/rcs.go#L82](internal/store/root/rcs.go#L82) | "should likely iterate over all stores" |
| [internal/store/leaf/fsck.go#L283](internal/store/leaf/fsck.go#L283) | "report these stats" |
| [internal/store/leaf/crypto.go#L99](internal/store/leaf/crypto.go#L99) | "do not hard code exceptions" |
| [docs/commands/show.md#L41](docs/commands/show.md#L41) | "We need to specify the expectations around new lines" |
| [ARCHITECTURE.md#L99](ARCHITECTURE.md#L99) | "There is a lot to be said about this package" |

### S-6: `autosyncLastRun` as Package Variable Without Synchronization

**Location:** [internal/action/sync.go#L28](internal/action/sync.go#L28)

```go
var autosyncLastRun time.Time  // no mutex
```

While gopass is primarily single-process, this global mutable state could race if autosync triggers from multiple goroutines. Should use `sync.Mutex` or `atomic.Value`.

---

## Improvement Suggestions

### I-1: Unified Secret Name Validation

Currently, name validation is scattered:
- `leaf/write.go` rejects `//` 
- `fs/store.go` uses `filepath.Clean()` (doesn't prevent `..`)
- `binary.go` uses `isFilePath()` heuristic + `isInStore()` check
- `create/wizard.go` uses `CleanFilename()` for path segments

**Suggestion:** Create a single `ValidateSecretPath(storeRoot, name string) error` function that validates all names at the boundary. Check for `..`, resolve against store root, reject escapes.

### I-2: Structured Exit Codes

The `exit` package defines numeric exit codes but the mapping between user-visible behavior and code is implicit. Consider a `--exit-code-help` flag or documenting exit codes for scripting users.

### I-3: Machine-Readable Output

Commands like `list`, `find`, `recipients`, and `audit` could benefit from `--json` output for scripting and integration. The `audit` command already supports `--format csv|html` — extending JSON to other commands would improve the tool's composability.

### I-4: `gopass doctor` / Self-Diagnostic Command

Users frequently encounter issues with GPG configuration, git setup, or missing dependencies. A `gopass doctor` command (similar to `brew doctor`) that checks:
- GPG binary availability and version
- Git configuration (user.name, user.email)
- Store permissions
- Recipient key validity
- Remote connectivity

would significantly reduce support burden.

### I-5: Completion of `show.safecontent` Pattern

The safecontent feature hides passwords by default but the allowlist/blocklist of hidden keys (`password`, `totp`, `hotp`, `otpauth`) is hardcoded. Making this configurable (e.g., `show.hidden-keys`) would let teams adapt it to their secret naming conventions.

### I-6: Better `gopass env` Alternatives

The `env` command (see security report C-2) could offer:
- `--stdin` mode: pipe secrets via stdin instead of environment
- `--file` mode: write to a temporary file, set path in env, auto-cleanup
- `--exec` mode: use `setpriv` or similar to clear env after subprocess exec

### I-7: Wizard Templates as First-Class Feature

The create wizard uses template YAML files but the template format, available attribute types (`string`, `hostname`, `password`, `multiline`), and configuration options (`AlwaysPrompt`, `Strict`, `Min`, `Max`) are not documented for end users who want to create custom templates. Documenting this would make the create wizard much more powerful for teams.

---

*This report was produced through manual static analysis. Findings tagged as "confirmed" were verified by reading source directly. Some architectural recommendations are subjective and should be weighed against the project's maintenance capacity and backward compatibility constraints.*
