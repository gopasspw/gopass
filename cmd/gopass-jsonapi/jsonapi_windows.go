// +build windows

package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/gopasspw/gopass/cmd/gopass-jsonapi/internal/jsonapi/manifest"
	"github.com/gopasspw/gopass/internal/out"
	"github.com/gopasspw/gopass/internal/termio"
	"github.com/gopasspw/gopass/pkg/ctxutil"

	"github.com/fatih/color"
	"github.com/urfave/cli/v2"
	"golang.org/x/sys/windows/registry"
)

// setup sets up manifest for gopass as native messaging host
func (s *jsonapiCLI) setup(c *cli.Context) error {
	ctx := ctxutil.WithGlobalFlags(c)
	browser, err := s.getBrowser(ctx, c)
	if err != nil {
		return fmt.Errorf("failed to get browser: %s", err)
	}

	globalInstall, err := s.getGlobalInstall(ctx, c)
	if err != nil {
		return fmt.Errorf("failed to get global flag: %s", err)
	}

	// Use windows specific folder to store wrapper and manifests
	defaultWrapperPath := filepath.Join(os.Getenv("LOCALAPPDATA"), "gopass")
	if globalInstall {
		defaultWrapperPath = filepath.Join(os.Getenv("PROGRAMDATA"), "gopass")
	}

	wrapperPath, err := s.getWrapperPath(ctx, c, defaultWrapperPath, manifest.NativeHostExeName)
	if err != nil {
		return fmt.Errorf("failed to get wrapper path: %s", err)
	}
	wrapperFileName := filepath.Join(wrapperPath, manifest.NativeHostExeName)

	manifestPath := c.String("manifest-path")
	if manifestPath == "" {
		manifestPath = filepath.Join(wrapperPath, browser, manifest.Name+".json")
	}

	regPath, err := manifest.GetRegistryPath(browser)
	if err != nil {
		return fmt.Errorf("failed to get registry path: %s", err)
	}

	_, mf, err := manifest.Render(browser, wrapperFileName, c.String("gopass-path"), globalInstall)
	if err != nil {
		return fmt.Errorf("failed to render manifest: %s", err)
	}

	if c.Bool("print") {
		out.Print(ctx, "Native Messaging Setup Preview:\nWrapper Script (%s)\n\nManifest File (%s = %s):\n%s\n", wrapperFileName, regPath, manifestPath, string(mf))
	}

	if install, err := termio.AskForBool(ctx, color.BlueString("Install manifest and wrapper?"), true); err != nil || !install {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(wrapperFileName), 0755); err != nil {
		return fmt.Errorf("failed to create wrapper path: %s", err)
	}

	if err := s.setRegistryValue(regPath, manifestPath, globalInstall); err != nil {
		return fmt.Errorf("failed to set registry value: %s", err)
	}

	// If the calling binary has native_host.exe as suffix listener will be started.
	if err := s.copyExecutionBinary(wrapperFileName); err != nil {
		return fmt.Errorf("failed to copy gopass binary to wrapper path: %s", err)
	}

	if err := os.MkdirAll(filepath.Dir(manifestPath), 0755); err != nil {
		return fmt.Errorf("failed to create manifest path: %s", err)
	}
	if err := ioutil.WriteFile(manifestPath, mf, 0644); err != nil {
		return fmt.Errorf("failed to write manifest file: %s", err)
	}

	return nil
}

func (s *jsonapiCLI) setRegistryValue(path string, value string, globalInstall bool) error {
	key := registry.CURRENT_USER
	if globalInstall {
		key = registry.LOCAL_MACHINE
	}

	k, err := registry.OpenKey(key, path, registry.WRITE)
	if err != nil {
		if err != registry.ErrNotExist {
			return err
		}
		k, _, err = registry.CreateKey(key, path, registry.ALL_ACCESS)
		if err != nil {
			return err
		}
	}
	defer k.Close()
	return k.SetStringValue("", value)
}

func (s *jsonapiCLI) copyExecutionBinary(destFileName string) error {
	srcFileName, err := os.Executable()
	if err != nil {
		return err
	}

	srcFile, err := os.Open(srcFileName)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	destFile, err := os.Create(destFileName) // creates if file doesn't exist
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, srcFile)
	if err != nil {
		return err
	}
	return destFile.Sync()
}
