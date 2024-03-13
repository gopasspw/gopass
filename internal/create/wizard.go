package create

import (
	"context"
	"fmt"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/gopasspw/gopass/internal/backend"
	"github.com/gopasspw/gopass/internal/config"
	"github.com/gopasspw/gopass/internal/cui"
	"github.com/gopasspw/gopass/internal/hook"
	"github.com/gopasspw/gopass/internal/out"
	"github.com/gopasspw/gopass/internal/set"
	"github.com/gopasspw/gopass/internal/store/root"
	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/gopasspw/gopass/pkg/debug"
	"github.com/gopasspw/gopass/pkg/fsutil"
	"github.com/gopasspw/gopass/pkg/gopass/secrets"
	"github.com/gopasspw/gopass/pkg/pwgen"
	"github.com/gopasspw/gopass/pkg/pwgen/pwrules"
	"github.com/gopasspw/gopass/pkg/termio"
	"github.com/martinhoefling/goxkcdpwgen/xkcdpwgen"
	"github.com/urfave/cli/v2"
	"gopkg.in/yaml.v3"
)

const (
	tplPath = ".gopass/create/"
)

// Attribute is a credential attribute that is being asked for
// when populating a template.
type Attribute struct {
	Name         string `yaml:"name"`
	Type         string `yaml:"type"`
	Prompt       string `yaml:"prompt"`
	Charset      string `yaml:"charset"`
	Min          int    `yaml:"min"`
	Max          int    `yaml:"max"`
	AlwaysPrompt bool   `yaml:"always_prompt"` // always prompt for the crendentials
}

// Template is an action template for the create wizard.
type Template struct {
	Name       string      `yaml:"name"`
	Priority   int         `yaml:"priority"`
	Prefix     string      `yaml:"prefix"`
	NameFrom   []string    `yaml:"name_from"`
	Welcome    string      `yaml:"welcome"`
	Attributes []Attribute `yaml:"attributes"`
}

// Wizard is the templateable credential creation wizard.
type Wizard struct {
	Templates []Template
}

// New creates a new instance of the wizard. It will parse the user
// supplied templates and add the default templates.
func New(ctx context.Context, s backend.Storage) (*Wizard, error) {
	w := &Wizard{}

	tpls, err := w.parseTemplates(ctx, s)
	if err != nil {
		return w, fmt.Errorf("could not parse templates: %w", err)
	}

	if len(tpls) < 1 {
		// no templates found, write default templates
		if err := w.writeTemplates(ctx, s); err != nil {
			return w, fmt.Errorf("could not write default templates: %w", err)
		}

		// then re-parse them
		tpls, err = w.parseTemplates(ctx, s)
		if err != nil {
			return w, fmt.Errorf("could not parse templates: %w", err)
		}
	}

	w.Templates = tpls

	return w, nil
}

func (w *Wizard) parseTemplates(ctx context.Context, s backend.Storage) ([]Template, error) {
	tpls, err := s.List(ctx, tplPath)
	if err != nil {
		return nil, err
	}

	parsedTpls := []Template{}
	for _, f := range tpls {
		if !strings.HasSuffix(f, ".yml") && !strings.HasSuffix(f, ".yaml") {
			debug.Log("ignoring unknown file extension: %s", f)

			continue
		}
		buf, err := s.Get(ctx, f)
		if err != nil {
			debug.Log("failed to read template %s: %s", f, err)

			continue
		}
		tpl := Template{}
		if err := yaml.Unmarshal(buf, &tpl); err != nil {
			debug.Log("failed to parse template %s: %s", f, err)
			out.Errorf(ctx, "Bad template %s: %s\n%s", f, err, string(buf))

			continue
		}

		parsedTpls = append(parsedTpls, tpl)
	}

	sort.Slice(parsedTpls, func(i, j int) bool {
		return parsedTpls[i].Priority < parsedTpls[j].Priority
	})

	return parsedTpls, nil
}

// ActionCallback is the callback for the creation calls to print and copy the credentials.
type ActionCallback func(context.Context, *cli.Context, string, string, bool) error

// Actions returns a list of actions that can be performed on the wizard. The actions directly
// interact with the underlying storage.
func (w *Wizard) Actions(s *root.Store, cb ActionCallback) cui.Actions {
	sort.Slice(w.Templates, func(i, j int) bool {
		return w.Templates[i].Priority < w.Templates[j].Priority
	})

	acts := make(cui.Actions, 0, len(w.Templates))
	for _, tpl := range w.Templates {
		acts = append(acts, cui.Action{
			Name: tpl.Name,
			Fn:   mkActFunc(tpl, s, cb),
		})
	}

	return acts
}

