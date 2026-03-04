## Releases

Note: Only members who have at least `write` [access](https://github.com/gopasspw/gopass/settings/access) to the gopass repo can create releases.

Gopass uses [goreleaser](https://goreleaser.com/) to create releases. The configuration is in the file [`.goreleaser.yml`](../.goreleaser.yml).

Goreleaser automates most but not all steps of a new release.

Note: We use semantic versioning for the command line interface and tool behaviour
but not for the API (i.e. `pkg/gopass`). Maintaining both properties in the
same repository / Go module is too cumbersome.

### Development overview

Preparing and creating a new release requires a number of steps.
Starting right after the previous release these are roughly:

* Create a new Milestone in the [GitHub issue tracker](https://github.com/gopasspw/gopass/milestones)
* Create or assign issues for the next Milestone
  * We use [Slack](https://gopassworkspace.slack.com/) to discuss prioritization and responsibilities
* After enough changes have been accumulated on the master branch we might agree to cut a new release
* Now we survey open issues for any "blockers" that should make it into the next release
  * This usually either happens in Slack or on semi-regular video calls
* After all blockers have been addressed we move the remaining issues to the next milestone and prepare the release

### Cutting a release

This section is a reference for contributors with write access to the gopass
repository.

### Preparation

gopass release should work with the latest upstream version of goreleaser.

```bash
go get -u github.com/goreleaser/goreleaser
cd $GOPATH/src/github.com/goreleaser/goreleaser
go install
```

### Releasing a new release

This subsection applies to a new release in direct succession of the previous one, i.e. releasing what's in the master branch. If you need to cherry-pick
and base a release off of a previous one see the next subsection.

We develop new features and fixes and feature branches which are frequently
merged into master in our own forks of the repository.

**Important: Do not push feature branches to the main repo.**

We have some helpers and automation in place to help us release new versions.
A new release can be prepared by anyone (doesn't need to be a maintainer).
Releasing it involves sending a PR that needs to be reviwed by the maintainers.
Once approved the PR will be merged and a tag needs to be pushed to trigger
the release process automation (using GitHub actions).

```bash
$ go run helpers/release/main.go
# Follow the instructions to release a new minor version.
# If you want to skip a patch level or bump the minor version
# specify a version argument.
$ go run helpers/release/main.go v1.18.2
# If that confuses the changelog parser, you can specify a previous version
# as well.
$ go run helpers/release/main.go v1.18.2 v1.17.2
```

After the helper ran it will show you instructions how to push your release
branch and create a PR for review. Once it's merged a maintainer only needs
to tag it and push the tag to trigger the release process automation.

Afterwards a maintainer should run the post-release automation that will
perform some cleanup, create new GitHub milestones and send out PRs to
rolling release distributions.

### Releasing a cherry-pick release

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

### Reproducible Builds

`gopass` supports [reproducible builds](https://reproducible-builds.org/). When
building from git [`SOURCE_DATE_EPOCH`](https://reproducible-builds.org/docs/source-date-epoch/)
can be used to override the compile date, .e.g `SOURCE_DATE_EPOCH=$(git log -1 --pretty=%ct)`.
When building a release `goreleaser` will automatically use the exact timestamp
of the last commit.

Internal paths are stripped using `-trimpath` and appropriate `-ldflags` (e.g. 
`-s`, `-w`). See the Makefile header for the exact set of flags.

