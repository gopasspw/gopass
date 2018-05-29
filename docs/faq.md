# FAQ

* *How does gopass relate to HashiCorp vault?* - While [Vault](https://www.vaultproject.io/) is for machines, gopass is for humans [#7](https://github.com/gopasspw/gopass/issues/7)
* `gopass show secret` displays `Error: Failed to decrypt` - This issue may happen if your GPG setup if broken. On MacOS try `brew link --overwrite gnupg`. You also may need to set `export GPG_TTY=$(tty)` in your `.bashrc` [#208](https://github.com/gopasspw/gopass/issues/208), [#209](https://github.com/gopasspw/gopass/issues/209)
* *gopass recipients add fails with Warning: No matching valid key found* - If the key you're trying to add is already in your key ring you may need to trust it. If this is your key run `gpg --edit-key [KEYID]; trust (set to ultimate); quit`, if this is not your key run `gpg --edit-key [KEYID]; lsign; save; quit`
* *How can gopass handle binary data?* - gopass is designed not to change to content of the secrets in any way except that it will add a final newline at the end of the secret if it does not have one already and the output is going to a terminal. This means that the output may mess up your terminal if it's not only text. In this case you should either encoded the secret to text (e.g. base64) before inserting or use the special `gopass binary` sub command that does that for you.
* *Why does gopass delete my whole KDE klipper history?* - KDEs klipper provides a clipboard history for your convenience. Since we currently can't figure out which entry may contain a secret copied to the clipboard, we just clear the whole history once the clipboard timer expires.
* *Can I use gopass as an token helper for Vault?* - Yes, there is [a repo](https://github.com/frntn/vault-token-helper-gopass) that provides the necessary scripts and instructions.
* *Does gopass support re-encryption?* - Adding or removing recipients with `gopass recipients add` or `gopass recipients remove` will automatically re-encrypt all affected secrets.
* *gopass can automatically import missing recipient keys but can it export them as well?* - When adding a recipient with `gopass recipients add` his public key will automatically exported to the store `.gpg-keys/<ID>`.
* *gopass fails with 'gpg failed to sign the data fatal: failed to write commit object'* - If git is not able to interface with GPG commits will fail because gopass configures git to sign commits by default. Have a look a [this question](https://stackoverflow.com/questions/39494631/gpg-failed-to-sign-the-data-fatal-failed-to-write-commit-object-git-2-10-0) for more information.
* *Can gopass be used with Terraform?* - Yes, there is a gopass-based [Terraform provider](https://github.com/camptocamp/terraform-provider-pass) available.

## API Stability

gopass is provided as an CLI program, not as a library. While we try to make the packages usable as libraries we make no guarantees whatsoever with respect to the API stability. The gopass version only reflects changes in the CLI commands.

If you use gopass as a library be sure to vendor it and expect breaking changes.

## Further Reading

* [GPGTools](https://gpgtools.org/) for MacOS
* [GitHub Help on GPG](https://help.github.com/articles/signing-commits-with-gpg/)
* [Git - the simple guide](http://rogerdudler.github.io/git-guide/)
