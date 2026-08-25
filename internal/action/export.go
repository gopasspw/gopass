package action

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/gopasspw/gopass/internal/action/exit"
	"github.com/gopasspw/gopass/internal/out"
	"github.com/gopasspw/gopass/internal/tree"
	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/urfave/cli/v3"
)

// exportedSecret is the JSON representation of a single secret. It keeps the
// password, the free-form body and all key/value attributes so the dump is
// lossless and can be transformed into whatever format another password
// manager expects on import.
type exportedSecret struct {
	Name       string              `json:"name"`
	Password   string              `json:"password,omitempty"`
	Body       string              `json:"body,omitempty"`
	Attributes map[string][]string `json:"attributes,omitempty"`
}

// Export writes all secrets, or those below the given prefixes, to a portable
// JSON document. The document goes to stdout unless --out names a file.
func (s *searchHandler) Export(ctx context.Context, cmd *cli.Command) error {
	ctx = ctxutil.WithGlobalFlags(ctx, cmd)

	names, err := s.Store.List(ctx, tree.INF)
	if err != nil {
		return exit.Error(exit.List, err, "failed to list store: %s", err)
	}
	sort.Strings(names)

	prefixes := cmd.Args().Slice()

	// The output is plaintext, so make sure the user understands that before
	// it ends up in a file or the shell history.
	out.Warningf(ctx, "Exported secrets are stored UNENCRYPTED. Handle the output with care.")

	exported := make([]exportedSecret, 0, len(names))
	var failed int
	for _, name := range names {
		if !hasAnyPrefix(name, prefixes) {
			continue
		}

		sec, err := s.Store.Get(ctx, name)
		if err != nil {
			out.Errorf(ctx, "Failed to decrypt %s: %v", name, err)
			failed++

			continue
		}

		es := exportedSecret{
			Name:     name,
			Password: sec.Password(),
			Body:     sec.Body(),
		}
		if keys := sec.Keys(); len(keys) > 0 {
			es.Attributes = make(map[string][]string, len(keys))
			for _, k := range keys {
				if vs, ok := sec.Values(k); ok {
					es.Attributes[k] = vs
				}
			}
		}
		exported = append(exported, es)
	}

	buf, err := json.MarshalIndent(exported, "", "  ")
	if err != nil {
		return exit.Error(exit.Unknown, err, "failed to encode secrets: %s", err)
	}
	buf = append(buf, '\n')

	if fn := cmd.String("out"); fn != "" {
		if err := os.WriteFile(fn, buf, 0o600); err != nil {
			return exit.Error(exit.IO, err, "failed to write %s: %s", fn, err)
		}
		out.Noticef(ctx, "Exported %d secrets to %s", len(exported), fn)
	} else {
		fmt.Fprint(stdout, string(buf))
	}

	if failed > 0 {
		out.Warningf(ctx, "%d secrets failed to decrypt", failed)
	}

	return nil
}

// hasAnyPrefix reports whether name is covered by any of the given prefixes. An
// empty prefix list matches everything. A prefix matches an exact name or a
// folder below it, so "foo" matches "foo" and "foo/bar" but not "foobar".
func hasAnyPrefix(name string, prefixes []string) bool {
	if len(prefixes) == 0 {
		return true
	}

	for _, p := range prefixes {
		if name == p || strings.HasPrefix(name, strings.TrimSuffix(p, "/")+"/") {
			return true
		}
	}

	return false
}
