package updater

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/blang/semver/v4"
	"golang.org/x/net/context/ctxhttp"
)

var (
	// APITimeout is how long we wait for the GitHub API.
	APITimeout = 30 * time.Second

	// BaseURL is exported for tests.
	BaseURL    = "https://api.github.com/repos/%s/%s/releases/latest"
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
// from GitHub.
func FetchLatestRelease(ctx context.Context) (Release, error) {
	owner := gitHubOrg
	repo := gitHubRepo

	ctx, cancel := context.WithTimeout(ctx, APITimeout)
	defer cancel()

	url := fmt.Sprintf(BaseURL, owner, repo)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return Release{}, nil
	}

	// pin to API version 3 to avoid breaking our structs
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := ctxhttp.Do(ctx, http.DefaultClient, req)
	if err != nil {
		return Release{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return Release{}, fmt.Errorf("request faild with %v (%v)", resp.StatusCode, resp.Status)
	}

	var rs Release
	if err := json.NewDecoder(resp.Body).Decode(&rs); err != nil {
		return rs, err
	}

	if !strings.HasPrefix(rs.TagName, "v") {
		return rs, fmt.Errorf("tag name %q is invalid, must start with 'v'", rs.TagName)
	}

	v, err := semver.Parse(rs.TagName[1:])
	if err != nil {
		return rs, fmt.Errorf("failed to parse version %q: %q", rs.TagName[1:], err)
	}
	rs.Version = v

	return rs, nil
}
