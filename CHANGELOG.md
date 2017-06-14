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
* [ENHANCEMENT] Make find compare case insenstive [#108]
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
