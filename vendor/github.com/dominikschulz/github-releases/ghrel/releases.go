package ghrel

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

var apiURL = "https://api.github.com/repos/%s/%s/releases"

type Asset struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
	URL  string `json:"browser_download_url"`
}

type Release struct {
	Id          int       `json:"id"`
	Name        string    `json:"name"`
	PublishedAt time.Time `json:"published_at"`
	Assets      []Asset   `json:"assets"`
}

func FetchLatestStableRelease(user, project string) (Release, error) {
	url := fmt.Sprintf(apiURL, user, project)
	resp, err := http.Get(url)
	if err != nil {
		return Release{}, fmt.Errorf("Failed to fetch from %s: %s", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return Release{}, fmt.Errorf("Failed to fetch from %s: %d - %s", url, resp.StatusCode, resp.Status)
	}
	var rs []Release
	err = json.NewDecoder(resp.Body).Decode(&rs)
	if err != nil {
		return Release{}, err
	}
	if len(rs) < 1 {
		return Release{}, fmt.Errorf("No releases")
	}
	for _, r := range rs {
		if strings.Contains(r.Name, "beta") || strings.Contains(r.Name, "rc") {
			continue
		}
		return r, nil
	}
	return Release{}, fmt.Errorf("No stable release found")
}
