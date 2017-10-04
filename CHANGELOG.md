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
* [ENHANCEMENT] Provide partial support for .gpg-id files in subfolders [#291]
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
* [BUGFIX] git - do not abort if a store has not remote [#261]
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
* [ENHANCEMENT] Disable auto-push while reencrypting [#171]
* [ENHANCEMENT] Configure git user and email before initial git commit [#185]
* [ENHANCEMENT] Add recursive git operations [#186]
* [ENHANCEMENT] Document missing config options [#188]
* [ENHANCEMENT] Only check and load missing GPG keys after git pull [#190]
* [ENHANCEMENT] Only encrypt for valid recipients [#191]
* [ENHANCEMENT] Check and import missing GPG keys on recipients show [#204]
* [ENHANCEMENT] Save recipients on show [#207]
* [ENHANCEMENT] Include GPG and Git version in gopass version output [#210]
* [ENHANCEMENT] Support more flexible YAML documents [#217]
* [ENHANCEMENT] Simplify mounts add by infering local path [#219]
* [ENHANCEMENT] Add contributor documentation [#222]
* [ENHANCEMENT] Re-use selected encryption key for git signing [#247]
* [ENHANCEMENT] Setup git push.default [#248]
* [BUGFIX] Fix nil-pointer check on non existing subtree [#183]
* [BUGFIX] Fix load-keys [#203]
* [BUGFIX] Only match mounts on folders [#240]
* [BUGFIX] Disable checkRecipients as it conflicts with alwaysTrust [#242]

## 1.2.0 / 2017-06-21

* [FEATURE] YAML support [#125]
* [FEATURE] Binary support [#136]
* [ENHANCEMENT] Increase test coverage [#160]
* [ENHANCEMENT] Use secure temporary storage on macOS [#144]
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
