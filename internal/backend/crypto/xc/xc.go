// Package xc implements a modern crypto backend for gopass.
// TODO(2.x) DEPRECATED and slated for removal in the 2.0.0 release.
package xc

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/gopasspw/gopass/internal/backend/crypto/xc/keyring"

	"github.com/blang/semver"
)

const (
	pubringFilename = ".gopass-pubring.xc"
	secringFilename = ".gopass-secring.xc"
	// Ext is the extension used by this backend
	Ext = "xc"
	// IDFile is the recipients list used by this backend
	IDFile = ".xc-ids"
)

type agentClient interface {
	Ping(context.Context) error
	Passphrase(context.Context, string, string) (string, error)
	Remove(context.Context, string) error
}

// XC is an experimental crypto backend
type XC struct {
	dir     string
	pubring *keyring.Pubring
	secring *keyring.Secring
	client  agentClient
}

// New creates a new XC backend
func New(dir string, client agentClient) *XC {
	skr, _ := keyring.LoadSecring(filepath.Join(dir, secringFilename))
	pkr, _ := keyring.LoadPubring(filepath.Join(dir, pubringFilename), skr)
	if client == nil {
		client = newAskPass()
	}
	return &XC{
		dir:     dir,
		pubring: pkr,
		secring: skr,
		client:  client,
	}
}

// RemoveKey removes a single key from the keyring
func (x *XC) RemoveKey(id string) error {
	if x.secring.Contains(id) {
		if err := x.secring.Remove(id); err != nil {
			return err
		}
		return x.secring.Save()
	}
	if x.pubring.Contains(id) {
		if err := x.pubring.Remove(id); err != nil {
			return err
		}
		return x.pubring.Save()
	}
	return fmt.Errorf("not found")
}

// Initialized returns an error if this backend is not properly initialized
func (x *XC) Initialized(ctx context.Context) error {
	if x == nil {
		return fmt.Errorf("XC not initialized")
	}
	if x.pubring == nil {
		return fmt.Errorf("pubring not initialized")
	}
	if x.secring == nil {
		return fmt.Errorf("secring not initialized")
	}
	if x.client == nil {
		return fmt.Errorf("client not initialized")
	}
	if err := x.client.Ping(ctx); err != nil {
		return fmt.Errorf("agent not running")
	}
	return nil
}

// Name returns xc
func (x *XC) Name() string {
	return "xc"
}

// Version returns 0.0.1
func (x *XC) Version(ctx context.Context) semver.Version {
	return semver.Version{
		Patch: 1,
	}
}

// Ext returns xc
func (x *XC) Ext() string {
	return Ext
}

// IDFile returns .xc-ids
func (x *XC) IDFile() string {
	return IDFile
}
