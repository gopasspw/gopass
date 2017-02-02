package action

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strconv"
	"strings"

	"github.com/ghodss/yaml"
	"github.com/justwatchcom/gopass/fsutil"
	"github.com/justwatchcom/gopass/password"
	"github.com/urfave/cli"
)

// Config handles changes to the gopass configuration
func (s *Action) Config(c *cli.Context) error {
	if len(c.Args()) < 1 {
		return s.printConfigValues()
	}

	if len(c.Args()) == 1 {
		return s.printConfigValues(c.Args()[0])
	}

	if len(c.Args()) > 2 {
		return fmt.Errorf("Usage: gopass config key value")
	}

	return s.setConfigValue(c.Args()[0], c.Args()[1])
}

func (s *Action) printConfigValues(filter ...string) error {
	out := make([]string, 0, 10)
	o := reflect.ValueOf(s.Store).Elem()
	for i := 0; i < o.NumField(); i++ {
		jsonArg := o.Type().Field(i).Tag.Get("json")
		if jsonArg == "" || jsonArg == "-" {
			continue
		}
		if !contains(filter, jsonArg) {
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
		out = append(out, fmt.Sprintf("%s: %s", jsonArg, strVal))
	}
	sort.Strings(out)
	for _, line := range out {
		fmt.Println(line)
	}
	return nil
}

func contains(haystack []string, needle string) bool {
	if len(haystack) < 1 {
		return true
	}
	for _, blade := range haystack {
		if blade == needle {
			return true
		}
	}
	return false
}

func (s *Action) setConfigValue(key, value string) error {
	if key == "version" {
		return fmt.Errorf("Can not change version")
	}
	if key != "path" {
		value = strings.ToLower(value)
	}
	o := reflect.ValueOf(s.Store).Elem()
	for i := 0; i < o.NumField(); i++ {
		jsonArg := o.Type().Field(i).Tag.Get("json")
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
				return fmt.Errorf("No a bool: %s", value)
			}
		case reflect.Int:
			iv, err := strconv.Atoi(value)
			if err != nil {
				return err
			}
			f.SetInt(int64(iv))
		default:
			continue
		}
	}
	return writeConfig(s.Store)
}

// hasConfig is a short hand for checking if the config file exists
func hasConfig() bool {
	return fsutil.IsFile(configFile())
}

// writeConfig saves the config
func writeConfig(s *password.RootStore) error {
	buf, err := yaml.Marshal(s)
	if err != nil {
		return err
	}
	if err := ioutil.WriteFile(configFile(), buf, 0600); err != nil {
		return err
	}
	return nil
}

// configFile returns the location of the config file. Either reading from
// GOPASS_CONFIG or using the default location (~/.gopass.yml)
func configFile() string {
	if cf := os.Getenv("GOPASS_CONFIG"); cf != "" {
		return cf
	}
	return filepath.Join(os.Getenv("HOME"), ".gopass.yml")
}
