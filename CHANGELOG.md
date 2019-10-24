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
