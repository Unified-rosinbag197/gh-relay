package server

import (
	"context"
	"net/http"

	"github.ibm.com/soub4i/gh-relay/internal/filter"
	"github.ibm.com/soub4i/gh-relay/internal/github"
	"github.ibm.com/soub4i/gh-relay/internal/session"
)

type gitHubAPI interface {
	GetTree(ctx context.Context, owner, repo, ref string) (*github.Tree, error)
	GetBlob(ctx context.Context, owner, repo, sha string) ([]byte, error)
	GetCommits(ctx context.Context, owner, repo, branch string) ([]github.CommitInfo, error)
}

// Config holds everything the server needs to operate.
type Config struct {
	Owner      string
	Repo       string
	Branch     string
	RepoInfo   *github.RepoInfo
	Branches   []string
	GitHub     gitHubAPI
	Sessions   *session.Manager
	Port       int
	Tree       *github.Tree
	AuditLog   *AuditLog
	PathFilter *filter.Policy
}

// Server is the gh-relay HTTP server.
type Server struct {
	cfg         Config
	mux         *http.ServeMux
	srv         *http.Server
	token       string
	renderedSPA []byte
}
