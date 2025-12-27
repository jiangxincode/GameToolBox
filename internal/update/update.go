package update

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"
)

type ReleaseInfo struct {
	TagName string `json:"tag_name"`
	HTMLURL string `json:"html_url"`
}

type HTTPStatusError struct {
	StatusCode int
	Status     string
}

func (e HTTPStatusError) Error() string {
	return fmt.Sprintf("github api status: %s", e.Status)
}

func isNotFound(err error) bool {
	var se HTTPStatusError
	if errors.As(err, &se) {
		return se.StatusCode == http.StatusNotFound
	}
	return false
}

// LatestRelease queries GitHub API for the latest release/tag of owner/repo.
//
// Strategy:
//  1. Try /releases/latest (requires GitHub "Releases" to exist)
//  2. If 404, fall back to /tags and use the first tag as "latest".
func LatestRelease(ctx context.Context, owner, repo string) (ReleaseInfo, error) {
	if info, err := latestFromReleases(ctx, owner, repo); err == nil {
		return info, nil
	} else if !isNotFound(err) {
		return ReleaseInfo{}, err
	}

	// No releases; fall back to tags.
	return latestFromTags(ctx, owner, repo)
}

func latestFromReleases(ctx context.Context, owner, repo string) (ReleaseInfo, error) {
	ctx, cancel := context.WithTimeout(ctx, 8*time.Second)
	defer cancel()

	apiURL := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest", owner, repo)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
	if err != nil {
		return ReleaseInfo{}, err
	}
	req.Header.Set("User-Agent", "GameToolBox")
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return ReleaseInfo{}, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return ReleaseInfo{}, HTTPStatusError{StatusCode: resp.StatusCode, Status: resp.Status}
	}

	var info ReleaseInfo
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return ReleaseInfo{}, err
	}
	info.TagName = strings.TrimSpace(info.TagName)
	info.HTMLURL = strings.TrimSpace(info.HTMLURL)
	if info.TagName == "" {
		return ReleaseInfo{}, fmt.Errorf("latest release has empty tag_name")
	}
	return info, nil
}

type tagItem struct {
	Name string `json:"name"`
}

func latestFromTags(ctx context.Context, owner, repo string) (ReleaseInfo, error) {
	ctx, cancel := context.WithTimeout(ctx, 8*time.Second)
	defer cancel()

	apiURL := fmt.Sprintf("https://api.github.com/repos/%s/%s/tags?per_page=1", owner, repo)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
	if err != nil {
		return ReleaseInfo{}, err
	}
	req.Header.Set("User-Agent", "GameToolBox")
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return ReleaseInfo{}, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return ReleaseInfo{}, HTTPStatusError{StatusCode: resp.StatusCode, Status: resp.Status}
	}

	var tags []tagItem
	if err := json.NewDecoder(resp.Body).Decode(&tags); err != nil {
		return ReleaseInfo{}, err
	}
	if len(tags) == 0 {
		return ReleaseInfo{}, fmt.Errorf("no tags found")
	}
	name := strings.TrimSpace(tags[0].Name)
	if name == "" {
		return ReleaseInfo{}, fmt.Errorf("latest tag has empty name")
	}

	// Link to tag list (or releases page if desired).
	return ReleaseInfo{TagName: name, HTMLURL: fmt.Sprintf("https://github.com/%s/%s/tags", owner, repo)}, nil
}
