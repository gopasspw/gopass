//go:build canary

package goeol

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.yaml.in/yaml/v3"
	"golang.org/x/mod/modfile"
)

// release is the subset of the endoflife.date release API response we care
// about. See https://endoflife.date/docs/api/v1/#/Products/product_release.
type release struct {
	Result struct {
		IsEOL   bool   `json:"isEol"`
		EOLFrom string `json:"eolFrom"`
	} `json:"result"`
}

// matrixGo is a list of Go versions as found in a GitHub Actions build matrix.
// It accepts both quoted strings and unquoted scalars (e.g. 1.26).
type matrixGo []string

// buildWorkflow is the subset of the build workflow we care about.
type buildWorkflow struct {
	Jobs map[string]struct {
		Strategy struct {
			Matrix struct {
				Go matrixGo `yaml:"go"`
			} `yaml:"matrix"`
		} `yaml:"strategy"`
	} `yaml:"jobs"`
}

const (
	goModPath    = "go.mod"
	buildYMLPath = ".github/workflows/build.yml"
	eolURL       = "https://endoflife.date/api/v1/products/go/releases/%s"
)

// UnmarshalYAML implements yaml.Unmarshaler.
func (m *matrixGo) UnmarshalYAML(value *yaml.Node) error {
	if value.Kind != yaml.SequenceNode {
		return fmt.Errorf("unexpected Go version node kind: %d", value.Kind)
	}

	for _, node := range value.Content {
		*m = append(*m, node.Value)
	}

	return nil
}

// majorMinor normalizes a Go version (e.g. "go1.26.0" or "1.26") to the
// major.minor form used by endoflife.date as the release key.
func majorMinor(version string) (string, error) {
	version = strings.TrimPrefix(version, "go")

	parts := strings.Split(version, ".")
	if len(parts) < 2 || parts[0] == "" || parts[1] == "" {
		return "", fmt.Errorf("unexpected Go version: %q", version)
	}

	return strings.Join(parts[:2], "."), nil
}

// releaseStatus fetches the endoflife.date release metadata for a major.minor
// Go version.
func releaseStatus(ctx context.Context, version string) (release, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf(eolURL, version), nil)
	if err != nil {
		return release{}, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return release{}, err
	}
	defer resp.Body.Close() //nolint:errcheck

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return release{}, err
	}

	if resp.StatusCode != http.StatusOK {
		return release{}, fmt.Errorf("unexpected response for Go %s: %s: %s", version, resp.Status, string(body))
	}

	var rel release
	if err := json.Unmarshal(body, &rel); err != nil {
		return release{}, err
	}

	return rel, nil
}

// eolContext returns a context with a timeout for the API requests, canceled
// once the test finishes.
func eolContext(t *testing.T) context.Context {
	t.Helper()

	ctx, cancel := context.WithTimeout(t.Context(), 30*time.Second)
	t.Cleanup(cancel)

	return ctx
}

// assertNotEOL fails the test if the given major.minor Go version has reached
// its end of life.
func assertNotEOL(t *testing.T, ctx context.Context, version string) {
	t.Helper()

	rel, err := releaseStatus(ctx, version)
	require.NoError(t, err)

	require.Falsef(t, rel.Result.IsEOL,
		"Go %s has been EOL since %s. Please update the Go version in go.mod and the GitHub Action workflows to a supported release.",
		version, rel.Result.EOLFrom)
}

// goModVersion returns the major.minor Go version targeted by our go.mod.
func goModVersion(t *testing.T) string {
	t.Helper()

	buf, err := os.ReadFile(filepath.Join("..", "..", filepath.FromSlash(goModPath)))
	require.NoError(t, err)

	mod, err := modfile.Parse(goModPath, buf, nil)
	require.NoError(t, err)
	require.NotNil(t, mod.Go, "go.mod has no go directive")

	version, err := majorMinor(mod.Go.Version)
	require.NoError(t, err)

	return version
}

// buildMatrixVersions returns the deduplicated major.minor Go versions used in
// the build matrix.
func buildMatrixVersions(t *testing.T) []string {
	t.Helper()

	buf, err := os.ReadFile(filepath.Join("..", "..", filepath.FromSlash(buildYMLPath)))
	require.NoError(t, err)

	var wf buildWorkflow
	require.NoError(t, yaml.Unmarshal(buf, &wf))

	seen := map[string]bool{}
	versions := make([]string, 0, len(wf.Jobs))

	for _, job := range wf.Jobs {
		for _, raw := range job.Strategy.Matrix.Go {
			version, err := majorMinor(raw)
			require.NoError(t, err)

			if seen[version] {
				continue
			}
			seen[version] = true

			versions = append(versions, version)
		}
	}

	slices.Sort(versions)

	return versions
}

// TestGoVersionNotEOL fails if the Go version targeted by our go.mod has
// reached its end of life. This is supposed to act as a canary so we don't
// forget to update the Go version before it goes out of support. See
// ../../.github/workflows/go-eol-canary.yml.
func TestGoVersionNotEOL(t *testing.T) {
	t.Parallel()

	assertNotEOL(t, eolContext(t), goModVersion(t))
}

// TestBuildMatrixNotEOL fails if any of the Go versions used in the build
// matrix of .github/workflows/build.yml has reached its end of life.
func TestBuildMatrixNotEOL(t *testing.T) {
	t.Parallel()

	versions := buildMatrixVersions(t)
	require.NotEmpty(t, versions, "no Go versions found in the build matrix of %s", buildYMLPath)

	for _, version := range versions {
		t.Run("go"+version, func(t *testing.T) {
			t.Parallel()

			assertNotEOL(t, eolContext(t), version)
		})
	}
}
