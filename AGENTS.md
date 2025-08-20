# Project Overview

gopass is a command line application that allows users to managed their passwords and other secrets inside encrypted files. Those files are usually encrypted using gpg (but other backends like age do exist). The files are usually managed using git (but other VCS backends exist as well). The CLI is primarily intended for human users.

Several integration exist, these are stand alone projects that use the exposed gopass API to interact with an existing password store.

gopass supports multiple password stores. It requires at least one root store but any number of additional stores can be mounted, just like filesystems on Linux, inside the root store. Each store can use a different encryption method and VCS.

The primary use case of using different password stores is to encrypt and share the content with a different set of recipients.

The project is specifically targeting users on all major platform, i.e. Linux, Unix, MacOS and Windows.

## Folder Structure

- `/docs`: Contains human readable documentation for the project.
- `/helpers`: Contains tools used to maintain the project. Users usually don't use those, these are mainly for developers and maintainers of the project. Do not touch this directory unless instructed to do so.
- `/internal`: Contains most of the implementation of the project. It is visibility restricted so other projects can not depend on it and we can be very liberal with breaking changes.
- `/pkg`: Contains the public API (inside `/pkg/gopass`) used by our integrations and other projects as well as necessary support packages to make using the API feasible.
- `/tests`: Contains only integration tests, i.e. those mock a real GPG-based gopass installation. They are quite slow but provide kind of a regression testing. Remember to add or adjust those when adding major new features.
- `/internal/action`: Contains the different CLI subcommands. Usually one file per top-level subcommand (e.g. the implementation for `gopass ls` is in `/internal/action/list.go`) with an accompanying `_test.go` file that contains the unit tests. All commands need to be registered in `/internal/action/commands.go`.
- `/internal/audit`: Contains the audit code that checks password stores for weak passwords or related issues.
- `/internal/backend`: Contains the different backend implementations for both encryption as well as version controlled storage. Storage implementations need to register themselves in `/internal/backend.StorageRegistry` while encryption backends need to register in `/internal/backend.CryptoRegistry`.
- `/internal/backend/crypto/age`: Contains the `age` encryption backend. It is a pure-Go implementation. Refer to the [docs](docs/backends/age.md) as well.
- `/internal/backend/crypto/gpg/cli`: Contains the `gpg` encryption backend. It mostly uses the `gpg` binary to support the different configurations (e.g. smart cards) which wouldn't be possible with existing pure-Go implementation. Refer to [docs](docs/backends/gpg.md) as well.
- `/internal/backend/crypto/plain`: Contains the plaintext backend (no encryption). This should only be used for testing. Users should never use this.
- `/internal/backend/storage/fossilfs`: Contains an experimental storage backend using the Fossil SCM. It might be removed in the future.
- `/internal/backend/storage/fs`: Contains a storage backend without SCM integration, i.e. it simply writes to files on disk without versioning support. Should usually only be used for tests or if users have some kind of transparent versioning system underneath.
- `/internal/backend/storage/gitfs`: Contains the primary storage backend that is using `git` to manage files.
- `/internal/config`: Contains our custom config handling. It is based on the git configuration file format as implemented by our [gitconfig](http://github.com/gopasspw/gitconfig) package. When reading config settings prefer to using `config.Bool(ctx, key)`, `config.String(ctx, key)` or `config.Int(ctx, key)`. Use the low-level methods only when those are not sufficient. Avoid touching the `legacy` package underneath unless asked to.
- `/internal/out`: Contains our output helpers. Prefer those over Go standard lib packages (like fmt) for consistency.
- `/internal/store`: Contains the core of the password store implementation (utilizing the configured backends).
- `/internal/store/root`: Contains the root store. This always exist once in a gopass process. It delegates most operations to one or more leaf stores.
- `/internal/store/leaf`: Contains the leaf store. There must be at least one initialized leaf store per gopass instance. But there can be as many as necessary.
- `/pkg/appdir`: Contains a facility for providing system-dependentt paths for application resources, like config or cache directories. It does honor the `GOPASS_HOMEDIR` variable. This is very useful for testing since a gopass instance running with this variable set to a temporary location will not interfere with the actual production instance a user might be using.
- `/pkg/clipboard`: Contains methods to interact with clipboards on all major operating systems. It is using our [clipboard](http://github.com/gopasspw/clipboard) package. It also supports clearing the clipboard after a given interval.
- `/pkg/ctxutil`: Provides the necessary plumbling to interact with config values stored in the context. Avoid adding new context keys if possible and prefer config values. But if adding context keys is necessary they should only be defined in this file.
- `/pkg/debug`: Contains a debug package with different verbosity levels. Use it to output debug information to a debug log.
- `/pkg/fsutil`: Contains various helpers for interacting with the filesystem, e.g. checking for presence of files or directories. Prefer those over implementing these checks from scratch.
- `/pkg/gopass`: Contains the public gopass API to interact with existing password stores. The `api` sub package contains the actual API and the `secrets` sub package the different secret types we support.
- `/pkg/pwgen`: Contains a pure-Go implementation of the `pwgen` utility.
- `/pkg/set`: Contains a generic set type.
- `/pkg/tempfile`: Contains utility functions for creating and dealing with temp files. It attempts to be more secure than the normal temp file functions from the stdlib. Prefer those over the stdlib.
- `/pkg/termio`: Contains functions for interacting with the user of the terminal.

## Libraries and Frameworks

- Avoid introducing new external dependencies unless absolutely necessary.
- If a new dependency is required, please state the reason.
- The project is licensed under the terms of the MIT license and we can only add compatible licenses. See [.license-lint.yml](.license-lint.yml) for a list of compatible licenses.
- We must avoid introducing CGo dependencies since this make cross-compiling infeasible.

## Testing instructions

- Always run `make test` and `make codequality` before submitting.
- Run `make fmt` to properly format the code. Run this before `make codequality`.
- Before mailing a PR run `make test-integration`

