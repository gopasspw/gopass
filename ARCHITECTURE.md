# Architecture

This document describes the high-level architecture of gopass. If you want to
get familiar with the code base you are in the right place.

## Overview

On the highest level gopass manages directories (called `stores` or `mounts`)
that contain (mostly) GPG encrypted text files. gopass transparently handles
encryption and decryption when accessing these files. It applies some heuristics
to parse the file content and support certain operations on that content.

`gopass` is licensed under the terms of the MIT license and we require
compatible licenses from our dependencies as well (when we link against them).

For licensing reasons and security considerations we try to keep the number of
external dependencies (libraries) well-arranged. Try to avoid adding new
dependencies unless absolutely necessary.

## Code Map

This section talks briefly about the various directories and some data
structures.

We're trying to clearly separate between our public API and implementation
details. To that extent we're in the process of moving packages to `internal/`
(and sometimes back to `pkg/`, if necessary).

A note on semantic versioning: `gopass` is both an CLI and an API (Go module).
The expectations around semantic versioning and Go modules make it difficult
to express both concerns in the same versioning scheme, e.g. does a breaking
change in the API require a major version bump even if nothing about the tool
(CLI) has changed? What about the other way round? Thus we have decided to
apply semantic versioning only to the CLI tool, not the Go module. This is not
ideal and might change with sufficient active contributors.

### `docs/backends`

This folder contains documentation about each of our supported backends. See
`internal/backend` below for more information about our backend design.

### `docs/commands`

This folder contains the specification of each sub command the tool offers.
We have many sub commands with sometimes dozens of flags each. In the past we
did encounter some inconsistencies and decided to introduce specifications for
each command. If the specification and the implementation disagree this should
be reported as a bug and fixed or the specification needs to be changed (but the
general assumptions should be that the specification is correct, not the code).

### `docs/usecases`

This directory contains an (incomplete) list of our core use cases, i.e. the
critical user journeys we aim to support. `gopass` can be used in various ways
and try to remain flexible and extensible, but if we encounter a conflict
between a blessed use case and a corner case we prefer the former.

### `helpers/`

This directory contains some release automation tooling that is supposed to be
invoked with `go run`. The changelog generator in `helpers/changelog` is used
by our GitHub Action based release automation and shouldn't be invoked manually.

The tooling in `helpers/release` will prepare a new release and helps to file a
release pull request will all the required updates in place.

### `internal/` and `pkg/`

`gopass` used to not have either of these and all our packages were rooted
directly in the repository. However we began to notice that other projects
were starting to depend directly on our internal packages and we sometimes
broke them. This put us and the other project into an unpleasant
situation so we tried to clarify the expectations by using Go's `internal/`
visibility rule to keep other projects from depending on our implementation
details.

Note: If we have a good reasons to use one of our `internal/` packages either
copy it (our license should rarely be an issue) or nicely ask us and explain
why something should move to `pkg/`.

As we are in the process of formalizing a proper API surface we sometimes need
to move packages from `internal/` to `pkg/`. The other direction might also
occur, but much less often.

### `internal/action`

This directory contains one file, and sometimes sub folders, for each command
`gopass` supports. These are mostly self-contained, but some (e.g. show / edit
/ find) need to depend on each other.

TODO: There is a lot to be said about this package, e.g. custom errors.

### `internal/backend`

`gopass` is built around the concept of multiple independent password stores
that can be mounted into one namespace, much like regular file systems. Each
of these stores can have a different storage and crypto backend. We used to
have independent revision control backends as well, but since the RCS (e.g.
git) interacts so closely with the storage (you can't use regular git w/o a
filesystem-based storage) we have merged storage and RCS backends.

The backend package defines the interfaces for the backend implementation
and provides a registry that returns the concrete backend from the list of
registered ones. Registration happens through blank imports of either the
`internal/backend/crypto` and `internal/backend/storage` packages.

Each backend needs to have a loader implementation in its `loader.go` (please
stick to this name). We try to auto-detect the most applicable backend when
initializing the process, but some backends look alike (e.g. a `fs` and an 
uninitialized `gitfs`). So the loader comes with a priority which is respected
during lookup.

### `internal/config`

TODO: backwards compat, loading, overrides, ...

### `internal/cui`

TODO: history, issues, outlook

### `internal/editor`

TODO: security

### `internal/queue`

TODO: motivation, usage

### `internal/store`

TODO: everything

### `internal/tree`

TODO: why

### `internal/updater`

TODO: security

### `pkg/...`

TODO: everything

### `tests`

TODO: integration tests

