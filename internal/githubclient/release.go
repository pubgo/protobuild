package githubclient

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type RepositoryRelease struct {
	TagName string `json:"tag_name"`
	Assets  []struct {
		Name               string `json:"name"`
		BrowserDownloadURL string `json:"browser_download_url"`
		Size               int64  `json:"size"`
	} `json:"assets"`
}

type ReleaseClient struct {
	owner string
	repo  string
}

func NewPublicRelease(owner, repo string) *ReleaseClient {
	return &ReleaseClient{owner: owner, repo: repo}
}

func (c *ReleaseClient) Latest(ctx context.Context) (*RepositoryRelease, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest", c.owner, c.repo)
	return fetchRelease(ctx, url)
}

func (c *ReleaseClient) GetByTag(ctx context.Context, tag string) (*RepositoryRelease, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/tags/%s", c.owner, c.repo, tag)
	return fetchRelease(ctx, url)
}

func fetchRelease(ctx context.Context, url string) (*RepositoryRelease, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("GitHub API error: %s", resp.Status)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var release RepositoryRelease
	if err := json.Unmarshal(body, &release); err != nil {
		return nil, err
	}
	return &release, nil
}
