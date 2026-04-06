# A-4: Fix `grep` Match and Error Counters

**Status:** open — not yet fixed  
**Source:** SECURITY_AUDIT_REPORT.md § M-1, § Q-5

---

## Background

`internal/action/grep.go` declares two integer counters, `matches` and
`errors`, intended to tally the number of matching secrets and decryption
failures respectively. Neither counter is ever incremented inside the loop:

```go
var matches int
var errors int
for _, v := range haystack {
    sec, err := s.Store.Get(ctx, v)
    if err != nil {
        out.Errorf(ctx, "failed to decrypt %s: %v", v, err)
        // errors++ missing here
        continue
    }
    if matchFn(string(sec.Bytes())) {
        out.Printf(ctx, "%s matches", color.BlueString(v))
        // matches++ missing here
    }
}
out.Printf(ctx, "\nScanned %d secrets. %d matches, %d errors", len(haystack), matches, errors)
```

As a result the summary line always reads `0 matches, 0 errors` regardless of
what the search actually found.

---

## Impact

This is a **correctness bug**, not a security vulnerability. Users who rely on
the summary count to verify grep results will receive misleading output and
cannot tell whether any secrets matched their query.

---

## Decision

Fix by adding `matches++` inside the `matchFn` branch and `errors++` inside
the error branch. This is a trivial one-line change per counter. No API or
behaviour changes are required.

---

## Implementation

In `internal/action/grep.go`, locate the loop body and add the two missing
increment statements:

```go
if err != nil {
    out.Errorf(ctx, "failed to decrypt %s: %v", v, err)
    errors++       // add this
    continue
}
if matchFn(string(sec.Bytes())) {
    out.Printf(ctx, "%s matches", color.BlueString(v))
    matches++      // add this
}
```

A test should assert that the summary line reports the correct counts after a
search against a known fixture store.
