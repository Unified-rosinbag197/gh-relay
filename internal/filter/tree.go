package filter

import (
	"path"

	"github.ibm.com/soub4i/gh-relay/internal/github"
)

// FilterTree returns the repository tree visible under policy.
func FilterTree(tree *github.Tree, policy *Policy) *github.Tree {
	if policy == nil || !policy.Enabled() || tree == nil {
		return tree
	}

	visiblePaths := make(map[string]struct{}, len(tree.Tree))
	for _, entry := range tree.Tree {
		if !policy.AllowPath(entry.Path) {
			continue
		}
		visiblePaths[entry.Path] = struct{}{}
		for dir := path.Dir(entry.Path); dir != "." && dir != "/"; dir = path.Dir(dir) {
			visiblePaths[dir] = struct{}{}
		}
	}

	filtered := *tree
	filtered.Tree = make([]github.TreeEntry, 0, len(tree.Tree))
	for _, entry := range tree.Tree {
		if _, ok := visiblePaths[entry.Path]; ok {
			filtered.Tree = append(filtered.Tree, entry)
		}
	}
	return &filtered
}
