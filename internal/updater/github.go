package updater

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/blang/semver/v4"
	"github.com/gopasspw/gopass/pkg/debug"
	"golang.org/x/net/context/ctxhttp"
)

var (
	// APITimeout is how long we wait for the GitHub API.
	APITimeout = 30 * time.Second

	// BaseURL is exported for tests.
	BaseURL = "https://api.github.com/repos/%s/%s/releases/latest"

	// PreReleaseBaseURL is exported for tests. It lists all releases,
	// including pre-releases, so that the highest semantic version can
	// be selected.
	PreReleaseBaseURL = "https://api.github.com/repos/%s/%s/releases?per_page=100"

	gitHubOrg  = "gopasspw"
	gitHubRepo = "gopass"
)

// Asset is a GitHub release asset.
type Asset struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	URL  string `json:"browser_download_url"`
}

// Release is a GitHub release.
type Release struct {
	ID          int            `json:"id"`
	Name        string         `json:"name"`
	TagName     string         `json:"tag_name"`
	Draft       bool           `json:"draft"`
	Prerelease  bool           `json:"prerelease"`
	PublishedAt time.Time      `json:"published_at"`
	Assets      []Asset        `json:"assets"`
	Version     semver.Version `json:"-"`
}

func downloadAsset(ctx context.Context, assets []Asset, suffix string) (string, []byte, error) {
	var url string

	var filename string

	for _, a := range assets {
		if !strings.HasSuffix(a.Name, suffix) {
			continue
		}

		url = a.URL
		filename = a.Name

		break
	}

	if url == "" {
		return "", nil, fmt.Errorf("asset with suffix %q not found", suffix)
	}

	buf, err := tryDownload(ctx, url)
	if err != nil {
		return "", nil, err
	}

	return filename, buf, nil
}

// FetchLatestRelease fetches meta-data about the latest Gopass release
// from GitHub. Pre-releases are ignored.
func FetchLatestRelease(ctx context.Context) (Release, error) {
	return fetchRelease(ctx, false)
}

// FetchLatestPrerelease fetches meta-data about the latest Gopass release
// from GitHub, including pre-releases such as release candidates.
func FetchLatestPrerelease(ctx context.Context) (Release, error) {
	return fetchRelease(ctx, true)
}

// fetchRelease fetches meta-data about the latest Gopass release from GitHub.
// If pre is true pre-releases are considered as well and the release with the
// highest semantic version wins.
func fetchRelease(ctx context.Context, pre bool) (Release, error) {
	owner := gitHubOrg
	repo := gitHubRepo

	ctx, cancel := context.WithTimeout(ctx, APITimeout)
	defer cancel()

	url := fmt.Sprintf(BaseURL, owner, repo)
	if pre {
		url = fmt.Sprintf(PreReleaseBaseURL, owner, repo)
	}

	body, err := get(ctx, url)
	if err != nil {
		return Release{}, err
	}

	if pre {
		var rels []Release
		if err := json.Unmarshal(body, &rels); err != nil {
			return Release{}, fmt.Errorf("failed to decode response: %w", err)
		}

		return newestRelease(rels)
	}

	var rs Release
	if err := json.Unmarshal(body, &rs); err != nil {
		return rs, fmt.Errorf("failed to decode response: %w", err)
	}

	return parseRelease(rs)
}

// newestRelease parses the versions of all given releases and returns the one
// with the highest semantic version, ignoring drafts and malformed tags.
func newestRelease(rels []Release) (Release, error) {
	var (
		newest Release
		found  bool
	)

	for _, rs := range rels {
		if rs.Draft {
			continue
		}

		rel, err := parseRelease(rs)
		if err != nil {
			debug.Log("skipping release: %s", err)

			continue
		}

		if !found || rel.Version.GT(newest.Version) {
			newest = rel
			found = true
		}
	}

	if !found {
		return Release{}, errors.New("no releases found")
	}

	return newest, nil
}

// parseRelease validates the tag name and populates the parsed version.
func parseRelease(rs Release) (Release, error) {
	if !strings.HasPrefix(rs.TagName, "v") {
		return rs, fmt.Errorf("tag name %q is invalid, must start with 'v'", rs.TagName)
	}

	v, err := semver.Parse(rs.TagName[1:])
	if err != nil {
		return rs, fmt.Errorf("failed to parse version %q: %w", rs.TagName[1:], err)
	}

	rs.Version = v

	return rs, nil
}

// get performs a GitHub API request and returns the response body.
func get(ctx context.Context, url string) ([]byte, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// pin to API version 3 to avoid breaking our structs
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := ctxhttp.Do(ctx, httpClient, req)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("request failed with %v (%v)", resp.StatusCode, resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	return body, nil
}
