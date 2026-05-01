package github

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

const githubAPIBase = "https://api.github.com"

func NewClient(token string) *Client {
	return &Client{
		token: token,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (c *Client) ValidateToken(ctx context.Context) error {
	resp, err := c.get(ctx, "/user")
	if err != nil {
		return fmt.Errorf("token validation request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return fmt.Errorf("invalid GitHub token: unauthorized")
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("token validation failed with status: %d", resp.StatusCode)
	}
	return nil
}

func (c *Client) GetRepo(ctx context.Context, owner, repo string) (*RepoInfo, error) {
	resp, err := c.get(ctx, fmt.Sprintf("/repos/%s/%s", owner, repo))
	if err != nil {
		return nil, fmt.Errorf("fetching repo info: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("repository %s/%s not found or not accessible with provided token", owner, repo)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fetching repo info failed with status: %d", resp.StatusCode)
	}

	var info RepoInfo
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return nil, fmt.Errorf("decoding repo info: %w", err)
	}
	return &info, nil
}

func (c *Client) GetTree(ctx context.Context, owner, repo, ref string) (*Tree, error) {
	// First resolve the ref to a commit SHA, then get the tree.
	commitSHA, err := c.resolveRef(ctx, owner, repo, ref)
	if err != nil {
		return nil, err
	}

	resp, err := c.get(ctx, fmt.Sprintf("/repos/%s/%s/git/trees/%s?recursive=1", owner, repo, commitSHA))
	if err != nil {
		return nil, fmt.Errorf("fetching tree: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fetching tree failed with status: %d", resp.StatusCode)
	}

	var tree Tree
	if err := json.NewDecoder(resp.Body).Decode(&tree); err != nil {
		return nil, fmt.Errorf("decoding tree: %w", err)
	}
	return &tree, nil
}

func (c *Client) GetBlob(ctx context.Context, owner, repo, sha string) ([]byte, error) {
	resp, err := c.get(ctx, fmt.Sprintf("/repos/%s/%s/git/blobs/%s", owner, repo, sha))
	if err != nil {
		return nil, fmt.Errorf("fetching blob: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("blob %s not found", sha)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fetching blob failed with status: %d", resp.StatusCode)
	}

	var blob Blob
	if err := json.NewDecoder(resp.Body).Decode(&blob); err != nil {
		return nil, fmt.Errorf("decoding blob: %w", err)
	}

	if blob.Encoding != "base64" {
		return nil, fmt.Errorf("unexpected blob encoding: %s", blob.Encoding)
	}

	clean := strings.ReplaceAll(blob.Content, "\n", "")
	data, err := base64.StdEncoding.DecodeString(clean)
	if err != nil {
		return nil, fmt.Errorf("decoding blob content: %w", err)
	}
	return data, nil
}

func (c *Client) GetCommits(ctx context.Context, owner, repo, branch string) ([]CommitInfo, error) {
	resp, err := c.get(ctx, fmt.Sprintf("/repos/%s/%s/commits?sha=%s&per_page=20", owner, repo, branch))
	if err != nil {
		return nil, fmt.Errorf("fetching commits: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fetching commits failed with status: %d", resp.StatusCode)
	}

	var commits []CommitInfo
	if err := json.NewDecoder(resp.Body).Decode(&commits); err != nil {
		return nil, fmt.Errorf("decoding commits: %w", err)
	}
	return commits, nil
}

func (c *Client) GetBranches(ctx context.Context, owner, repo string) ([]string, error) {
	resp, err := c.get(ctx, fmt.Sprintf("/repos/%s/%s/branches?per_page=100", owner, repo))
	if err != nil {
		return nil, fmt.Errorf("fetching branches: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fetching branches failed with status: %d", resp.StatusCode)
	}

	var raw []struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, fmt.Errorf("decoding branches: %w", err)
	}

	names := make([]string, len(raw))
	for i, b := range raw {
		names[i] = b.Name
	}
	return names, nil
}

func (c *Client) resolveRef(ctx context.Context, owner, repo, ref string) (string, error) {
	resp, err := c.get(ctx, fmt.Sprintf("/repos/%s/%s/commits/%s", owner, repo, ref))
	if err != nil {
		return "", fmt.Errorf("resolving ref %q: %w", ref, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return "", fmt.Errorf("branch/ref %q not found in %s/%s", ref, owner, repo)
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("resolving ref failed with status: %d", resp.StatusCode)
	}

	var commit struct {
		SHA string `json:"sha"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&commit); err != nil {
		return "", fmt.Errorf("decoding ref response: %w", err)
	}
	return commit.SHA, nil
}

func (c *Client) get(ctx context.Context, path string) (*http.Response, error) {
	url := githubAPIBase + path
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	req.Header.Set("User-Agent", "gh-relay/1.0")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	return resp, nil
}
