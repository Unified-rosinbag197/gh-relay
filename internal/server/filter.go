package server

import (
	"context"

	"github.ibm.com/soub4i/gh-relay/internal/filter"
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
	return filter.FilterTree(tree, s.cfg.PathFilter)
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
