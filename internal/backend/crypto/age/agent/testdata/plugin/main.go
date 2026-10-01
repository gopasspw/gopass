// This test plugin uses native age keys and requires no hardware.
package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"filippo.io/age"
	"filippo.io/age/plugin"
)

type identity struct {
	age.Identity
	plugin *plugin.Plugin
	mode   string
}

func (i *identity) Unwrap(stanzas []*age.Stanza) ([]byte, error) {
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
