package config

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/justwatchcom/gopass/pkg/backend"

	"github.com/pkg/errors"
)

// StoreConfig is a per-store (root or mount) config
type StoreConfig struct {
	AskForMore     bool         `yaml:"askformore"`     // ask for more data on generate
	AutoClip       bool         `yaml:"autoclip"`       // decide whether passwords are automatically copied or not
	AutoImport     bool         `yaml:"autoimport"`     // import missing public keys w/o asking
	AutoSync       bool         `yaml:"autosync"`       // push to git remote after commit, pull before push if necessary
	ClipTimeout    int          `yaml:"cliptimeout"`    // clear clipboard after seconds
	Concurrency    int          `yaml:"concurrency"`    // allow to run multiple thread when batch processing
	EditRecipients bool         `yaml:"editrecipients"` // edit recipients when confirming
	NoColor        bool         `yaml:"nocolor"`        // do not use color when outputing text
	NoConfirm      bool         `yaml:"noconfirm"`      // do not confirm recipients when encrypting
	NoPager        bool         `yaml:"nopager"`        // do not invoke a pager to display long lists
	Path           *backend.URL `yaml:"path"`           // path to the root store
	SafeContent    bool         `yaml:"safecontent"`    // avoid showing passwords in terminal
	UseSymbols     bool         `yaml:"usesymbols"`     // always use symbols when generating passwords
	Notifications  bool         `yaml:"notifications"`  // enable desktop notifications
}

func (c *StoreConfig) checkDefaults() error {
	if c == nil {
		return nil
	}
	if c.Path == nil {
		c.Path = backend.FromPath("")
	}
	if c.Concurrency == 0 {
		c.Concurrency = 1
	}
	return nil
}

// ConfigMap returns a map of stringified config values for easy printing
func (c *StoreConfig) ConfigMap() map[string]string {
	m := make(map[string]string, 20)
	o := reflect.ValueOf(c).Elem()
	for i := 0; i < o.NumField(); i++ {
		jsonArg := o.Type().Field(i).Tag.Get("yaml")
		if jsonArg == "" || jsonArg == "-" {
			continue
		}
		f := o.Field(i)
		var strVal string
		switch f.Kind() {
		case reflect.String:
			strVal = f.String()
		case reflect.Bool:
			strVal = fmt.Sprintf("%t", f.Bool())
		case reflect.Int:
			strVal = fmt.Sprintf("%d", f.Int())
		case reflect.Ptr:
			switch bup := f.Interface().(type) {
			case *backend.URL:
				if bup == nil {
					continue
				}
				strVal = bup.String()
			}
		default:
			continue
		}
		m[jsonArg] = strVal
	}
	return m
}

// SetConfigValue will try to set the given key to the value in the config struct
func (c *StoreConfig) SetConfigValue(key, value string) error {
	if key != "path" {
		value = strings.ToLower(value)
	}
	o := reflect.ValueOf(c).Elem()
	for i := 0; i < o.NumField(); i++ {
		jsonArg := o.Type().Field(i).Tag.Get("yaml")
		if jsonArg == "" || jsonArg == "-" {
			continue
		}
		if jsonArg != key {
			continue
		}
		f := o.Field(i)
		switch f.Kind() {
		case reflect.String:
			f.SetString(value)
			return nil
		case reflect.Bool:
			if value == "true" {
				f.SetBool(true)
				return nil
			} else if value == "false" {
				f.SetBool(false)
				return nil
			} else {
				return errors.Errorf("not a bool: %s", value)
			}
		case reflect.Int:
			iv, err := strconv.Atoi(value)
			if err != nil {
				return errors.Wrapf(err, "failed to convert '%s' to int", value)
			}
			f.SetInt(int64(iv))
			return nil
		default:
			continue
		}
	}
	return errors.New("unknown config option")
}

func (c *StoreConfig) String() string {
	return fmt.Sprintf("StoreConfig[AskForMore:%t,AutoClip:%t,AutoImport:%t,AutoSync:%t,ClipTimeout:%d,Concurrency:%d,EditRecipients:%t,NoColor:%t,NoConfirm:%t,NoPager:%t,Notifications:%t,Path:%s,SafeContent:%t,UseSymbols:%t]", c.AskForMore, c.AutoClip, c.AutoImport, c.AutoSync, c.ClipTimeout, c.Concurrency, c.EditRecipients, c.NoColor, c.NoConfirm, c.NoPager, c.Notifications, c.Path, c.SafeContent, c.UseSymbols)
}
