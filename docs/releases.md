# Releases

Note: Only members who have at least `write` [access](https://github.com/gopasspw/gopass/settings/access) to the gopass repo can create releases.

Gopass uses [goreleaser](https://goreleaser.com/) to create releases. The configuration is in the file [`.goreleaser.yml`](../.goreleaser.yml).

Goreleaser automates most but not all steps of a new release.

Note: We use semantic versioning for the command line interface and tool behaviour
but not for the API (i.e. `pkg/gopass`). Maintaining both properties in the
same repository / Go module is too cumbersome.

## Development overview

Preparing and creating a new release requires a number of steps.
Starting right after the previous release these are roughly:

* Create a new Milestone in the [GitHub issue tracker](https://github.com/gopasspw/gopass/milestones)
* Create or assign issues for the next Milestone
  * We use [Slack](https://gopassworkspace.slack.com/) to discuss prioritization and responsibilities
* After enough changes have been accumulated on the master branch we might agree to cut a new release
* Now we survey open issues for any "blockers" that should make it into the next release
  * This usually either happens in Slack or on semi-regular video calls
* After all blockers have been addressed we move the remaining issues to the next milestone and prepare the release

## Cutting a release

This section is a reference for contributors with write access to the gopass
repository.

## Preparation

gopass release should work with the latest upstream version of goreleaser.

```bash
go get -u github.com/goreleaser/goreleaser
cd $GOPATH/src/github.com/goreleaser/goreleaser
go install
```

## Releasing a new release

This subsection applies to a new release in direct succession of the previous one, i.e. releasing what's in the master branch. If you need to cherry-pick
and base a release off of a previous one see the next subsection.

This is also the starting point for release candidates. The release helper accepts
any valid semantic version, including prerelease identifiers such as `-rc.1`.

We develop new features and fixes and feature branches which are frequently
merged into master in our own forks of the repository.

**Important: Do not push feature branches to the main repo.**

We have some helpers and automation in place to help us release new versions.
A new release can be prepared by anyone (doesn't need to be a maintainer).
Releasing it involves sending a PR that needs to be reviwed by the maintainers.
Once approved the PR will be merged and a tag needs to be pushed to trigger
the release process automation (using GitHub actions).

If you want to inspect the computed versions and changelog input before writing
files or creating a release branch, run the helper in dry-run mode first.

```bash
go run helpers/release/main.go --dry-run
go run helpers/release/main.go
# Follow the instructions to release a new minor version.
# If you want to skip a patch level or bump the minor version
# specify a version argument.
go run helpers/release/main.go v1.18.2
# If that confuses the changelog parser, you can specify a previous version
# as well.
go run helpers/release/main.go v1.18.2 v1.17.2
```

## Releasing a release candidate

Release candidates are regular semver prereleases. Use versions like
`v1.19.0-rc.1`, `v1.19.0-rc.2`, and so on.

The GitHub release will automatically be marked as a prerelease because
`.goreleaser.yml` uses `release.prerelease: auto`.

Use this workflow when you want to let users test a near-final build before the
stable tag is published:

1. Start from a clean, up-to-date `master` branch.
1. Pick the target stable version you are preparing, for example `v1.19.0`.
1. Prepare the first release candidate with an explicit prerelease version:

```bash
go run helpers/release/main.go --dry-run v1.19.0-rc.1
go run helpers/release/main.go v1.19.0-rc.1
```

1. Push the generated `release/v1.19.0-rc.1` branch and open a PR against `master`.
1. After the PR is reviewed and merged, create and push the signed tag:

```bash
git tag -s v1.19.0-rc.1
git push origin v1.19.0-rc.1
```

1. Ask testers to use prereleases explicitly, for example with `gopass update --pre`.

If you need another release candidate after additional fixes landed on `master`,
run the helper again with the next prerelease number:

```bash
go run helpers/release/main.go v1.19.0-rc.2
```

The helper will automatically use the latest existing `v1.19.0-rc.N` tag as the
previous version for changelog generation. You can still pass the previous version
explicitly if you need to override that behavior:

```bash
go run helpers/release/main.go v1.19.0-rc.2 v1.19.0-rc.1
```

Once the final release is ready, run the helper without a prerelease suffix. If
`VERSION` still contains the release candidate version, the helper will promote it
to the final stable version automatically:

```bash
go run helpers/release/main.go
# or explicitly
go run helpers/release/main.go v1.19.0
```

Notes:

* Do not set `PATCH_RELEASE=true` for normal release candidates that should be cut from `master`. That mode is only for staying on the current branch, e.g. for cherry-pick releases.
* The changelog for the final stable release will still cover the full delta since the last stable tag, even if one or more release candidates were published before.
* Maintainers should create RC tags from the merged `release/vX.Y.Z-rc.N` branch and push them as signed tags named exactly `vX.Y.Z-rc.N` so goreleaser publishes them as prereleases.

After the helper ran it will show you instructions how to push your release
branch and create a PR for review. Once it's merged a maintainer only needs
to tag it and push the tag to trigger the release process automation.

Afterwards a maintainer should run the post-release automation that will
perform some cleanup, create new GitHub milestones and send out PRs to
rolling release distributions.

### Verifying published assets

Release automation publishes `SHA256SUMS` together with a keyless cosign bundle
named `SHA256SUMS.sigstore.json`. After downloading both assets from a GitHub
release, verify them with:

```bash
cosign verify-blob \
  --bundle SHA256SUMS.sigstore.json \
  SHA256SUMS
```

Once the checksum file is verified, use it to validate the archive or package
you downloaded with your usual checksum tooling.

## Releasing a cherry-pick release

This subsection applies to a new release that should be based on a previous
release that is not a direct ancestor of the master branch, i.e. because
breaking changes were introduced or other releases have happend in between.

This can still use our release automation but it will require some adjustments:

* Check out the previous release tag (we usually only publish a release branch if we need to): `git checkout v1.12.2`
* Create a new release preparation branch (the release automation will create the actual release branch later): `git checkout -b prep/v.1.12.3`
* Cherry-pick the changes from the previous release into the release preparation branch: `git cherry-pick -x HASH1 HASH2 ...`
  * Resolve any conflicts, make sure all tests pass and `git cherry-pick --continue`
* Trigger the release preparation, it should pick up any changelog entries from the cherry-picked commit messages: `PATCH_RELEASE=true go run helpers/release/main.go`
  * `PATCH_RELEASE=true` instructs it to not change the current branch to master
* Push the release branch printed at the end to the repository (or your fork) and open a PR.
  * IMPORANT: This PR will not be merged into master! We will just use it to create a tag and trigger the release automation.

## Reproducible Builds

`gopass` supports [reproducible builds](https://reproducible-builds.org/). When
building from git [`SOURCE_DATE_EPOCH`](https://reproducible-builds.org/docs/source-date-epoch/)
can be used to override the compile date, .e.g `SOURCE_DATE_EPOCH=$(git log -1 --pretty=%ct)`.
When building a release `goreleaser` will automatically use the exact timestamp
of the last commit.

Internal paths are stripped using `-trimpath` and appropriate `-ldflags` (e.g.
`-s`, `-w`). See the Makefile header for the exact set of flags.
