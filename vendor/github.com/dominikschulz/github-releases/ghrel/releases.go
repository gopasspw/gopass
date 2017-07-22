package ghrel

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/blang/semver"
)

var (
	BaseURL = "https://api.github.com/repos/%s/%s/releases"
	sem     = regexp.MustCompile(`(?:^|\D)(\d+\.\d+\.\d+\S*)(?:$|\s)`)
)

type Asset struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
	URL  string `json:"browser_download_url"`
}

type Release struct {
	Id          int       `json:"id"`
	Name        string    `json:"name"`
	TagName     string    `json:"tag_name"`
	Draft       bool      `json:"draft"`
	Prerelease  bool      `json:"prerelease"`
	PublishedAt time.Time `json:"published_at"`
	Assets      []Asset   `json:"assets"`
}

func (r Release) Version() semver.Version {
	match := sem.FindStringSubmatch(r.TagName)
	if len(match) < 2 {
		match = sem.FindStringSubmatch(r.Name)
	}
	if len(match) < 2 {
		return semver.Version{}
	}
	if sv, err := semver.ParseTolerant(match[1]); err == nil {
		return sv
	}
	return semver.Version{}
}

func fetchReleases(user, project string) ([]Release, error) {
	url := fmt.Sprintf(BaseURL, user, project)
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("Failed to fetch from %s: %s", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("Failed to fetch from %s: %d - %s", url, resp.StatusCode, resp.Status)
	}
	var rs []Release
	err = json.NewDecoder(resp.Body).Decode(&rs)
	if err != nil {
		return nil, err
	}

	return rs, nil
}

func findStableRelease(rs []Release) (Release, error) {
	for _, r := range rs {
		if strings.Contains(r.Name, "beta") || strings.Contains(r.Name, "rc") || r.Draft || r.Prerelease {
			continue
		}
		return r, nil
	}
	return Release{}, fmt.Errorf("No stable release found")
}

func FetchLatestStableRelease(user, project string) (Release, error) {
	rs, err := fetchReleases(user, project)
	if err != nil {
		return Release{}, err
	}
	if len(rs) < 1 {
		return Release{}, fmt.Errorf("No releases")
	}
	return findStableRelease(rs)
}
