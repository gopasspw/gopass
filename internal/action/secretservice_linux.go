//go:build linux

package action

import (
	"context"
	"fmt"

	"github.com/godbus/dbus/v5"
	"github.com/gopasspw/gopass/internal/out"
	"github.com/gopasspw/gopass/internal/secretservice"
	"github.com/urfave/cli/v3"
)

// secretServiceCommand returns the `gopass secret-service` command. It is
// Linux-only because it speaks the freedesktop.org D-Bus Secret Service API.
func (s *Action) secretServiceCommand() *cli.Command {
	return &cli.Command{
		Name:  "secret-service",
		Usage: "Run a D-Bus Secret Service daemon backed by gopass",
		Description: "" +
			"Implements the org.freedesktop.secrets D-Bus API so that desktop " +
			"applications (Firefox, Chromium, VS Code, Electron apps, " +
			"NetworkManager, secret-tool, ...) store their secrets in the same " +
			"gopass password store as the CLI.",
		Commands: []*cli.Command{
			{
				Name:  "serve",
				Usage: "Start the Secret Service daemon (blocks until interrupted)",
				Description: "" +
					"Connects to the D-Bus session bus, acquires the well-known name " +
					"org.freedesktop.secrets and serves the Secret Service API until " +
					"the process is interrupted (e.g. with Ctrl+C). Secrets are " +
					"stored in the gopass store below --prefix; the volatile session " +
					"collection is kept in memory only.",
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:  "replace",
						Usage: "Replace any existing Secret Service provider (e.g. GNOME Keyring)",
					},
					&cli.StringFlag{
						Name:  "prefix",
						Value: "secret-service",
						Usage: "gopass path prefix under which secrets are stored",
					},
					&cli.BoolFlag{
						Name:  "notify-on-access",
						Usage: "Send a desktop notification whenever a secret is read",
					},
				},
				Action: s.SecretServiceServe,
			},
			{
				Name:  "status",
				Usage: "Check whether a Secret Service provider is running",
				Description: "" +
					"Queries the D-Bus session bus for the owner of " +
					"org.freedesktop.secrets. It reports whether any provider " +
					"(gopass or another one such as GNOME Keyring) currently serves " +
					"the Secret Service API.",
				Action: s.SecretServiceStatus,
			},
			{
				Name:  "install",
				Usage: "Install the systemd user unit and D-Bus session activation file",
				Description: "" +
					"Writes a systemd user unit (gopass-secret-service.service) and a " +
					"D-Bus session activation file (org.freedesktop.secrets.service). " +
					"Run `systemctl --user daemon-reload` and then " +
					"`systemctl --user enable --now gopass-secret-service` afterwards.",
				Action: s.SecretServiceInstall,
			},
			{
				Name:  "uninstall",
				Usage: "Remove the systemd user unit and D-Bus session activation file",
				Description: "" +
					"Removes the systemd user unit and the D-Bus session activation " +
					"file written by `gopass secret-service install`. Missing files " +
					"are ignored, so the command is safe to run repeatedly.",
				Action: s.SecretServiceUninstall,
			},
		},
	}
}

// SecretServiceServe runs the Secret Service daemon.
func (*Action) SecretServiceServe(ctx context.Context, cmd *cli.Command) error {
	svc, err := secretservice.New(ctx, secretservice.Config{
		Prefix:         cmd.String("prefix"),
		Replace:        cmd.Bool("replace"),
		NotifyOnAccess: cmd.Bool("notify-on-access"),
	})
	if err != nil {
		return err
	}

	out.Printf(ctx, "Starting gopass secret-service on org.freedesktop.secrets")
	out.Printf(ctx, "Press Ctrl+C to stop.")

	return svc.Serve(ctx)
}

// SecretServiceStatus reports whether a Secret Service provider currently owns
// the well-known bus name.
func (*Action) SecretServiceStatus(ctx context.Context, _ *cli.Command) error {
	conn, err := dbus.SessionBus()
	if err != nil {
		return fmt.Errorf("failed to connect to the session bus: %w", err)
	}

	var hasOwner bool
	if err := conn.BusObject().Call("org.freedesktop.DBus.NameHasOwner", 0, secretservice.ServiceName).Store(&hasOwner); err != nil {
		return fmt.Errorf("failed to query the bus: %w", err)
	}

	if hasOwner {
		out.Printf(ctx, "%s is owned (a Secret Service provider is running)", secretservice.ServiceName)

		return nil
	}

	out.Printf(ctx, "no provider owns %s", secretservice.ServiceName)

	return nil
}

// SecretServiceInstall writes the systemd user unit and the D-Bus session
// activation file for the current user.
func (*Action) SecretServiceInstall(ctx context.Context, _ *cli.Command) error {
	paths := secretservice.DefaultInstallPaths()

	written, err := secretservice.Install(paths, "")
	if err != nil {
		return err
	}

	for _, p := range written {
		out.Printf(ctx, "Installed %s", p)
	}

	out.Printf(ctx, "")
	out.Printf(ctx, "To start the daemon now, run:")
	out.Printf(ctx, "  systemctl --user daemon-reload")
	out.Printf(ctx, "  systemctl --user enable --now gopass-secret-service")
	out.Printf(ctx, "")
	out.Printf(ctx, "If another provider (e.g. GNOME Keyring) owns %s, start the", secretservice.ServiceName)
	out.Printf(ctx, "daemon with `gopass secret-service serve --replace` or disable that provider.")

	return nil
}

// SecretServiceUninstall removes the systemd user unit and the D-Bus session
// activation file.
func (*Action) SecretServiceUninstall(ctx context.Context, _ *cli.Command) error {
	paths := secretservice.DefaultInstallPaths()

	removed, err := secretservice.Uninstall(paths)
	if err != nil {
		return err
	}

	if len(removed) == 0 {
		out.Printf(ctx, "Nothing to remove.")

		return nil
	}

	for _, p := range removed {
		out.Printf(ctx, "Removed %s", p)
	}

	out.Printf(ctx, "")
	out.Printf(ctx, "Run `systemctl --user daemon-reload` to apply the change.")

	return nil
}
