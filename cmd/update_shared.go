package cmd

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

const githubReleaseURL = "https://api.github.com/repos/alex-vit/plop/releases/latest"

type ghRelease struct {
	TagName string    `json:"tag_name"`
	Assets  []ghAsset `json:"assets"`
}

type ghAsset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
}

// fetchLatestRelease queries the GitHub releases API and returns the latest release.
func fetchLatestRelease() (*ghRelease, error) {
	req, err := http.NewRequest(http.MethodGet, githubReleaseURL, nil) //nolint:noctx
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned %d", resp.StatusCode)
	}

	var rel ghRelease
	if err := json.NewDecoder(resp.Body).Decode(&rel); err != nil {
		return nil, err
	}
	return &rel, nil
}

// isNewer reports whether latest is a higher semver than current.
// Versions are expected as "X.Y.Z" (no "v" prefix).
func isNewer(latest, current string) bool {
	if current == "" || current == "dev" {
		return false // dev builds don't auto-update
	}
	lp := parseSemver(latest)
	cp := parseSemver(current)
	if lp == nil || cp == nil {
		return false
	}
	for i := range 3 {
		if lp[i] > cp[i] {
			return true
		}
		if lp[i] < cp[i] {
			return false
		}
	}
	return false
}

func parseSemver(s string) []int {
	parts := strings.SplitN(s, ".", 3)
	if len(parts) != 3 {
		return nil
	}
	nums := make([]int, 3)
	for i, p := range parts {
		n, err := strconv.Atoi(p)
		if err != nil {
			return nil
		}
		nums[i] = n
	}
	return nums
}

func versionDisplay() string {
	if Version == "" {
		return "dev"
	}
	return Version
}
