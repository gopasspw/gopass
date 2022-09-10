## 1.14.6 / 2022-09-10

* [BUGFIX] Do not show setup message on version (#2327)
* [BUGFIX] Remove exported public keys of removed (#2328, #2315)
* [ENHANCEMENT] Document extension model. (#2329, #2290)

## 1.14.5 / 2022-09-03

* [BUGFIX] Fix fsck progress bar. Mostly. (#2303)
* [DOCUMENTATION] fix in recommended vim setting (#2318)

## 1.14.4 / 2022-08-02

* [BREAKING] gopass otp will automatically update the counter key in HTOP secrets! (#2278)
* [BUGFIX] Allow removing unknown recipients with --force (#2253)
* [BUGFIX] Honor PASSWORD_STORE_DIR (#2272)
* [BUGFIX] Honor OTP key period from URL (#2278)
* [BUGFIX] Wizard: Enforce min and max length. (#2293)
* [CLEANUP] Use Go 1.19 (#2296)
* [ENHANCEMENT] Automatically sync once a week (#2191)
* [ENHANCEMENT] Scan for vulnerabilities and add SBOM on (#2268)
* [ENHANCEMENT] Use packages.gopass.pw for APT packages (#2261)

## 1.14.3 / 2022-05-31

* [BUGFIX] Do not print progress bar on otp --clip (#2243)
* [BUGFIX] Removing shadowing warning when using -o/--password (#2245)
* [CLEANUP] Deprecate OutputIsRedirected in favour of IsTerminal (#2248)
* [DOCUMENTATION] Adding doc about YAML entries and unsafe-keys (#2244)
* [ENHANCEMENT] Allow deleting multiple secrets (#2239)

## 1.14.2 / 2022-05-22

* [BUGFIX] Fix gpg identity detection (#2218, #2179)
* [BUGFIX] Handle different line breaks in recipient (#2221, #2220)
* [BUGFIX] Stop eating secrets on move (#2211, #2210)
* [ENHANCEMENT] Add flag to keep env variable capitalization (#2226, #2225)
* [ENHANCEMENT] Environment variable GOPASS_PW_DEFAULT_LENGTH can be used to overwrite default password length of 24 characters. (#2219)

## 1.14.1 / 2022-05-02

* [BUGFIX] Do not print missing public key for age. (#2166)
* [BUGFIX] Improve convert output (#2171)
* [BUGFIX] fix errors in zsh completions (#2005)
* [CLEANUP] Migrating to a maintained version of openpgp (#2193)
* [ENHANCEMENT] Avoid decryption on move or copy (#2183, #2181)
* [UX] Upgrade xkcdpwgen to a new version that removes German (#2187)

## 1.14.0 / 2022-03-16

* Add --chars option to print subset of secrets (#2155, #2068)
* [BUGFIX] Always re-encrypt when fsck is invoked with --decrypt. (#2119, #2015)
* [BUGFIX] Body only entries are detected now by show -o (#2109)
* [BUGFIX] Do not hide git error messages (#2118, #1959)
* [BUGFIX] Fix completion when password name contains (#2150)
* [BUGFIX] Fix template func arg order (#2117, #2116)
* [BUGFIX] Fixes an issue where recipients remove may fail (#2147, #1964)
* [BUGFIX] Fixes an issue where recipients remove may fail (#2147, #1964)
* [BUGFIX] Handle from prefix correctly on mv (#2110, #2079)
* [BUGFIX] Handle unencoded secret on cat (#2105)
* [BUGFIX] Make man page consistent with other docs (#2133)
* [BUGFIX] Reject invalid salt with MD5Crypt templates (#2128)
* [BUGFIX] depend *.deb on gnupg instead of dummy (#2050)
* [CLEANUP] Deprecate gopasspw/pinentry (#2095)
* [CLEANUP] Use Go 1.18 (#2156)
* [CLEANUP] Use debug.ReadBuildInfo (#2032)
* [DOCUMENTATION] Fixed link to passwordstore.org (#2129)
* [DOCUMENTATION] document 'gopass cat' (#2051)
* [DOCUMENTATION] improve 'gopass cat' (#2070)
* [DOCUMENTATION] improve 'gopass show -revision -<N>' (#2070)
* [ENHANCEMENT] Add age subcommand (#2103, #2098)
* [ENHANCEMENT] Add gopass audit --expiry (#2067)
* [ENHANCEMENT] Add gopass process (#2066, #1913)
* [ENHANCEMENT] Allow overriding GPG path (#2153)
* [ENHANCEMENT] Automatically export creators key to the (#2159, #1919)
* [ENHANCEMENT] Bump to Go 1.18 (#2058)
* [ENHANCEMENT] Enforce TLSv1.3 (#2085)
* [ENHANCEMENT] Generics (#2034, #2030)
* [ENHANCEMENT] Hide password on MacOS clipboards (#2065)
* [ENHANCEMENT] Passage compat improvements (#2060, #2060)
* [ENHANCEMENT] gopass git invokes git directly (#2102)
* [ENHANCEMENT] Template support for the create wizard (#2064)
* [ENHANCENMENT] Check for MacOS Keychain storing the GPG (#2144)
* [EXPERIMENTAL] Support the Fossil SCM (#2092, #2022)
* [FEATURE] Add env variables for custom clipboard commands. (#2091, #2042)
* [FEATURE] only accept keys with "encryption" key capability (#2047, #1917, #1917)
* [TESTING] Improve two line test ambiguity. (#2091, #2042)
* [TESTING] Use a helper to unset env vars in clipboard tests. (#2091, #2042)
* [UX] OTP code now runs in loop until canceled or used with -o (#2041)

## 1.13.1 / 2022-01-15

* [BUGFIX] Handle from prefix correctly on mv (#2110, #2079)
* [BUGFIX] Handle unencoded secret on cat

## 1.13.0 / 2021-11-13

* [BUGFIX] Do not print OTP progress bar if not in terminal (#2019)
* [BUGFIX] Don't prompt to retype password unnecessarily (#1983)
* [BUGFIX] Fix AutoClip handling on generate (#2024, #2023)
* [BUGFIX] Replace Build Status badge in README (#2016)
* [BUGFIX] The field 'parsing' is now honored with legacy config pre v1.12.7 (#1997)
* [BUGFIX] Use default git branch on setup (#2026, #1945)
* [ENHANCEMENT] Adding a MSI installer for Windows (#2001)
* [ENHANCEMENT] Move password prompts to stderr (#2004)
* [FEATURE] Add capitalized words to memorable passwords (#1985, #1984)
* [UX] Use new progress bar for OTP expiry time (#2019)

## 1.12.8 / 2021-08-28

* [BUGFIX] Use same default for partial config files (#1968)
* [CLEANUP] Remove GOPASS_NOCOLOR in favor of NO_COLOR (#1937, #1936)
* [ENHACNEMENT] Add gopass merge (#1979, #1948)
* [ENHANCEMENT] Add --symbols to gopass pwgen (#1966)
* [ENHANCEMENT] Warn on untracked files (#1972)

## 1.12.7 / 2021-07-02

* DOCUMENTATION Fixed Single Line Formating for Clone Documentation (#1943)
* [BUGFIX] Allow --strict to be chained with --symbols (#1952, #1941)
* [BUGFIX] Normalize recipient IDs before comparison (#1953, #1900)
* [BUGFIX] Use /tmp for GIT_SSH_COMMAND on Mac (#1951, #1896)
* [ENHANCEMENT] Add warning when parsing content (#1950)

## 1.12.6 / 2021-05-01

* [BUGFIX] Do not recurse with a key (#1907, #1906)
* [BUGFIX] Fix SSH control path (#1899, #1896)
* [BUGFIX] Fix gopass env with subtrees (#1894, #1893)
* [BUGFIX] Honor create -s flag (#1891)
* [BUGFIX] Ignore commented values in gpg config (#1901, #1898)
* [ENHANCEMENT] Add better usage instructions (#1912)

## 1.12.5 / 2021-03-27

* [BUGFIX] Allow subkeys (#1843, #1841, #1842)
* [BUGFIX] Avoid logging credentials (#1886, #1883)
* [BUGFIX] Fix SSH Command override on termux (#1881)
* [CLEANUP] Moving pkg/pinentry to gopasspw/pinentry (#1876)
* [ENHANCEMENT] Add -f flag to create (#1867, #1811)
* [ENHANCEMENT] Add gopass ln (#1828)
* [ENHANCEMENT] Add proper diff numbers on sync (#1882)
* [ENHANCEMENT] Update password rules (#1861)

## 1.12.4 / 2021-03-20

* [BUGFIX] Bring back --yes (#1862, #1858)
* [BUGFIX] Fix make install on BSD (#1859)

## 1.12.3 / 2021-03-20

* [BUGFIX] Fix generate -c (#1846, #1844)
* [BUGFIX] Fix gopass update (#1838, #1837)
* [BUGFIX] Fix progress bar on 32 bit archs (#1855, #1854)
* [CLEANUP] Remove the custom formula in favour of the official one. (#1847)
* [ENHANCEMENT] Install manpage when using `make install` (#1845)

## 1.12.2 / 2021-03-13

* [BUGFIX] Do not fail if reminder is unavailable (#1835, #1832)
* [BUGFIX] Do not shadow directories (#1817, #1813)
* [BUGFIX] Do not trigger ClamAV FP (#1810, #1807)
* [BUGFIX] Fix -o (#1822)
* [BUGFIX] Honor Ctrl+C while waiting for user input (#1805, #1800)
* [ENHANCEMENT] Add gopass.1 man page (#1827, #1824)
* [UX] Adding the grep command to --help (#1826, #1825)

## 1.12.1 / 2021-02-17

* [BUGFIX] Enable updater on Windows (#1790, #1789)
* [BUGFIX] Fix progress bar nil pointer access (#1790, #1789)
* [BUGFIX] Fix % char in passwords being treated as formatting (#1794, #1793, #1801)
* [ENHANCEMENT] Add ARCHITECTURE.md (#1787)
* [ENHANCEMENT] Added a env var to disable reminders (#1792)
* [ENHANCEMENT] Remind to run gopass update/fsck/audit after 90d (#1792)

## 1.12.0 / 2021-02-11

WARNING: The self updater does not support updating from 1.11.0 to 1.12.0. Our
release infrastructure does not support the key type used in 1.11.0.

NOTE: This release drops the integrations that were moved to their own repos,
i.e. `git-credential-gopass`, `gopass-hibp`, `gopass-jsonapi` and
`gopass-summon-provider`.

We have implemented proper release signing and verification for the self
updater and brought it back.

* [BUGFIX] Add signature verification for updater (#1717, #1676)
* [BUGFIX] Allow using tilde (#1713, #872)
* [BUGFIX] Always allow removing mounts (#1748, #1746)
* [BUGFIX] Ask passphrase upon key generation (#1715, #1698)
* [BUGFIX] Do not overwrite age keyring (#1734, #1678)
* [BUGFIX] Remove empty parents on gopass rm -r (#1725, #1723)
* [BUGFIX] The empty password must now be confirmed too (#1719)
* [BUGFIX] Use the first GPG found in path on Windows (#1751, #1635)
* [BUGFIX] Warn about --throw-keyids (#1759, #1756)
* [BUGFIX] fixed mixed case keys for key-value, all keys are lower case now (#1778)
* [CLEANUP] Remove migrated binaries (#1712, #1673, #1649, #1652, #1631, #1165, #1711, #1670, #1639)
* [CLEANUP] Remove the ondisk backend (#1720)
* [ENHANCEMENT] Add -A and -B to pwgen (#1716)
* [ENHANCEMENT] Add Pinentry CLI fallback (#1697, #1655)
* [ENHANCEMENT] Add REPL cmd lock (#1744)
* [ENHANCEMENT] Add optional pinentry unescaping (#1621)
* [ENHANCEMENT] Add tpl funcs for Bcrypt and Argon2 (#1706, #1689)
* [ENHANCEMENT] Add windows support to the self updater (#1724, #1722)
* [ENHANCEMENT] Confirm new age keyring passphrases (#1747)
* [ENHANCEMENT] KV secrets are now key-values, supporting multiple same key with different values (#1741)
* [ENHANCEMENT] UTF-8 emojis (#1715, #1698)
* [ENHANCEMENT] Use gpgconf to the the gpg binary (#1758, #1757)
* [ENHANCEMENT] Use main as the git default branch (#1749, #1742)
* [ENHANCEMENT] Use persistent SSH connections (#1755)
* [TESTING] Adding DI to Github Actions (#1728)

## 1.11.0 / 2020-01-12

This is an important bugfix release that should resolve several outstanding
issues and concerns. Since 1.10.0 was released was engaged in a lot of
discussions and realized that compatibility is more important than we first
thought. So we're rolling back some breaking changes and revise some parts of
our roadmap. We will strive to remain compatible with other password store
implementations - but remember this is a goal, not a promise. This means we'll
continue using compatible secrets formats as well as GPG and Git.

* [BUGFIX] Allow secret names to have a colon in the name
* [BUGFIX] Apply limit in list correctly
* [BUGFIX] Correcting newlines handling
* [BUGFIX] Correct missing padding to TOTP entry
* [BUGFIX] Create cache folder if doesn't exist. Relevant
* [BUGFIX] Disable gopass update
* [BUGFIX] Disabling all kind of parsing of the input
* [BUGFIX] Do not duplicate key password in K/V secrets
* [BUGFIX] Do not search for new secrets
* [BUGFIX] fixes gopass-jsonapi for MacTools GPGSuite users.
* [BUGFIX] Fix legacy config parsing
* [BUGFIX] fsck won't correct recipients without --decrypt
* [BUGFIX] Insert is not resetting the pw now if a key:value pair is specified inline
* [BUGFIX] Insert is now parsing its stdin input
* [BUGFIX] Invalidate GPG key list after generation
* [BUGFIX] List no longer uses the store size as its default depth
* [BUGFIX] Nil dereference in cui
* [BUGFIX] Pass arguments to a notification program
* [BUGFIX] Password insert prompt now works on Windows but
* [BUGFIX] Re-adding the global --yes flag
* [BUGFIX] Remove GPG location caching
* [BUGFIX] Restore path-removal from old config-format
* [BUGFIX] Show now correctly handles -C and -u together
* [BUGFIX] The deprecation warning is now output on stderr
* [BUGFIX] Trim version prefix in jsonapi
* [CLEANUP] Remove MIME
* [CLEANUP] Remove the unfinished xc backend
* [CLEANUP] Update to minio/v7
* [DOCUMENTATION] Edited features.md
* [DOCUMENTATION] Improve contributing guide.
* [DOCUMENTATION] Slight updates to reflect the recent code
* [ENHANCEMENT] Adding a trailing separator to the listed folders
* [ENHANCEMENT] Adding the flag show -n to disable output parsing
* [ENHANCEMENT] Adding the option parsing to disable all parsing
* [ENHANCEMENT] fsck now detects leftover Mime secrets
* [ENHANCEMENT] Full windows support
* [ENHANCEMENT] Prompt for edit search result
* [ENHANCEMENT] Re-introduce gopass -c
* [ENHANCEMENT] Show GPG --gen-key error to the user
* [ENHANCEMENT] This is required when using e.g. Gnome Keyring.
* [ENHANCEMENT] Use 32 byte salt by default
* [UX] Preserve content across retries

## 1.10.1

* [BUGFIX] Fix the Makefile
* [BUGFIX] Remove misleading config error message
* [BUGFIX] Re-use existing root store
* [BUGFIX] Use standard Unix directories on MacOS

## 1.10.0

WARNING: This release contains a few breaking changes as well as necessary
packaging changes.

This release is building the foundation for an eventual 2.0 release
which will drop many legacy features and significantly shrink the
codebase to ensure long term maintainability. The goal is to remove
the support for multiple backends and any external dependencies,
including `git` and `gpg` binaries. By default the tool should be easy to use,
secure and modern. We will still support our flagship use cases,
like working in teams. Also gopass might eventually move to an
fully encrypted backend where we don't leak information through
filenames.

Any gopass 1.x release should still be compatible with any
password store implementation (possibly with some caveats).
Beyond that we plan to drop any compatibility goals.

If you are using different Password Store implementations to access your
secrets, e.g. on mobile devices, you might want to run `gopass config mime false`
before performing any kind of write operation on the password store. Otherwise
mutated secrets will be written using the new native gopass MIME format and
might not be readable from other implementations.

This release adds documentation for all supported subcommands in the `docs/commands`
folder and starts define our core use cases in the `docs/usecases` folder.
Please note that the command documentation also serves as a specification on
how these commands are supposed to operate.

Note: We have accumulated too many changes so we've decided to skip the 1.9.3
release and issue the first release of the 1.10. series.

Note to package maintainers: This release adds additional binaries which should
be included in any binary re-distribution of gopass.

* [BREAKING] New secrets format
* [BUGFIX] Allow deleting shadowed secret
* [BUGFIX] Correctly handle exportkeys and auto import for noop
* [BUGFIX] Do not allow malformed secrets
* [BUGFIX] Do not return error on no grep matches
* [BUGFIX] Fix config panic with mounts
* [BUGFIX] Fix fsck progress bar.
* [BUGFIX] Fix git init
* [BUGFIX] Fix optional key passed through find
* [BUGFIX] Fix tree shadowing.
* [BUGFIX] Handle relative path during init
* [BUGFIX] Honor generate --print
* [BUGFIX] Honor trust level during onboarding.
* [BUGFIX] Print RCS error message
* [BUGFIX] Print config parse error to STDERR
* [BUGFIX] Properly initialize crypto during onboarding and
* [BUGFIX] env command: do not crash if called without a command to execute
* [CLEANUP] Merge Storage and RCS backends
* [CLEANUP] Move internal packages to internal
* [CLEANUP] Remove autoclip for gopass show
* [CLEANUP] Remove config option confirm
* [CLEANUP] Remove curses UI
* [CLEANUP] Remove the --sync flag to gopass show
* [CLEANUP] Rename --force to --unsafe for show
* [CLEANUP] Rename xkcd generator options
* [DEPRECATION] Mark gopass git as deprecated
* [DEPRECATION] Remove AutoPrint
* [DEPRECATION] Remove askformore, autosync
* [DEPRECATION] Retire editrecipients option
* [DOCUMENTATION] Document audit, generate, insert and show
* [DOCUMENTATION] Document list flags
* [DOCUMENTATION] Improve documentation of Zsh completion setup
* [ENHANCEMENT] Add GOPASS_DISABLE_MIME to disable new
* [ENHANCEMENT] Add arm and arm64 binaries
* [ENHANCEMENT] Add gopass API (unstable)
* [ENHANCEMENT] Add regexp support to gopass grep
* [ENHANCEMENT] Add zxcvbn password strength checker
* [ENHANCEMENT] Avoid direct show on gopass search
* [ENHANCEMENT] Cache gpg binary location
* [ENHANCEMENT] Ignore binary secrets for audit
* [ENHANCEMENT] Introduce --generator flag
* [ENHANCEMENT] Introduce unsafe-keys
* [ENHANCEMENT] Make audit report passwords not changed
* [ENHANCEMENT] Make show --qr flag complementary
* [ENHANCEMENT] New Debug package
* [ENHANCEMENT] New progress bar
* [ENHANCEMENT] Print password before sync
* [ENHANCEMENT] Provide more helpful config parse errors
* [ENHANCEMENT] Rewrite tree implementation
* [ENHANCEMENT] Show recipients from subfolder id files
* [ENHANCEMENT] Speed up gpg store init
* [ENHANCEMENT] Support changing path with gopass config
* [ENHANCEMENT] Support relative revisions for show
* [ENHANCEMENT] Warn if vim might be leaking secrets
* [ENHANCEMENT] env command: more tests
* [FEATURE] Add Password Rules and Domain Alias support
* [FEATURE] Add experimental backend converter
* [FEATURE] Add remote config for ondisk storage
* [FEATURE] Add remote sync support for the ondisk backend
* [FEATURE] Add summon provider
* [FEATURE] Pinentry API: support OPTION API call
* [FEATURE] REPL
* [TESTING] Add a test to detect shadowing issue with mount

## 1.9.2 / 2020-05-13

* [BUGFIX] Bring back the custom fish completion.
* [BUGFIX] Disable AutoClip when redirecting stdout
* [ENHANCEMENT] Create new sub stores in XDG compliant locations.

## 1.9.1 / 2020-05-09

* [BUGFIX] Do not copy to clipboard with -f
* [BUGFIX] Encrypt parent directory if leaf node exists.
* [BUGFIX] Fix -c and -C for default show action.
* [BUGFIX] Hide git-credential store warning.
* [BUGFIX] Honor notifications setting.
* [BUGFIX] Simplify autoclip behavior
* [DEPRECATION] Remove PASSWORD_STORE_DIR support
* [ENHANCEMENT] Add exportkeys option.
* [ENHANCEMENT] Add memorable password generator
* [ENHANCEMENT] Add preliminary age encryption support.

## 1.9.0 / 2020-05-01

* [ENHANCEMENT] Proper windows support [#1295]
* [ENHANCEMENT] Add pwgen subcommand [#1308]
* [ENHANCEMENT] Only decrypt when needed [#1289]
* [ENHANCEMENT] Full unattended password generation [#1259]
* [ENHANCEMENT] Add -C flag [#1272]
* [ENHANCEMENT] Migrate to urface/cli/v2 [#1276]
* [ENHANCEMENT] Support Termux [#913]
* [BUGFIX] Do not fail if nothing to commit [#1168, #1103]
* [BUGFIX] Restore PASSWORD_STORE_DIR support [#1213]
* [BUGFIX] Do not remove empty second line [#1235]
* [BUGFIX] Do not disable color if no PAGER is available [#1244]
* [BUGFIX] Do not overwrite entry when reading from STDIN [#1245]
* [BUGFIX] Commit when using concurrency gt 1 [#1246]
* [BUGFIX] Do not error out when listing a leaf node [#1300]
* [BUGFIX] Do not overwrite config if PASSWORD_STORE_DIR is set [#1286]
* [BUGFIX] Fix go get support [#1288]
* [DEPRECATION] Remove Dockerfile [#1309]
* [DEPRECATION] Remove Bintray [#1304]
* [DEPRECATION] Deprecate OTP, Binary, YAML git-credentials and xc support [#1301]
* [DEPRECATION] Remove support for OpenPGP (library), GoGit, Vault, Consul and encrypted configs [#1290, #1283, #1282, #1279]

## 1.8.6 / 2019-07-26

* [ENHANCEMENT] Add --password to otp command [#1150]
* [ENHANCEMENT] Support adding key values with colons [#1128]
* [BUGFIX] Allow overwriting directories with --force [#1149]
* [BUGFIX] Sort list of stores when adding recipients [#1144]
* [BUGFIX] Sort recipients by Name not by ID [#1143]
* [BUGFIX] Handle slashes in recipient names [#1139]

## 1.8.5 / 2019-03-03

* [ENHANCEMENT] Improve template handling [#1029]
* [ENHANCEMENT] Remove empty directories [#1009]
* [ENHANCEMENT] Improve performance of unclip [#923]
* [ENHANCEMENT] Add AutoPrint option [#1065]
* [ENHANCEMENT] Follow the rsync convention for cp/mv commands [#1055]
* [BUGFIX] Fix bash completion for MSYS on Windows [#1053]
* [BUGFIX] Git clone failing [#1036]

## 1.8.4 / 2018-12-26

* [ENHANCEMENT] Evaluate templates when inserting single secrets [#1023]
* [ENHANCEMENT] Add fuzzy search dialog for gopass otp [#1021]
* [ENHANCEMENT] Add edit option to search dialog [#1019]
* [ENHANCEMENT] Introduce build tags for experimental features [#1000]
* [BUGFIX] Fix recursive delete [#1024]
* [BUGFIX] Abort tests on critical failures [#997]
* [BUGFIX] Zsh autocompletion [#996]

## 1.8.3 / 2018-11-19

* [ENHANCEMENT] Add zsh autocompletion for insert and generate [#988]
* [ENHANCEMENT] Set exit code for filtered ls without result [#983]
* [ENHANCEMENT] Improve generate command [#948]
* [ENHANCEMENT] Print summary for grep [#943]
* [ENHANCEMENT] Documentation updates [#924, #890, #918, #919, #920, #944, #952, #958, #969, #985]
* [ENHANCEMENT] jsonapi: Add windows support for configure [#904]
* [ENHANCEMENT] jsonapi: Add getVersion [#893]
* [ENHANCEMENT] Support symlinks for fs storage backend [#886]
* [BUGFIX] Offer store selection with exactly one mount point as well [#987]
* [BUGFIX] Edit entry selected by fuzzy search [#979]
* [BUGFIX] Fix path handling on windows [#970]
* [BUGFIX] Remove quotes [#967]
* [BUGFIX] Properly handle git add for removed files [#946]
* [BUGFIX] HAndle already mounted and not initialized errors [#945]
* [BUGFIX] Fix HIBP command options [#936]
* [BUGFIX] Offer secret selection on edit command [#929]
* [BUGFIX] jsonapi: add initialize [#903]
* [BUGFIX] Update external dependencies [#884, #932, #981]
* [BUGFIX] Use valid crypto backend for key selection [#889]

## 1.8.2 / 2018-06-28

* [ENHANCEMENT] Improve fsck output [#859]
* [ENHANCEMENT] Enable notifications on FreeBSD [#863]
* [ENHANCEMENT] Redirect errors to stderr [#880]
* [ENHANCEMENT] Do not writer version to config [#883]
* [BUGFIX] Fix commit on move [#860]
* [BUGFIX] Properly check store initialization [#865]

## 1.8.1 / 2018-06-08

* [BUGFIX] Trim fsck path [#856]
* [BUGFIX] Handle URL parse errors in create [#855]

## 1.8.0 / 2018-06-06

This release includes several possibly breaking changes.
The `gopass move` implementation was refactored to properly support moving
entries and subtrees across mount points. This may change the behaviour slightly.
Also the build flags were changed to build PIE binaries. This should not affect
the runtime behaviour, but we could not test this on all platforms, yet.

* [BREAKING] Make move work recursively and across stores [#821]
* [FEATURE] Add git credential caching [#743]
* [FEATURE] Add local recipient integrity checks [#800 #826]
* [ENHANCEMENT] Handle key-value pairs on generate and insert [#790]
* [ENHANCEMENT] Add gpg.listKeys caching [#804]
* [ENHANCEMENT] Add append mode for gopass insert [#807]
* [ENHANCEMENT] Support external password generators [#811]
* [ENHANCEMENT] Add gopass generate completion heuristic [#817]
* [ENHANCEMENT] Add revive linter checks [#822]
* [ENHANCEMENT] Remove -static build flag, enable CGO and -buildmode=PIE [#823]
* [ENHANCEMENT] Warn if RCS backend is noop during gopass sync [#825]
* [ENHANCEMENT] Support for special password rules on generate [#832]
* [ENHANCEMENT] Improve create wizard [#842]
* [ENHANCEMENT] Honor templates on generate [#847]
* [ENHANCEMENT] Support NO_COLOR [#851]
* [BUGFIX] Reset clipboard timer on repeated copy [#813]
* [BUGFIX] Add --force to git add invocation [#839]
* [BUGFIX] Rename updater GitHub Organisation [#818]
* [BUGFIX] Default to origin master for git pull [#819]
* [BUGFIX] Properly propagate RCS backend on gopass clone [#820]
* [BUGFIX] Fix sub store config propagation [#837 #841]
* [BUGFIX] Use default for password store dir [#846]
* [BUGFIX] Properly handle autosync on recipients save [#848]
* [BUGFIX] Resolve key IDs to fingerprints before adding or removing [#850]

## 1.7.2 / 2018-05-28

* [BUGFIX] Fix tilde expansion [#802]

## 1.7.1 / 2018-05-25

* [BUGFIX] Add nogit compat handler [#792]
* [BUGFIX] Fix reencrypt [#796]

## 1.7.0 / 2018-05-22

* [FEATURE] Pluggable crypto, storage and RCS backends. Including a pure-Go NaCl based crypto backend [#645] [#680] [#736] [#777]
* [FEATURE] Password history [#660]
* [FEATURE] Vault backend [#723] [#730]
* [FEATURE] Consul backend [#697]
* [FEATURE] HIBPv2 Dump and API support [#666] [#706]
* [FEATURE] Select recipients per secret [#703]
* [FEATURE] Add experimental OpenPGP crypto backend [#670]
* [ENHANCEMENT] Support HIBPv2 API and Dumps [#666]
* [ENHANCEMENT] Robust K/V parser with YAML fallback [#659]
* [ENHANCEMENT] Restrict fsck to given path [#721]
* [ENHANCEMENT] Refactor [#702] [#708] [#715] [#722] [#731]
* [ENHANCEMENT] Proper Makefile dependencies [#707]
* [ENHANCEMENT] Auto-copy with safecontent [#685]
* [ENHANCEMENT] Add disable notifications option [#690]
* [ENHANCEMENT] Migrate from govendor to dep [#688]
* [ENHANCEMENT] Improve test coverage [#732] [#781] [#782]
* [ENHANCEMENT] Improvate YAML handling [#739]
* [ENHANCEMENT] Audit freshly generated passwords [#761]
* [BUGFIX] Use sh instead of bash [#699]
* [BUGFIX] Lookup correct remote for current branch [#692]
* [BUGFIX] Fix GPG binary detection on Windows [#681] [#693]
* [BUGFIX] Version [#727]
* [BUGFIX] Git init [#729]
* [BUGFIX] Secret.String() [#738]
* [BUGFIX] Fix generate --symbols [#742] [#783]

## 1.6.11 / 2018-02-20

* [ENHANCEMENT] Documentation updates [#648] [#656]
* [ENHANCEMENT] Add secret completions to edit command in zsh [#654]
* [BUGFIX] Avoid escaping values added to secrets [#658]
* [BUGFIX] Fix parsing of GPG UIDs [#650]

## 1.6.10 / 2018-01-18

* [ENHANCEMENT] Add Travis MacOS builds [#618]
* [ENHANCEMENT] Make gopass build on DragonFlyBSD [#619]
* [ENHANCEMENT] Increase test coverage [#621] [#622] [#624]
* [BUGFIX] Properly handle sub-store configuration [#625]
* [BUGFIX] Fix Makefile [#615] [#617]
* [BUGFIX] Fix failing tests on MacOS [#614]

## 1.6.9 / 2018-01-05

* [BUGFIX] Fix update URL check [#610]

## 1.6.8 / 2018-01-05

* [ENHANCEMENT] Add OpenBSD Ksh completion [#586]
* [ENHANCEMENT] Increase test coverage [#589] [#590] [#592] [#595] [#596] [#597] [#601] [#602] [#603] [#604]
* [ENHANCEMENT] Update Documentation and Dockerfile [#591] [#605]
* [BUGFIX] Use Termwiz CUI on OpenBSD [#588]
* [BUGFIX] Fix create wizard [#594]
* [BUGFIX] Use persistent bufio.Reader [#607]

## 1.6.7 / 2017-12-31

* [ENHANCEMENT] Add --sync flag to gopass show [#544]
* [ENHANCEMENT] Update dependencies [#547]
* [ENHANCEMENT] Use gocui for terminal UI [#562]
* [ENHANCEMENT] Increase test coverage [#548] [#549] [#567] [#568] [#570] [#572] [#574] [#575] [#577] [#578] [#583] [#584]
* [ENHANCEMENT] Add Dockerfile [#561]
* [ENHANCEMENT] Add zsh and fish completion generator [#565]
* [ENHANCEMENT] Add go-fuzz instrumentation [#576]
* [BUGFIX] Catch URL parse errors [#546]

## 1.6.6 / 2017-12-20

* [FEATURE] Selective Sync [#538]
* [ENHANCEMENT] Make termwiz honor copy flag [#534]
* [ENHANCEMENT] Make shell completion respect binary name [#536]
* [ENHANCEMENT] Refactor [#533] [#540] [#541] [#542]
* [BUGFIX] Show git output [#529]

## 1.6.5 / 2017-12-15

* [ENHANCEMENT] Handle errors gracefully [#524]
* [BUGFIX] Follow symlinks [#519]
* [BUGFIX] Improve GPG binary detection [#520] [#522]

## 1.6.4 / 2017-12-13

* [ENHANCEMENT] Support desktop notifications on Mac and Windows [#513]
* [BUGFIX] Fix slice out of bounds error [#517]
* [BUGFIX] Allow .password-store to be a symlink [#516]
* [BUGFIX] Respect --store flag to git sub command [#512]

## 1.6.3 / 2017-12-12

* [ENHANCEMENT] Avoid altering YAML secrets unless necessary [#508]
* [ENHANCEMENT] Documentation updates [#493] [#509]
* [ENHANCEMENT] Abort if no GPG binary was found [#506]
* [ENHANCEMENT] Support GOPASS_GPG_OPTS and GOPASS_UMASK [#504]
* [BUGFIX] Create .gpg-keys if it does not exist [#507]

## 1.6.2 / 2017-12-02

* [FEATURE] Add gopass fix command [#471]
* [ENHANCEMENT] Add pledge support on OpenBSD [#469]
* [ENHANCEMENT] Improve no clipboard warning [#484]
* [BUGFIX] Allow OTP entry in password field [#467]
* [BUGFIX] Default to vi if no other editor is available [#479]
* [BUGFIX] Avoid auto-search running non-interactively [#483]

## 1.6.1 / 2017-11-15

* [FEATURE] Add generic OTP action [#440]
* [ENHANCEMENT] Ignore any secret that does not end with .gpg [#461]
* [ENHANCEMENT] Add option to display only the password [#455]
* [ENHANCEMENT] Disable fuzzy search for gopass find [#454]
* [BUGFIX] Fix .gpg-id selection for sub folders [#465]
* [BUGFIX] Set gpg.program if possible [#464]
* [BUGFIX] Allow access to secrets shadowed by a folder [#463]
* [BUGFIX] Set GPG_TTY [#452]
* [BUGFIX] Fix termbox UI on OpenBSD [#446]
* [BUGFIX] Fix tests and paths on Windows [#421] [#431] [#442] [#450]

## 1.6.0 / 2017-11-03

* [FEATURE] Add Desktop notifications (Linux/DBus only) [#434] [#435]
* [ENHANCEMENT] Show public key identities before importing [#427]
* [ENHANCEMENT] Initialize local git config on gopass clone [#429]
* [ENHANCEMENT] Do not print generated passwords by default [#430]
* [ENHANCEMENT] Clear KDE Klipper History on clipboard clearing [#434]
* [ENHANCEMENT] Refactor git backend [#437]
* [BUGFIX] Fix recipients remove when using email as identifier [#436]

## 1.5.1 / 2017-10-25

* [ENHANCEMENT] Re-introduce usecolor config option [#414]
* [ENHANCEMENT] Improve documentation [#407] [#409] [#416] [#417]
* [ENHANCEMENT] Add language switch for xckd-style generation [#406]
* [BUGFIX] Fix GPG binary detection [#419]
* [BUGFIX] Fix tests on windows [#421]

## 1.5.0 / 2017-10-17

* [FEATURE] Add secret creation wizard [#386]
* [FEATURE] Add onboarding wizard [#387]
* [FEATURE] Wizard for recipients add/remove [#359]
* [FEATURE] XKCD#936 inspired password generation [#368]
* [FEATURE] Add update wizard [#395]
* [ENHANCEMENT] Overhaul documentation [#383] [#384]
* [ENHANCEMENT] Attempt to get TOTP key from YAML [#376]
* [ENHANCEMENT] Allow find to take -c [#378]
* [ENHANCEMENT] Improve terminal wizard [#385]
* [ENHANCEMENT] Improve responsiveness by context usage [#388]
* [ENHANCEMENT] Improve output readability [#392] [#393]
* [ENHANCEMENT] Automatic GPG key generation [#391]
* [BUGFIX] Relax YAML document marker handling [#398]

## 1.4.1 / 2017-10-05

* [BUGFIX] Support pre-1.3.0 configs [#382]
* [BUGFIX] Turn YAML errors into warnings [#380]

## 1.4.0 / 2017-10-04

* [FEATURE] Add fuzzy search [#317]
* [FEATURE] Allow restricting charset of generated passwords [#270]
* [FEATURE] Check quality of newly inserted passwords with crunchy [#276]
* [FEATURE] JSON API [#326]
* [FEATURE] Per-Mount configuration options [#330]
* [FEATURE] Terminal selection of results [#259]
* [FEATURE] gopass sync [#303]
* [ENHANCEMENT] Build with Go 1.9 [#294]
* [ENHANCEMENT] Display single find result directly [#265]
* [ENHANCEMENT] Global --yes flag [#327]
* [ENHANCEMENT] Improve error handling and propagation [#280]
* [ENHANCEMENT] Omit newline when not writing to a terminal [#325]
* [ENHANCEMENT] Only commit once per recipient batch operation [#329]
* [ENHANCEMENT] Provide partial support for .gpg-id files in sub folders [#291]
* [ENHANCEMENT] Trim any trailing newlines or carriage returns in show output [#296]
* [ENHANCEMENT] Use contexts [#310]
* [ENHANCEMENT] Use contexts to cancel long running operations [#358]
* [ENHANCEMENT] Use default editors [#286]
* [ENHANCEMENT] Improve documentation [#365]
* [ENHANCEMENT] Print selected entry [#372]
* [BUGFIX] Confirm removal of directories [#309]
* [BUGFIX] Only confirm recipients once during batch operations [#328]
* [BUGFIX] Only overwrite password on insert [#323]
* [BUGFIX] Avoid Show/Find recursion [#360]
* [BUGFIX] Remove deprecated special case for .yaml files [#362]
* [BUGFIX] Do not offer invalid keys [#364]
* [BUGFIX] Assign path only if resolving symlink succeeds [#370]

## 1.3.2 / 2017-08-22

* [BUGFIX] Fix git version output [#274]

## 1.3.1 / 2017-08-15

* [BUGFIX] Enable AutoSync by default [#267]
* [BUGFIX] git - do not abort if a store has no remote [#261]
* [BUGFIX] Fix IFS in bash completion [#268]

## 1.3.0 / 2017-08-11

* [BREAKING] Enforce YAML document markers [#193]
* [BREAKING] Simplify configuration [#213]
* [BREAKING] Align gopass init flags with other commands [#252]
* [FEATURE] Implement pager feature [#163]
* [FEATURE] Add basic fish completion [#168]
* [FEATURE] Add version check [#205]
* [FEATURE] Add gopass audit command [#228]
* [FEATURE] Add gopass audit hibp command [#239]
* [ENHANCEMENT] Disable auto-push while re-encrypting [#171]
* [ENHANCEMENT] Configure git user and email before initial git commit [#185]
* [ENHANCEMENT] Add recursive git operations [#186]
* [ENHANCEMENT] Document missing config options [#188]
* [ENHANCEMENT] Only check and load missing GPG keys after git pull [#190]
* [ENHANCEMENT] Only encrypt for valid recipients [#191]
* [ENHANCEMENT] Check and import missing GPG keys on recipients show [#204]
* [ENHANCEMENT] Save recipients on show [#207]
* [ENHANCEMENT] Include GPG and Git version in gopass version output [#210]
* [ENHANCEMENT] Support more flexible YAML documents [#217]
* [ENHANCEMENT] Simplify mounts add by inferring local path [#219]
* [ENHANCEMENT] Add contributor documentation [#222]
* [ENHANCEMENT] Re-use selected encryption key for git signing [#247]
* [ENHANCEMENT] Setup git push.default [#248]
* [BUGFIX] Fix nil-pointer check on non existing sub tree [#183]
* [BUGFIX] Fix load-keys [#203]
* [BUGFIX] Only match mounts on folders [#240]
* [BUGFIX] Disable checkRecipients as it conflicts with alwaysTrust [#242]

## 1.2.0 / 2017-06-21

* [FEATURE] YAML support [#125]
* [FEATURE] Binary support [#136]
* [ENHANCEMENT] Increase test coverage [#160]
* [ENHANCEMENT] Use secure temporary storage on MacOS [#144]
* [ENHANCEMENT] Use goreleaser [#151]
* [BUGFIX] Fix git invocation [#140]
* [BUGFIX] Fix missing recipients on init [#141]
* [BUGFIX] Fix sorting of mount points [#148]

## 1.1.2 / 2017-06-14

* [BUGFIX] Fix gopass init --store [#129]
* [BUGFIX] Fix gopass init [#127]

## 1.1.1 / 2017-06-13

* [ENHANCEMENT] Allow files and folders with the same name [#124]
* [ENHANCEMENT] Improve error messages [#121]
* [ENHANCEMENT] Add rm aliases to remove commands [#119]
* [BUGFIX] Several bug fixes for multi-repository handling [#123]

## 1.1.0 / 2017-05-31

* [FEATURE] Support templates [#1]
* [FEATURE] QR Code output [#64]
* [ENHANCEMENT] If entry was not found start search [#109]
* [ENHANCEMENT] Do not write color codes unless terminal [#111]
* [ENHANCEMENT] Make find compare case insensitive [#108]
* [ENHANCEMENT] Enforce UNIX style line endings [#105]
* [ENHANCEMENT] Use XDG_CONFIG_HOME [#67]
* [ENHANCEMENT] Support symlinks [#41]
* [ENHANCEMENT] Add nocolor config flag [#33]
* [ENHANCEMENT] Accept args for editor [#30]
* [BUGFIX] Build fixes for Windows [#14]

## 1.0.2 / 2017-03-24

* [ENHANCEMENT] Improve mounts and init commands [#87]
* [ENHANCEMENT] Document behavior of `-c` [#82]
* [ENHANCEMENT] Pass custom arguments to dmenu completion [#72]
* [ENHANCEMENT] Build with Go 1.8 [#65]
* [BUGFIX] Improve recursive deletes [#55]
* [BUGFIX] Bypass prompts on gopass insert --force [#66]
* [BUGFIX] Able to store secrets, but with errors [#13]
* [BUGFIX] Don't prompt if input from stdin [#58]
* [BUGFIX] Git add fails to "add" removed files [#57]

## 1.0.1 / 2017-02-13

* [FEATURE] Add dmenu support [#47]
* [ENHANCEMENT] Extend GOPASS_DEBUG coverage [#31]
* [ENHANCEMENT] Accept args for editor [#30]
* [ENHANCEMENT] Use gpg2 if available [#9]
* [BUGFIX] Fix git error handling in saveRecipients [#32]
* [BUGFIX] Check if ExpirationDate is set [#28]
* [BUGFIX] Change user.signkey to user.signingkey [#26]
* [BUGFIX] Only copy the first line to the clipboard [#21]
* [BUGFIX] Add search alias to find [#8]

## 1.0.0 / 2017-02-02

* [ENHANCEMENT] Support mounted sub-stores
* [ENHANCEMENT] git auto-push and auto-pull
* [ENHANCEMENT] git-style config editing
* [ENHANCEMENT] Simplified recipient management
* [ENHANCEMENT] Interactive questions for missing parameters