func mkActFunc(tpl Template, s *root.Store, cb ActionCallback) func(context.Context, *cli.Context) error { //nolint:cyclop
	debug.Log("creating action func for %+v, cb: %p", tpl, cb)

	return func(ctx context.Context, c *cli.Context) error {
		name := c.Args().First()
		store := c.String("store")

		// select store.
		if store == "" {
			store = cui.AskForStore(ctx, s)
		}
		ctx = config.WithMount(ctx, store)

		force := c.Bool("force")

		if err := hook.Invoke(ctx, "create.pre-hook", name); err != nil {
			return err
		}

		sec := secrets.New()

		out.Print(ctx, tpl.Welcome)

		// genPW is needed for the callback
		var genPw bool
		// password is needed for the callback
		var password string
		// hostname is needed in later iterations (e.g. password rule lookup)
		var hostname string
		// wantForName is a list of attributes that will be used to build the name
		wantForName := set.Map(tpl.NameFrom)
		// nameParts are the components the name will be built from
		var nameParts []string
		// step is only used for printing the progress
		var step int
		for _, v := range tpl.Attributes {
			step++
			k := v.Name

			// if no prompt is set default to the key
			if v.Prompt == "" {
				v.Prompt = strings.ToTitle(k)
			}

			switch v.Type {
			case "string":
				sv, err := termio.AskForString(ctx, fmtfn(2, strconv.Itoa(step), v.Prompt), "")
				if err != nil {
					return err
				}
				if v.Min > 0 && len(sv) < v.Min {
					return fmt.Errorf("%s is too short (needs %d)", v.Name, v.Min)
				}
				if v.Max > 0 && len(sv) > v.Min {
					return fmt.Errorf("%s is too long (at most %d)", v.Name, v.Max)
				}
				if wantForName[k] {
					nameParts = append(nameParts, sv)
				}
				_ = sec.Set(k, sv)
			case "hostname":
				var def string
				if k == "username" {
					def = config.String(ctx, "create.default-username")
				}
				sv, err := termio.AskForString(ctx, fmtfn(2, strconv.Itoa(step), v.Prompt), def)
				if err != nil {
					return err
				}
				hostname = extractHostname(sv)
				if hostname == "" {
					return fmt.Errorf("can not parse URL %s", sv)
				}
				if wantForName[k] {
					nameParts = append(nameParts, hostname)
				}
				if u := pwrules.LookupChangeURL(ctx, hostname); u != "" {
					_ = sec.Set("password-change-url", u)
				}
				_ = sec.Set(k, sv)
			case "password":
				var err error
				if !v.AlwaysPrompt {
					genPw, err = termio.AskForBool(ctx, fmtfn(2, strconv.Itoa(step), "Generate Password?"), true)
					if err != nil {
						return err
					}
				}

				if genPw { //nolint:nestif
					password, err = generatePassword(ctx, hostname, v.Charset)
					if err != nil {
						return err
					}
				} else {
					password, err = termio.AskForPassword(ctx, v.Prompt, true)
					if err != nil {
						return err
					}
					if v.Min > 0 && len(password) < v.Min {
						return fmt.Errorf("%s is too short (needs %d)", v.Name, v.Min)
					}
					if v.Max > 0 && len(password) > v.Min {
						return fmt.Errorf("%s is too long (at most %d)", v.Name, v.Max)
					}
				}

				sec.SetPassword(password)
			}
		}

		// now we can generate a name. If it's already take we can the user for an alternative
		// name.

		// make sure the store is properly separated from the name.
		if store != "" {
			store += "/"
		}

		// by default create will generate a name for the secret based on the user
		// input. Only when the force flag is given it will accept a secrets path
		// as the first argument.
		if name == "" || !force {
			for i, s := range nameParts {
				nameParts[i] = fsutil.CleanFilename(s)
			}
			name = fmt.Sprintf("%s%s/%s", store, tpl.Prefix, filepath.Join(nameParts...))
		}
		if force && !strings.HasPrefix(name, store) {
			out.Warningf(ctx, "User supplied secret name %q does not match requested mount %q. Ignoring store flag.", name, store)
		}

		// force will also override the check for existing entries.
		if s.Exists(ctx, name) && !force {
			step++
			var err error
			name, err = termio.AskForString(ctx, fmtfn(2, strconv.Itoa(step), "Secret already exists. Choose another path or enter to overwrite"), name)
			if err != nil {
				return err
			}
		}

		if err := s.Set(ctxutil.WithCommitMessage(ctx, "Created new entry"), name, sec); err != nil {
			return fmt.Errorf("failed to set %q: %w", name, err)
		}
		out.OKf(ctx, "Credentials saved to %q", name)

		return cb(ctx, c, name, password, genPw)
	}
}

// generatePasssword will walk through the password generation steps.
func generatePassword(ctx context.Context, hostname, charset string) (string, error) {
	defaultLength, _ := config.DefaultPasswordLengthFromEnv(ctx)

	if charset != "" {
		length, err := termio.AskForInt(ctx, fmtfn(4, "a", "How long?"), 4)
		if err != nil {
			return "", err
		}

		return pwgen.GeneratePasswordCharset(length, charset), nil
	}
	if _, found := pwrules.LookupRule(ctx, hostname); found {
		out.Noticef(ctx, "Using password rules for %s ...", hostname)
		length, err := termio.AskForInt(ctx, fmtfn(4, "b", "How long?"), defaultLength)
		if err != nil {
			return "", err
		}

		return pwgen.NewCrypticForDomain(ctx, length, hostname).Password(), nil
	}
	xkcd, err := termio.AskForBool(ctx, fmtfn(4, "a", "Human-pronounceable passphrase?"), false)
	if err != nil {
		return "", err
	}
	if xkcd {
		length, err := termio.AskForInt(ctx, fmtfn(4, "b", "How many words?"), config.DefaultXKCDLength)
		if err != nil {
			return "", err
		}
		g := xkcdpwgen.NewGenerator()
		g.SetNumWords(length)
		g.SetDelimiter(" ")
		g.SetCapitalize(true)

		return string(g.GeneratePassword()), nil
	}

	length, err := termio.AskForInt(ctx, fmtfn(4, "b", "How long?"), defaultLength)
	if err != nil {
		return "", err
	}

	symbols, err := termio.AskForBool(ctx, fmtfn(4, "c", "Include symbols?"), false)
	if err != nil {
		return "", err
	}

	corp, err := termio.AskForBool(ctx, fmtfn(4, "d", "Strict rules?"), false)
	if err != nil {
		return "", err
	}
	if corp {
		return pwgen.GeneratePasswordWithAllClasses(length, symbols)
	}

	return pwgen.GeneratePassword(length, symbols), nil
}
