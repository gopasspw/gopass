# A-5: Template Engine Uses `text/template` Instead of `html/template`

**Status:** deferred — current design is acceptable given the constraints  
**Source:** SECURITY_AUDIT_REPORT.md § M-2

---

## Background

The gopass template engine (`internal/tpl/`) uses Go's `text/template` package.
Go's standard library ships two closely related template packages:

| Package | Auto-escaping | Intended use |
|---------|--------------|--------------|
| `text/template` | No | Arbitrary text generation |
| `html/template` | Yes (HTML/JS/CSS contexts) | HTML document generation |

`html/template` is often recommended for security-sensitive contexts because it
prevents XSS by automatically escaping output that is interpolated into HTML
attributes, script blocks, and URL parameters.

---

## Why `html/template` Does Not Directly Apply Here

gopass templates generate **plain text secrets**, not HTML documents. Switching
to `html/template` would:

1. **Corrupt non-HTML output.** Characters like `<`, `>`, `&`, `"`, and `'`
   are common in passwords and secret values. `html/template` would HTML-escape
   these on output (e.g. `&` → `&amp;`), producing incorrect secrets.

2. **Provide no meaningful security benefit.** The auto-escaping in
   `html/template` protects against XSS injection into HTML pages. The gopass
   template engine renders secrets to a terminal or file — not a browser. There
   is no HTML parsing context to exploit.

The current `text/template` approach is therefore **correct for the use case**.

---

## Residual Risk

`text/template` allows calling methods on any value passed into the template.
The current `payload` struct passed to templates contains only string fields,
limiting the callable surface to string methods. The risk would escalate if:

- `payload` gains a field or method that returns a complex type with
  side-effect-bearing methods.
- A template function is added whose return type exposes dangerous methods.

---

## Decision

Keep `text/template`. Document the constraint here so that future contributors:

1. Do **not** add methods with observable side effects to the `payload` struct.
2. Audit any new template function's return types for unexpectedly callable
   methods.
3. Reconsider this decision if the template engine ever grows an HTML rendering
   mode (e.g. for browser integrations), in which case a separate HTML-specific
   template path using `html/template` should be introduced rather than
   switching the existing engine wholesale.
