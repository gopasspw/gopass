# semdup

`semdup` reports named production functions with identical normalized Go SSA.
Use the report to select candidates for manual review. Do not treat a match as
an instruction to merge functions.

The analyzer preserves constants, operators, call targets, control-flow edges,
and a conservative side-effect summary. It ignores function names, parameter
names, local value names, and source positions. It excludes tests, generated
SSA wrappers, initializers, and anonymous closures.

Run the analyzer from the repository root:

```sh
go run ./helpers/semdup \
  -exclude github.com/gopasspw/gopass/helpers \
  -min-instructions 2 \
  ./...
```

Generate machine-readable output:

```sh
go run ./helpers/semdup \
  -json \
  -exclude github.com/gopasspw/gopass/helpers \
  ./... > semdup.json
```

Classify every match as one of:

- common internal primitive;
- intentional adapter;
- compatibility facade;
- platform contract;
- similar implementation with different semantics;
- deletion candidate.

Confirm a proposed consolidation with differential or property tests. The SSA
fingerprint does not prove equivalence for reflection, unsafe operations,
external process behavior, environment state, or undocumented caller
contracts.
