package config

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

// StoreConfig is a per-store (root or mount) config
type StoreConfig struct {
	AskForMore  bool   `yaml:"askformore"`  // ask for more data on generate
	AutoImport  bool   `yaml:"autoimport"`  // import missing public keys w/o asking
	AutoSync    bool   `yaml:"autosync"`    // push to git remote after commit, pull before push if necessary
	ClipTimeout int    `yaml:"cliptimeout"` // clear clipboard after seconds
	NoConfirm   bool   `yaml:"noconfirm"`   // do not confirm recipients when encrypting
	NoPager     bool   `yaml:"nopager"`     // do not invoke a pager to display long lists
	Path        string `yaml:"path"`        // path to the root store
	SafeContent bool   `yaml:"safecontent"` // avoid showing passwords in terminal
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
		strVal := ""
		switch f.Kind() {
		case reflect.String:
			strVal = f.String()
		case reflect.Bool:
			strVal = fmt.Sprintf("%t", f.Bool())
		case reflect.Int:
			strVal = fmt.Sprintf("%d", f.Int())
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
		case reflect.Bool:
			if value == "true" {
				f.SetBool(true)
			} else if value == "false" {
				f.SetBool(false)
			} else {
				return errors.Errorf("No a bool: %s", value)
			}
		case reflect.Int:
			iv, err := strconv.Atoi(value)
			if err != nil {
				return errors.Wrapf(err, "failed to convert '%s' to int", value)
			}
			f.SetInt(int64(iv))
		default:
			continue
		}
	}
	return nil
}
