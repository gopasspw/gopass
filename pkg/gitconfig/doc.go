// Package gitconfig implements a pure Go parser of Git SCM config files. The support
// is currently not matching git exactly, e.g. includes, urlmatches and multivars are currently
// not supported. And while we try to preserve the original file a much as possible
// when writing we currently don't exactly retain (insignificant) whitespaces.
//
// The reference for this implementation is https://mirrors.edge.kernel.org/pub/software/scm/git/docs/git-config.html
//
// # Usage
//
// Use gitconfig.LoadAll with an optional workspace argument to process configuration
// input from these locations in order (i.e. the later ones take precedence):
//
//   - `system` - /etc/gitconfig
//   - `global` - `$XDG_CONFIG_HOME/git/config` or `~/.gitconfig`
//   - `local` - `<workdir>/config`
//   - `worktree` - `<workdir>/config.worktree`
//   - `command` - GIT_CONFIG_{COUNT,KEY,VALUE} environment variables
//
// Note: We do not support parsing command line flags directly, but one
// can use the SetEnv method to set flags from the command line in the config.
//
// # Customization
//
// `gopass` and other users of this package can easily customize file and environment
// names by utilizing the exported variables from the Configs struct:
//
//   - SystemConfig
//   - GlobalConfig (can be set to the empty string to disable)
//   - LocalConfig
//   - WorktreeConfig
//   - EnvPrefix
//
// Note: For tests users will want to set `NoWrites = true` to avoid overwriting
// their real configs.
//
// Example
//
//	import "github.com/gopasspw/gopass/pkg/gitconfig"
//
//	func main() {
//		cfg := gitconfig.New()
//		cfg.SystemConfig = "/etc/gopass/config"
//		cfg.GlobalConfig = ""
//		cfg.EnvPrefix = "GOPASS_CONFIG"
//		cfg.LoadAll(".")
//		_ = cfg.Get("core.notifications")
//	}
//
// # Versioning and Compatibility
//
// We aim to support the latest stable release of Git only.
// Currently we do not provide any backwards compatibility
// and semantic versioning. Once this package has become
// mostly feature complete and if there is interest from
// other projects in using it we may choose to move it to
// it's own repository and start proper versioning.
//
// # Known limitations
//
// * Worktree support is only partial
package gitconfig
