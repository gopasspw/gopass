package updater

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFetchLatestRelease(t *testing.T) {
	tests := []struct {
		name           string
		responseBody   string
		responseStatus int
		expectedError  bool
		expectedTag    string
	}{
		{
			name: "successful fetch",
			responseBody: `{
				"id": 1,
				"name": "v1.0.0",
				"tag_name": "v1.0.0",
				"draft": false,
				"prerelease": false,
				"published_at": "2021-01-01T00:00:00Z",
				"assets": []
			}`,
			responseStatus: http.StatusOK,
			expectedError:  false,
			expectedTag:    "v1.0.0",
		},
		{
			name:           "invalid status code",
			responseBody:   "",
			responseStatus: http.StatusInternalServerError,
			expectedError:  true,
		},
		{
			name: "invalid tag name",
			responseBody: `{
				"id": 1,
				"name": "1.0.0",
				"tag_name": "1.0.0",
				"draft": false,
				"prerelease": false,
				"published_at": "2021-01-01T00:00:00Z",
				"assets": []
			}`,
			responseStatus: http.StatusOK,
			expectedError:  true,
		},
		{
			name: "invalid JSON",
			responseBody: `{
				"id": 1,
				"name": "v1.0.0",
				"tag_name": "v1.0.0",
				"draft": false,
				"prerelease": false,
				"published_at": "2021-01-01T00:00:00Z",
				"assets": [`,
			responseStatus: http.StatusOK,
			expectedError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.responseStatus)
				_, _ = w.Write([]byte(tt.responseBody))
			}))
			defer server.Close()

			BaseURL = server.URL + "/repos/%s/%s/releases/latest"

			ctx := t.Context()
			release, err := FetchLatestRelease(ctx)

			if tt.expectedError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectedTag, release.TagName)
			}
		})
	}
}

func TestDownloadAsset(t *testing.T) {
	tests := []struct {
		name          string
		assets        []Asset
		suffix        string
		expectedError bool
		expectedName  string
	}{
		{
			name: "successful download",
			assets: []Asset{
				{Name: "asset1.txt", URL: "/asset1.txt"},
				{Name: "asset2.txt", URL: "/asset2.txt"},
			},
			suffix:        ".txt",
			expectedError: false,
			expectedName:  "asset1.txt",
		},
		{
			name: "asset not found",
			assets: []Asset{
				{Name: "asset1.bin", URL: "/asset1.bin"},
			},
			suffix:        ".txt",
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte("test content"))
			}))
			defer server.Close()

			for i, a := range tt.assets {
				tt.assets[i].URL = server.URL + a.URL
			}

			ctx := t.Context()

			name, _, err := downloadAsset(ctx, tt.assets, tt.suffix)

			if tt.expectedError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectedName, name)
			}
		})
	}
}
