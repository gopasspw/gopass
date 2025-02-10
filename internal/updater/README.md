# Release signing and key rollover documentation

Audience: core maintainers

This document captures the necessary steps to perform regular (usually every 2nd year) release signing key rollover.

## Updating the gopass binary

The gopass self-updater is invoked when calling `gopass update`. It works only if the binary is writable by the user running the command. It is specifically not designed to update any gopass
packages installed by a package manager.

The updater first tries to ensure that it is supposed to update the binary (usually because it can write to the binary location) and then fetches the latest release from GitHub. If this ever causes trouble we could cache this info and proxy requests through gopass.pw.

If there is a new release it will fetch both `SHA256SUMS` and `SHA256SUMS.sig` assets from the latest release and verify the signature matches one of the built-in updater keys.

If the checksum file is verified we continue to fetch the actual binary archive and compare that against
the (verified) checksum file and replace the binary.

All of this is implemented by the files in this directory.

## Publishing assets for the updater during releases

When a new release is cut we rely on GoReleaser and GitHub Action workflows to update all necessary assets.
The configuration for those is spread across the repository and the GitHub Action configuration.

A new release is published by pushing a version tag (`v*`) to the repository. Once that happens the GHA workflow `autorelease.yml` is kicked off. It is configured in `.github/workflows/autorelease.yml` and through a number of injected environment variables from the GHA settings. Most importantly `GPG_PRIVATE_KEY` which contains the armored
GPG private part of the current release signing key and the respective passphrase in `PASSPHRASE`.

GoReleaser is controlled by `.goreleaser.yml` in the root of this repository. The relevant sections there are `checksum` to ensure a checksum file is generated and the `signs` section to sign the checksum file using the provided `GPG_FINGERPRINT` in the workflow.

## Managing keys and related assets

The relase signing key is set to expire every other year, so we need to follow a certain key rotation protocol to allow for a seamless key rotation.

* At T-6 Month we should notice that `TestGPGVerifyIn6Months` starts to fail.
    * There is likely a loss obstrusive way to achieve that, but I'll leave it at that for now.
* We should then create an issue to track the key rollover (this should never happen in secret). The entire security posture isn't perfect but that's the best I can do with my resources. Help always appreciated.
* For the actual rollout we first need to generate a new key. That needs to be done by exactly one core maintainer with write access to the repo and the GHA secrets since only they can inject the new key, fingerprint and passphrase.
* To generate the key run: `gpg --expert --full-generate-key` and select `RSA and RSA`, `3072` (bits) and a validity of `2y`. Use `Gopass Release Signing Key YYYY` as the name, `GitHub Actions only` as the comment and `release@gopass.pw` as the email.
  * Note: If you're correctly following the 6 Months advance notice process, use the next year for the name.
  * RSA isn't perfect but we used to have some compatability issues with non-RSA keys. Feel free to revisit this in the future. Keep the keep size to a reasonable value. Last time I checked 4096 did seem a bit excessive and with different algorithms these numbers will need to change as well. In doubt the `BSI TR-02102-1` should have a reasonable recommendation.
  * Use a strong, random passphrase. Since you should never have to type it anywhere make it long, cryptic and stop worrying about it, e.g. `gopass pwgen 32`.
* If you are me, you should probably also save a copy of both parts of new key into your gopass maintainer password store. For convenience add the Key ID / Fingerprint into the secret so you don't have to import the key just to get the Key ID.
  * Hint: `gpg --output 0xKEYID.pub --armor --export 0xKEYID` and `gpg --output 0xKEYID.private --armor --export-secret-key 0xKEYID`
* You should also sign the new key with the old key and possibly your personal key and push it to some keyservers.
  * Use `gpg --yes -u 0xOLDKEYID --sign-key 0xKEYID` to sign.
* Now export the public key and inject it into the pubkeys slice in `verify.go`. Add a comment with the year and the key id.
* As usual send a PR and get this merged. Consider kicking off a new release, if that makes sense.
* STOP HERE. For a seamless key rollover we need to wait until most users had a chance to update to a version that has both the old and the new keys. So if possible wait a few months at least. Keep the GH issue open and assigned to track that process.
* After ~5 Months continue here.
* Regenerate the test signature for verify_test.go
  * Create a file that contains `gopass-sign-test\n` and run `gpg -u 0xKEYID --armor --output /tmp/testdata.sig --detatch-sign testdata`. Use the correct KEYID (the one of the NEW key).
  * Hint: Make sure the input only contains one line break, not two.
  * Paste the content of /tmp/testdata.sig into the `testSignature` in `verify_test.go`. Make sure all tests pass.
* Navigate to https://github.com/gopasspw/gopass/settings/secrets/actions
    * Paste the armored private part of the new key into the existing `GPG_PRIVATE_KEY` secret.
    * Paste the corresponding passphrase into `PASSPHRASE`.
* At this point you should be able to safely delete the old public key from verify.go and kick off a new release.
* At the very end upload the new key to some keyservers:  `gpg --send-keys 0xKEYID` and possibly `gpg --keyserver pgp.mit.edu --send-keys 0xKEYID`.
  * In case you mess up during key generation you might need to start over and you don't want to have conflicting keys on a keyserver where you can't delete them.
