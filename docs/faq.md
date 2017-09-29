# FAQ

* *How does gopass relate to HashiCorp vault?* - While [Vault](https://www.vaultproject.io/) is for machines, `gopass` is for humans [#7](https://github.com/justwatchcom/gopass/issues/7)
* `gopass show secret` displays `Error: Failed to decrypt` - This issue may happen if your gpg setup if broken. On MacOS try `brew link --overwrite gnupg`. You also may need to set `export GPG_TTY=$(tty)` in your `.bashrc` [#208](https://github.com/justwatchcom/gopass/issues/208), [#209](https://github.com/justwatchcom/gopass/issues/209)
* *gopass recpients add fails with Warning: No matching valid key found* - If the key you're trying to add is already in your keyring you may need to trust it. If this is your key run `gpg --edit-key [KEYID]; trust (set to ultimate); quit`, if this is not your key run `gpg --edit-key [KEYID]; lsign; save; quit`

## API Stability

`gopass` is provided as an CLI program, not as a library. While we try to make the
packages usable as libraries we make no guarantees whatsoever with respect to
the API stability. The `gopass` version only reflects changes in the CLI commands.

If you use `gopass` as a library be sure to vendor it and expect breaking changes.

## Further Reading

* [GPGTools](https://gpgtools.org/) for macOS
* [GitHub Help on GPG](https://help.github.com/articles/signing-commits-with-gpg/)
* [Git - the simple guide](http://rogerdudler.github.io/git-guide/)

