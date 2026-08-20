# Conventions

Normative reference for commit messages, versioning, branch and tag names, and
file naming.

See [CONTRIBUTING.md](../CONTRIBUTING.md) for the contribution workflow and
[docs/hacking.md](hacking.md) for the development environment.

## 1. Commit messages

Commit subjects follow [Conventional Commits 1.0.0](https://www.conventionalcommits.org/en/v1.0.0/).

### 1.1 Format

```
<type>[(<scope>)][!]: <description>

[optional body]

[optional footers]
```

Rules:

* Limit the subject line to 72 characters.
* Write the description in the imperative mood: "add", not "adds" or "added".
* Start the description in lowercase. Do not end it with a period.
* Include a `Signed-off-by:` footer in every commit, as required by
  [CONTRIBUTING.md](../CONTRIBUTING.md). Conventional Commits governs the
  subject line; the Developer Certificate of Origin adds a trailer. Both apply.
* Write the pull request title as a valid Conventional Commit.

The pull request title requirement follows from the merge strategy. Pull
requests are squash-merged: pull request #3489, titled `fix: Avoid NPE when
attempting to edit a non-existing secret`, landed on `master` as the
single-parent commit `eb1368e4` with the subject `fix: Avoid NPE when
attempting to edit a non-existing secret (#3489)`. The pull request title, not
the individual commit subjects, is the string that reaches `CHANGELOG.md`.

### 1.2 Types

This list is exhaustive. Reject any other type.

| Type | Meaning | SemVer impact | Changelog section |
|---|---|---|---|
| `feat` | New user-visible capability | MINOR | Added |
| `fix` | Bug fix | PATCH | Fixed |
| `security` | Security fix or hardening | PATCH | Security |
| `perf` | Performance improvement, no behaviour change | PATCH | Changed |
| `refactor` | Internal restructuring, no behaviour change | PATCH | Changed |
| `revert` | Reverts a previous commit | PATCH | Changed |
| `deps` | Dependency version change | PATCH | Changed |
| `docs` | Documentation only | none | omitted |
| `test` | Tests only | none | omitted |
| `build` | Build system: Makefile, goreleaser, Dockerfile | none | omitted |
| `ci` | GitHub Actions, linter configuration, workflows | none | omitted |
| `chore` | Housekeeping that fits nothing above | none | omitted |

Two entries extend the set suggested by the specification:

* `security` populates the `Security` section of the changelog, which
  Keep a Changelog defines as a first-class category. Five commits since
  2025-01-01 already use it.
* `deps` is the preferred type for hand-written dependency commits.
  Accept `chore(deps)` as well: Dependabot emits it, and 147 of the 344
  commit subjects since 2025-01-01 carry the `chore` type, nearly all of
  them dependency bumps.

The following appear as commit types in the history and are not types. Use them
as scopes: `otp` (2 commits since 2025-01-01), `age`, `bug`, `fscopy`, and
`openbsd` (1 each).

### 1.3 Scopes

The scope is optional. Prefer to supply one. Do not invent scopes; omit the
scope when none fits. Omit the scope for a change spanning several areas rather
than listing more than one.

**Commands** — the 41 top-level subcommands: 39 registered by
`(*Action).GetCommands` in `internal/action/commands.go`, plus `pwgen` from
`internal/action/pwgen` and `completion`, both added by `getCommands` in
`main.go`:

```
alias audit cat clone completion config convert copy create delete doctor edit
env find fsck fscopy fsmove generate grep history init insert link list merge
mounts move otp process pwgen rcs recipients reorg setup show sum sync
templates unclip update version
```

**Backends:** `age`, `gpg`, `plain`, `cryptfs`, `fossilfs`, `fs`, `gitfs`, `jjfs`

**Subsystems:** `action`, `audit`, `backend`, `cache`, `completion`, `config`,
`create`, `cui`, `editor`, `hashsum`, `hook`, `notify`, `out`, `queue`,
`recipients`, `reminder`, `store`, `store/leaf`, `store/root`, `tpl`, `tree`,
`updater`

**Public API:** `pkg/gopass`, `api`, `secrets`, `appdir`, `clipboard`,
`ctxutil`, `debug`, `fsutil`, `otp`, `passkey`, `pinentry`, `protect`, `pwgen`,
`qrcon`, `set`, `tempfile`, `termio`

**Meta:** `deps`, `release`, `changelog`, `docs`, `adr`, `ci`, `build`

Write `pkg/gopass` out in full as a scope. Public-API changes must be greppable
against the [A-12](adr/A-12-pkg-api-stability.md) stability contract.

### 1.4 Breaking changes

This repository has two independent compatibility surfaces. Mark them
differently. Do not conflate them.

| Surface broken | Marker | Version impact |
|---|---|---|
| gopass CLI: commands, flags, output format, exit codes, config keys, store format | `!` after the type or scope, **and** a `BREAKING CHANGE:` footer | MAJOR |
| Go module `pkg/gopass` only: an exported symbol removed or changed | a `PKG-BREAK:` footer, **no** `!` | none; MINOR at most |
| Both | `!` + `BREAKING CHANGE:` + `PKG-BREAK:` | MAJOR |

Do not mark a module-only break with `!`. No CLI user can observe such a
change, and `!` demands a major release.

Example of a module-only break, per [A-12](adr/A-12-pkg-api-stability.md):

```
refactor(pkg/gopass): drop Store.GetRevision in favour of Store.History

Deprecated since 1.17.0; the two-minor / three-month window has elapsed.

PKG-BREAK: pkg/gopass: remove Store.GetRevision, use Store.History instead
Signed-off-by: Your Name <your@example.com>
```

The `PKG-BREAK:` footer is the machine-readable form of the `[PKG-BREAK]`
changelog tag mandated by A-12.

### 1.5 Changelog control footers

| Footer | Effect |
|---|---|
| `Changelog-Section: Added\|Changed\|Deprecated\|Removed\|Fixed\|Security` | Overrides the type-to-section mapping. Required for `Removed` and `Deprecated`, which have no corresponding commit type. |
| `RELEASE_NOTES=<text>` or `RELEASE_NOTES=n/a` | Overrides the bullet text, or suppresses the entry. |
| `PKG-BREAK: <text>` | Emits a `[PKG-BREAK]`-prefixed bullet. |

## 2. Versioning

This project follows [Semantic Versioning 2.0.0](https://semver.org/spec/v2.0.0.html).

### 2.1 Covered by SemVer

SemVer applies to the gopass command line interface and its observable
behaviour: subcommands, flags and their aliases, the stdout and stderr
contracts, JSON output schemas, exit codes ([docs/exit-codes.md](exit-codes.md)),
configuration keys ([docs/config.md](config.md)), the on-disk store format, and
the generated shell completions and man page.

### 2.2 Not covered by SemVer

The Go module `github.com/gopasspw/gopass` carries no independent semver
guarantees. `pkg/gopass` is governed by [A-12](adr/A-12-pkg-api-stability.md)
and by the package documentation in `pkg/gopass/doc.go`, which declare it
best-effort stable:

* Additive changes — new exported symbols, new functional-option parameters —
  may appear in any release.
* Breaking changes — removal or signature change of an exported symbol, change
  to an interface method set or to error semantics — require a `[PKG-BREAK]`
  entry in `CHANGELOG.md` and a deprecation window of two minor releases or
  three months, whichever is longer.

### 2.3 Precedence

Release a change that breaks only `pkg/` consumers as MINOR or PATCH. Do not
bump MAJOR for it. Release a change that breaks CLI users as MAJOR regardless
of its `pkg/` impact.

### 2.4 Pre-releases

Use `vX.Y.Z-rc.N` only: SemVer pre-release identifiers, dot-separated, numeric.
Sign the tag. Cut it from the merged `release/vX.Y.Z-rc.N` branch. See
[docs/releases.md](releases.md).

## 3. Branch names

```
<type>/<kebab-slug>[-<issue>]
```

* Use a commit type from §1.2 as `<type>`.
* Restrict the slug to `[a-z0-9-]`, plus the single `/` separator. Use no spaces
  and no underscores. Limit the whole name to 60 characters.
* Append the issue number when one exists:
  `refactor/separate-storage-rcs-3411`.
* Do not create these prefixes by hand; the release automation owns them:
  `release/vX.Y.Z` and `release/vX.Y.Z-rc.N`, created by
  `helpers/release/main.go`, and `prep/vX.Y.Z`, documented in
  [docs/releases.md](releases.md).
* Do not push feature branches to `gopasspw/gopass`. Work on a fork.

## 4. Tag names

* Name releases `vX.Y.Z` and pre-releases `vX.Y.Z-rc.N`.
* Sign every tag (`git tag -s`).
* Create no other tags in this repository.

`.github/workflows/autorelease.yml` triggers on `push: tags: 'v*'`.
`helpers/release/main.go` strips the leading `v` and parses the remainder as
semver.

## 5. File names

### 5.1 General rules

* Use lowercase kebab-case, restricted to `[a-z0-9._-]`.
* Use no spaces.
* Zero-pad any numeric component that participates in lexical ordering.
* Prefix with `YYYY-MM-DD` when chronological order matters.
* Order components from most general to most specific, left to right.

### 5.2 Architecture decision records

Name records `docs/adr/A-NN-<kebab-slug>.md`, with `NN` zero-padded to two
digits.

* Write the H1 as `# A-NN: <Title>`, matching the file name.
* Open the record with a `**Status:**` line — `proposed`, `accepted`,
  `deferred`, `implemented`, `partially implemented`, or `superseded by A-NN` —
  and a `**Source:**` line.
* Never reuse a number. Supersede a decision by writing a new record.
* Update [docs/adr/README.md](adr/README.md) in the same commit that adds or
  supersedes a record.

### 5.3 Other documentation

* Name command specifications `docs/commands/<command>.md`, where the file name
  is exactly the command name.
* Name backend documents `docs/backends/<backend>.md` and use cases
  `docs/usecases/<kebab-slug>.md`.
* Use lowercase throughout. Keep the SCREAMING_CASE names of root-level
  documents, which is the GitHub convention.

### 5.4 Go files

* Use lowercase names with no underscores, except for the reserved suffixes
  below. Prefer a single word matching the primary type or the command.
* Place a command implementation in `internal/action/<command>.go` with an
  accompanying `internal/action/<command>_test.go`.
* Reserved suffixes: `_test.go`; the `GOOS` and `GOARCH` suffixes `_unix.go`,
  `_windows.go`, `_linux.go`, `_darwin.go`, `_others.go`; `_gen.go` for
  generated code; `_fuzz.go` for fuzz targets.
* Give a build-tag variant that is not `GOOS`-based a descriptive suffix and an
  explicit `//go:build` line, for example `screenshot_supported.go` and
  `screenshot_stub.go`.
* Split a file above roughly 600 lines along an existing type or feature seam.

## 6. Changelog

`CHANGELOG.md` follows [Keep a Changelog 1.1.0](https://keepachangelog.com/en/1.1.0/):
an `## [Unreleased]` section at the top, then one `## [X.Y.Z] - YYYY-MM-DD`
section per release, each containing only the non-empty subsections of
`Added`, `Changed`, `Deprecated`, `Removed`, `Fixed`, and `Security`.

`helpers/release` generates the entries from commit subjects at release time,
using the mapping in §1.2 and the footers in §1.5. Do not edit `CHANGELOG.md`
by hand, with two exceptions. Add a hand-written `## [Unreleased]` entry for:

* any change to the exported surface of `pkg/`, as required by
  [A-12](adr/A-12-pkg-api-stability.md); and
* any `security:` change.
