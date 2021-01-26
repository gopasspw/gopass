# `gitfs` storage backend

This is the default storage backend. It stores the encrypted data directly in the filesystem. It uses an external git binary to provide history and remote sync operations.

gopass configures git to use persistent ssh connections. If you do not want
this set `GIT_SSH_COMMAND` to an empty string to override the built-in default.
