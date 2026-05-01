package github

import "net/http"

type Client struct {
	token      string
	httpClient *http.Client
}

// TreeEntry represents a single node in a repository's file tree.
type TreeEntry struct {
	Path string `json:"path"`
	Mode string `json:"mode"`
	Type string `json:"type"` // "blob" or "tree"
	SHA  string `json:"sha"`
	Size int    `json:"size"`
	URL  string `json:"url"`
}

// Tree represents a GitHub git tree response.
type Tree struct {
	SHA       string      `json:"sha"`
	URL       string      `json:"url"`
	Tree      []TreeEntry `json:"tree"`
	Truncated bool        `json:"truncated"`
}

// Blob represents a GitHub git blob (file content).
type Blob struct {
	Content  string `json:"content"`
	Encoding string `json:"encoding"`
	SHA      string `json:"sha"`
	Size     int    `json:"size"`
}

// RepoInfo holds basic repository metadata.
type RepoInfo struct {
	FullName      string `json:"full_name"`
	DefaultBranch string `json:"default_branch"`
	Private       bool   `json:"private"`
	Description   string `json:"description"`
}

// CommitInfo holds minimal commit metadata.
type CommitInfo struct {
	SHA    string `json:"sha"`
	Commit struct {
		Message string `json:"message"`
		Author  struct {
			Name string `json:"name"`
			Date string `json:"date"`
		} `json:"author"`
	} `json:"commit"`
}

// RateLimitInfo reads the X-RateLimit headers from any GitHub response.
type RateLimitInfo struct {
	Limit     string
	Remaining string
	Reset     string
}
