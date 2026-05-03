package server

import (
	"context"
	"path"

	"github.ibm.com/soub4i/gh-relay/internal/github"
)

func (s *Server) pathFilterEnabled() bool {
	return s.cfg.PathFilter != nil && s.cfg.PathFilter.Enabled()
}

func (s *Server) treeForBranch(ctx context.Context, branch string) (*github.Tree, error) {
	if branch == s.cfg.Branch && s.cfg.Tree != nil {
		return s.cfg.Tree, nil
	}
	return s.cfg.GitHub.GetTree(ctx, s.cfg.Owner, s.cfg.Repo, branch)
}

func (s *Server) filteredTree(tree *github.Tree) *github.Tree {
	if !s.pathFilterEnabled() || tree == nil {
		return tree
	}

	visiblePaths := make(map[string]struct{}, len(tree.Tree))
	for _, entry := range tree.Tree {
		if !s.cfg.PathFilter.AllowPath(entry.Path) {
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

func treeHasBlob(tree *github.Tree, repoPath, sha string) bool {
	if tree == nil {
		return false
	}
	for _, entry := range tree.Tree {
		if entry.Path == repoPath && entry.Type == "blob" && entry.SHA == sha {
			return true
		}
	}
	return false
}
