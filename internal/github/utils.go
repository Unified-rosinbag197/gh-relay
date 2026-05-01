package github

import (
	"fmt"
	"io"
	"net/http"
	"strings"
)

// ExtractRateLimit reads rate limit headers.
func ExtractRateLimit(resp *http.Response) RateLimitInfo {
	return RateLimitInfo{
		Limit:     resp.Header.Get("X-RateLimit-Limit"),
		Remaining: resp.Header.Get("X-RateLimit-Remaining"),
		Reset:     resp.Header.Get("X-RateLimit-Reset"),
	}
}

// ParseOwnerRepo splits "owner/repo" into its components.
func ParseOwnerRepo(ownerRepo string) (owner, repo string, err error) {
	parts := strings.SplitN(ownerRepo, "/", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("invalid repository format %q: expected \"owner/repo\"", ownerRepo)
	}
	return parts[0], parts[1], nil
}

// ReadAll is a helper to drain and close a response body.
func ReadAll(rc io.ReadCloser) ([]byte, error) {
	defer rc.Close()
	return io.ReadAll(rc)
}
