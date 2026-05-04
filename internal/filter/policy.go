// Package filter implements repository-relative path allow/deny policies.
package filter

import (
	"fmt"
	"path"
	"strings"
)

// Policy contains normalized allow and deny path patterns.
type Policy struct {
	Allow []string
	Deny  []string
}

// NewPolicy parses comma-separated allow and deny patterns.
func NewPolicy(allowCSV, denyCSV string) (*Policy, error) {
	allow, err := parsePatterns("allow", allowCSV)
	if err != nil {
		return nil, err
	}
	deny, err := parsePatterns("deny", denyCSV)
	if err != nil {
		return nil, err
	}
	return &Policy{Allow: allow, Deny: deny}, nil
}

// Enabled reports whether the policy has any allow or deny rule.
func (p *Policy) Enabled() bool {
	return p != nil && (len(p.Allow) > 0 || len(p.Deny) > 0)
}

// AllowPath reports whether a repository-relative path is visible.
func (p *Policy) AllowPath(repoPath string) bool {
	if p == nil || !p.Enabled() {
		return true
	}

	clean, ok := normalizePath(repoPath)
	if !ok {
		return false
	}

	for _, candidate := range selfAndAncestors(clean) {
		for _, pattern := range p.Deny {
			if matchPattern(pattern, candidate) {
				return false
			}
		}
	}

	if len(p.Allow) == 0 {
		return true
	}
	for _, pattern := range p.Allow {
		if matchPattern(pattern, clean) {
			return true
		}
	}
	return false
}

func parsePatterns(kind, csv string) ([]string, error) {
	if strings.TrimSpace(csv) == "" {
		return nil, nil
	}

	var patterns []string
	for _, raw := range strings.Split(csv, ",") {
		pattern, ok := normalizePattern(raw)
		if !ok {
			if strings.TrimSpace(raw) == "" {
				continue
			}
			return nil, fmt.Errorf("invalid %s pattern %q: path traversal is not allowed", kind, strings.TrimSpace(raw))
		}
		if pattern == "" {
			continue
		}
		if _, err := path.Match(pattern, ""); err != nil {
			return nil, fmt.Errorf("invalid %s pattern %q: %w", kind, pattern, err)
		}
		patterns = append(patterns, pattern)
	}
	return patterns, nil
}

func normalizePattern(pattern string) (string, bool) {
	pattern = strings.TrimSpace(pattern)
	if pattern == "" {
		return "", true
	}
	pattern = strings.ReplaceAll(pattern, `\`, "/")
	pattern = strings.TrimLeft(pattern, "/")
	if pattern == "" {
		return "", true
	}
	if hasTraversalSegment(pattern) {
		return "", false
	}

	clean := path.Clean(pattern)
	if clean == "." || clean == ".." || strings.HasPrefix(clean, "../") {
		return "", false
	}
	return clean, true
}

func normalizePath(repoPath string) (string, bool) {
	repoPath = strings.TrimSpace(repoPath)
	if repoPath == "" {
		return "", false
	}
	repoPath = strings.ReplaceAll(repoPath, `\`, "/")
	repoPath = strings.TrimLeft(repoPath, "/")
	if repoPath == "" || hasTraversalSegment(repoPath) {
		return "", false
	}

	clean := path.Clean(repoPath)
	if clean == "." || clean == ".." || strings.HasPrefix(clean, "../") {
		return "", false
	}
	return clean, true
}

func hasTraversalSegment(repoPath string) bool {
	for _, segment := range strings.Split(repoPath, "/") {
		if segment == ".." {
			return true
		}
	}
	return false
}

func matchPattern(pattern, repoPath string) bool {
	if strings.HasSuffix(pattern, "/**") {
		base := strings.TrimSuffix(pattern, "/**")
		return repoPath == base || strings.HasPrefix(repoPath, base+"/")
	}

	target := repoPath
	if !strings.Contains(pattern, "/") {
		target = path.Base(repoPath)
	}
	ok, err := path.Match(pattern, target)
	return err == nil && ok
}

func selfAndAncestors(repoPath string) []string {
	paths := []string{repoPath}
	for dir := path.Dir(repoPath); dir != "." && dir != "/"; dir = path.Dir(dir) {
		paths = append(paths, dir)
	}
	return paths
}
