# `fsmove` command

Note: The implementations for `fscopy` and `fsmove` are exactly the same. The only difference is that `fsmove` will remove the source after a successful copy.

The `fsmove` command works like the Unix `mv` or `rsync` binaries. It allows moving either single entries or whole folders in and out of the gopass mount.

If the source is a directory, the source directory is re-created at the destination if no trailing slash is found. Otherwise the contained secrets are placed into the destination directory (similar to what `rsync` does).

## Synopsis

```
# Decrypt the secret "leaf" and place in /home/user/.secrets directory
$ gopass fsmove path/to/leaf /home/user/.secrets
# Move the content of path/to/somedir to new/dir/somedir relative to the current working dir
$ gopass fsmove path/to/somedir new/dir
# Move a whole folder of secrets into their proper place from the root of the file-system
$ gopass fsmove path/to/secrets /
```

## Modes of operation

* Move a single secret into or out of the target mount
* Move a folder of secrets, possibly with sub folders, into or out of the target mount

## Flags

Flag | Aliases | Description
---- | ------- | -----------
`--force` | `-f` | Overwrite existing destination without asking.

## Details

* In the case of ambiguity, i.e. a file and secret have the same name, gopass will make no assumption and gracefully print an error. One may specify a "file://" scheme in this case to indicate which argument is meant to target the file-system
