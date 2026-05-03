package server

import (
	"net/http"

	"github.ibm.com/soub4i/gh-relay/internal/github"
	"github.ibm.com/soub4i/gh-relay/internal/session"
)

// Config holds everything the server needs to operate.
type Config struct {
	Owner         string
	Repo          string
	Branch        string
	RepoInfo      *github.RepoInfo
	Branches      []string
	GitHub        *github.Client
	Sessions      *session.Manager
	Port          int
	AuditLog      *AuditLog
	AllowDownload bool
}

// Server is the gh-relay HTTP server.
type Server struct {
	cfg         Config
	mux         *http.ServeMux
	srv         *http.Server
	token       string
	renderedSPA []byte
}
