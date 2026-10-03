// This test plugin uses native age keys and requires no hardware.
package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"filippo.io/age"
	"filippo.io/age/plugin"
)

type identity struct {
	age.Identity
	plugin *plugin.Plugin
	mode   string
}

func (i *identity) Unwrap(stanzas []*age.Stanza) ([]byte, error) {
	if path := os.Getenv("GOPASS_TEST_PLUGIN_COUNT"); path != "" {
		f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600)
		if err != nil {
			return nil, err
		}
		_, err = f.WriteString("unwrap\n")
		_ = f.Close()
		if err != nil {
			return nil, err
		}
	}
	if ready := os.Getenv("GOPASS_TEST_PLUGIN_READY"); ready != "" {
		if err := os.WriteFile(ready, []byte("ready"), 0o600); err != nil {
			return nil, err
		}
		deadline := time.Now().Add(10 * time.Second)
		for {
			if _, err := os.Stat(os.Getenv("GOPASS_TEST_PLUGIN_RELEASE")); err == nil {
				break
			}
			if time.Now().After(deadline) {
				return nil, fmt.Errorf("test hardware authentication timed out")
			}
			time.Sleep(10 * time.Millisecond)
		}
	}
	if i.mode == "cancel" {
		return nil, fmt.Errorf("test authentication canceled")
	}
	var err error
	switch i.mode {
	case "msg":
		err = i.plugin.DisplayMessage("test message")
	case "request-secret", "request-public":
		_, err = i.plugin.RequestValue("test value", i.mode == "request-secret")
	case "confirm":
		_, err = i.plugin.Confirm("test confirmation", "yes", "no")
	}
	if err != nil {
		return nil, err
	}
	if i.mode != "decrypt" {
		return nil, fmt.Errorf("unexpected successful interaction: %s", i.mode)
	}
	return i.Identity.Unwrap(stanzas)
}

func main() {
	p, err := plugin.New("gopasstest")
	if err != nil {
		panic(err)
	}
	p.HandleIdentity(func(data []byte) (age.Identity, error) {
		mode, key, ok := strings.Cut(string(data), "|")
		if !ok {
			return nil, fmt.Errorf("missing test mode")
		}
		id, err := age.ParseX25519Identity(key)
		if err != nil {
			return nil, err
		}
		return &identity{Identity: id, plugin: p, mode: mode}, nil
	})
	p.RegisterFlags(nil)
	flag.Parse()
	os.Exit(p.Main())
}
