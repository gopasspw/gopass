# A-12: Structured Debug Logging via log/slog

**Status:** proposed — decision pending maintainer input  
**Source:** Architectural Review 2026-05-03 (filed as issue #3416)

---

## Background

`pkg/debug/` implements conditional debug output using printf-style calls:
`debug.Log(format, args...)` and `debug.V(n).Log(format, args...)`. Output is
activated by `--debug` / `GOPASS_DEBUG` and written to stderr as unstructured
text.

`log/slog` (stdlib since Go 1.21; the project requires Go 1.25) provides
structured, levelled logging with pluggable handlers (text or JSON) at zero
additional dependency cost.

---

## A-12-1: Unstructured output limits observability

**Location:** `pkg/debug/` and all `debug.Log(...)` call sites across `internal/`

Printf-format debug lines cannot be reliably parsed by log aggregators or CI
tooling. Correlating events across a gopass invocation requires string matching
rather than field queries.

**Assessment:** Low severity for interactive CLI use; meaningful improvement
for automated environments (scripted pipelines, integration test harnesses,
issue reproduction).

---

## A-12-2: Options

### Option A — Migrate pkg/debug to log/slog (recommended)

Replace `pkg/debug`'s internal implementation with `slog` as the backing
logger:

- `debug.Log(format, args...)` → `slog.Debug(fmt.Sprintf(format, args...))`
  at the package boundary, or restructured as `slog.Debug(msg, key, val, ...)`
  where individual call sites permit.
- `debug.V(n).Log(...)` → `slog.Log(ctx, slog.LevelDebug-slog.Level(n), ...)`
  with a minimum-level filter.
- Add `--log-format json` flag (or `GOPASS_LOG_FORMAT=json` config key) to
  switch to `slog.NewJSONHandler(os.Stderr, nil)`.

The public surface of `pkg/debug` (`Log`, `V`, `LogN`) need not change; the
migration can be entirely internal to that package, leaving all call sites
across `internal/` untouched.

**Effort:** ~1 day. Internal change to `pkg/debug`; no call-site churn.

### Option B — Keep current format, no change

Retain printf-style output. Acceptable if interactive CLI use is the only
target and CI log parsing is not a priority.

**Effort:** Zero.

### Option C — Add JSON flag only, no slog migration

Add a `--log-format json` flag that manually formats `debug.Log` output as
`{"level":"debug","msg":"..."}` without adopting slog internally.

**Effort:** ~2 hours. Avoids slog internally but duplicates log-formatting
logic and forecloses structured key-value improvements at call sites.

---

## A-12-3: Recommendation

Option A. The `pkg/debug` public API is unchanged; the migration is confined to
that package. The benefit (machine-readable output, structured key-value pairs,
no new dependency) is disproportionate to the effort. The `--log-format json`
flag can be added in the same PR.

Incremental improvements — restructuring individual `debug.Log(format, ...)`
calls to proper `slog.Debug(msg, key, val, ...)` pairs across `internal/` —
can follow as a separate, lower-priority pass and do not need to block the
initial migration.
