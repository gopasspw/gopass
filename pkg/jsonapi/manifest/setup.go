package manifest

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"

	homedir "github.com/mitchellh/go-homedir"
)

type configuredFile struct {
	path    string
	content []byte
}

// PrintSummary prints path and content of manifest and wrapper script
func PrintSummary(browser, wrapperPath, libpath string, global bool) error {
	manifestFile, err := getManifest(browser, wrapperPath, libpath, global)
	if err != nil {
		return err
	}

	printConfiguredFile("Native Messaging Host Manifest", manifestFile)

	wrapperFile, err := getWrapper(wrapperPath)
	if err != nil {
		return err
	}
	printConfiguredFile("Wrapper", wrapperFile)

	return nil
}

// SetUp actually creates the manifest and wrapper scripts
func SetUp(browser, wrapperPath, libpath string, global bool) error {
	manifestFile, err := getManifest(browser, wrapperPath, libpath, global)
	if err != nil {
		return err
	}

	if err := writeConfiguredFile("Manifest", manifestFile, 0644); err != nil {
		return err
	}

	wrapperFile, err := getWrapper(wrapperPath)
	if err != nil {
		return err
	}

	return writeConfiguredFile("Wrapper", wrapperFile, 0755)
}

func printConfiguredFile(preamble string, file configuredFile) {
	fmt.Println(preamble)
	fmt.Printf("\npath: %s\n", file.path)
	fmt.Printf("\n### File content: ###\n%s\n###\n\n", file.content)
}

func writeConfiguredFile(name string, file configuredFile, perm os.FileMode) error {
	dir := filepath.Dir(file.path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	if err := ioutil.WriteFile(file.path, []byte(file.content), perm); err != nil {
		return err
	}

	fmt.Printf("\n%s written to %s\n", name, file.path)
	return nil
}

func getManifest(browser, wrapperPath, libpath string, global bool) (configuredFile, error) {
	file := configuredFile{}
	manifestPath, err := getManifestPath(browser, libpath, global)
	if err != nil {
		return file, err
	}

	file.path = manifestPath
	file.content, err = getManifestContent(browser, wrapperPath)
	return file, err
}

func getWrapper(wrapperPath string) (configuredFile, error) {
	file := configuredFile{
		path: path.Join(wrapperPath, wrapperName),
	}

	gopassPath, err := getGopassPath()
	if err != nil {
		return file, err
	}

	file.content = getWrapperContent(gopassPath)
	return file, nil
}

func getManifestPath(browser, libpath string, globalInstall bool) (string, error) {
	location, err := getLocation(browser, libpath, globalInstall)
	if err != nil {
		return "", err
	}

	expanded, err := homedir.Expand(location)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf(expanded, name), nil
}

func getGopassPath() (string, error) {
	return os.Executable()
}

func getWrapperContent(gopassPath string) []byte {
	return []byte(fmt.Sprintf(wrapperTemplate, gopassPath))
}

func getManifestContent(browser, wrapperPath string) ([]byte, error) {
	switch browser {
	case "firefox":
		jsonManifest := firefoxManifest{}
		jsonManifest.InitFields(path.Join(wrapperPath, wrapperName))
		jsonManifest.AllowedExtensions = firefoxOrigins
		return json.MarshalIndent(jsonManifest, "", "    ")
	case "chrome":
		fallthrough
	case "chromium":
		jsonManifest := chromeManifest{}
		jsonManifest.InitFields(path.Join(wrapperPath, wrapperName))
		jsonManifest.AllowedOrigins = chromeOrigins
		return json.MarshalIndent(jsonManifest, "", "    ")
	default:
		return nil, fmt.Errorf("no manifest template for browser %s", browser)
	}
}

func (m *manifestBase) InitFields(wrapperPath string) {
	m.Name = name
	m.Type = connectionType
	m.Path = wrapperPath
	m.Description = description
}
