package main

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/blang/semver/v4"
)

var (
	wixTpl = `<Wix xmlns="http://schemas.microsoft.com/wix/2006/wi">
  <Product Id="*" Name="gopass" UpgradeCode="{{ .UpgradeCode }}" Language="1033" Codepage="1252" Version="{{ .Version }}" Manufacturer="gopass">
    <Property Id="PREVIOUSVERSIONSINSTALLED" Secure="yes"/>
    <Upgrade Id="{{ .UpgradeCode }}">
      <UpgradeVersion Minimum="0.0.0" Property="PREVIOUSVERSIONSINSTALLED" IncludeMinimum="yes" IncludeMaximum="no"/>
    </Upgrade>
    <InstallExecuteSequence>
      <RemoveExistingProducts Before="InstallInitialize"/>
    </InstallExecuteSequence>
    <Package InstallerVersion="200" Compressed="yes" Comments="Windows Installer Package" InstallScope="perUser"/>
    <Media Id="1" Cabinet="app.cab" EmbedCab="yes"/>
    <Icon Id="icon.ico" SourceFile="{{ .Icon }}"/>
    <Property Id="ARPPRODUCTICON" Value="icon.ico"/>
    <Directory Id="TARGETDIR" Name="SourceDir">
      <Directory Id="LocalAppDataFolder">
        <Directory Id="INSTALLDIR" Name="gopass">
          <Component Id="gopass.exe" Guid="*">
            <File Id="gopass.exe" Source="{{ .Binary }}" Name="gopass.exe"/>
            <Shortcut Id="StartMenuShortcut" Advertise="no" Icon="icon.ico" Name="gopass" Directory="ProgramMenuFolder" WorkingDirectory="INSTALLDIR" Description=""/>
            <Shortcut Id="DesktopShortcut" Advertise="no" Icon="icon.ico" Name="gopass" Directory="DesktopFolder" WorkingDirectory="INSTALLDIR" Description=""/>
          </Component>
        </Directory>
      </Directory>
    </Directory>
    <Feature Id="App" Level="1">
      <ComponentRef Id="gopass.exe"/>
    </Feature>
  </Product>
</Wix>
`
	upgradeCode = "6c1bd458-7d1b-4311-848d-d0fe1a65af66"
	icon        = "docs/logo.ico"
	source      = "dist/gopass_windows_amd64_v1/gopass.exe"
)

const logo = `
   __     _    _ _      _ _   ___   ___
 /'_ '\ /'_'\ ( '_'\  /'_' )/',__)/',__)
( (_) |( (_) )| (_) )( (_| |\__, \\__, \
'\__  |'\___/'| ,__/''\__,_)(____/(____/
( )_) |       | |
 \___/'       (_)
`

func main() {
	// render template to temp dir
	// run: wixl /tmp/file.xml -o gopass-ARCH.msi --arch x86|x64
	ctx := context.Background()

	fmt.Print(logo)
	fmt.Println()
	fmt.Println("🌟 Creating gopass Windows MSI package.")

	curVer, err := versionFile()
	if err != nil {
		panic(err)
	}

	fmt.Println()
	fmt.Printf("✅ Current version is: %s\n", curVer.String())

	td, err := os.MkdirTemp("", "gopass-")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(td)

	tmpl, err := template.New("wix").Parse(wixTpl)
	if err != nil {
		panic(err)
	}

	wCfg := filepath.Join(td, "wix.xml")
	fh, err := os.Create(wCfg)
	if err != nil {
		panic(err)
	}
	if err := tmpl.Execute(fh, struct {
		UpgradeCode string
		Version     string
		Icon        string
		Binary      string
	}{
		UpgradeCode: upgradeCode,
		Version:     curVer.String(),
		Icon:        icon,
		Binary:      source,
	}); err != nil {
		panic(err)
	}
	fh.Close()

	fmt.Printf("✅ Wrote Wix XML config to: %s\n", wCfg)

	msiPkg := fmt.Sprintf("dist/gopass-%s-windows-%s.msi", "x64", curVer.String())
	cmd := exec.CommandContext(ctx, "wixl", wCfg, "-o", msiPkg, "--arch", "x64")
	buf := &bytes.Buffer{}
	cmd.Stdout = buf
	cmd.Stderr = buf
	if err := cmd.Run(); err != nil {
		fmt.Printf("wixl failed: %s", buf.String())
		panic(err)
	}

	fmt.Printf("✅ Created MSI package at: %s\n", msiPkg)
	fmt.Println()
}

func versionFile() (semver.Version, error) {
	buf, err := os.ReadFile("VERSION")
	if err != nil {
		return semver.Version{}, err
	}

	return semver.Parse(strings.TrimSpace(string(buf)))
}
