package gitconfig

var (
	// GlobalConfig is the location of the (optional) global (i.e. user-wide) config file.
	globalConfig = ".gitconfig"
	// LocalConfig is the name of the local (per-workdir) configuration.
	localConfig = "config"
	// WorktreeConfig is the name of the local worktree configuration. Can be used to override
	// a committed local config.
	worktreeConfig = "config.worktree"
	// EnvPrefix is the prefix for the environment variables controlling and overriding config variables.
	envPrefix = "GIT_CONFIG"
	// Name is the name of the config directory (e.g. git or gopass).
	name = "git"
)
