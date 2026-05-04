// Package server implements the gh-relay local HTTP server.
// It serves a file-browser SPA and proxies read-only GitHub API requests.
package server

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"strings"
	"time"
)

func New(cfg Config) *Server {
	s := &Server{cfg: cfg, mux: http.NewServeMux()}
	tok, _ := cfg.Sessions.Issue()
	rendered := []byte(strings.Replace(spaHTML, "/*__RELAY_TOKEN__*/", `var __RELAY_TOKEN__ = "`+tok+`";`, 1))
	s.renderedSPA = rendered
	s.token = tok
	s.registerRoutes()
	var handler http.Handler = s.mux
	if cfg.AuditLog != nil {
		handler = s.auditMiddleware(s.mux)
	}
	s.srv = &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Port),
		Handler:      handler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 60 * time.Second,
		IdleTimeout:  120 * time.Second,
	}
	return s
}

func (s *Server) Start(ctx context.Context) error {
	ln, err := net.Listen("tcp", s.srv.Addr)
	if err != nil {
		return fmt.Errorf("listening on %s: %w", s.srv.Addr, err)
	}

	errCh := make(chan error, 1)
	go func() {
		if err := s.srv.Serve(ln); err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
	}()

	select {
	case <-ctx.Done():
		shutCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		return s.srv.Shutdown(shutCtx)
	case err := <-errCh:
		return err
	}
}

func (s *Server) registerRoutes() {
	// All /api/* routes require a valid token in the X-Relay-Token header.
	s.mux.HandleFunc("/favicon.ico", http.NotFound)
	s.mux.HandleFunc("/v1/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	s.mux.HandleFunc("/api/info", s.requireToken(s.handleInfo))
	s.mux.HandleFunc("/api/tree", s.requireToken(s.handleTree))
	s.mux.HandleFunc("/api/blob", s.requireToken(s.handleBlob))
	s.mux.HandleFunc("/api/commits", s.requireToken(s.handleCommits))

	s.mux.HandleFunc("/", s.handleSPA)
}

func (s *Server) requireToken(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		tok := r.Header.Get("X-Relay-Token")
		if tok == "" || tok != s.token {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(`{"error":"invalid or expired session"}`))
			return
		}
		next(w, r)
	}
}

func (s *Server) auditMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasPrefix(r.URL.Path, "/api/") {
			next.ServeHTTP(w, r)
			return
		}

		rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rec, r)

		if rec.status == http.StatusUnauthorized {
			return
		}

		ip := getClientIP(r)
		filePath := r.URL.Query().Get("path")
		record := AuditRecord{
			timestamp: time.Now(),
			endpoint:  r.URL.Path,
			filePath:  filePath,
			branch:    r.URL.Query().Get("branch"),
			ipAddress: ip,
		}
		s.cfg.AuditLog.add(record)

		if filePath != "" {
			log.Printf("[audit] Guest viewed: %s (from %s)", filePath, ip)
		} else {
			log.Printf("[audit] %s %s (from %s)", r.Method, r.URL.Path, ip)
		}
	})
}

// handleInfo returns static repository metadata.
func (s *Server) handleInfo(w http.ResponseWriter, r *http.Request) {
	type infoResponse struct {
		Owner         string   `json:"owner"`
		Repo          string   `json:"repo"`
		Branch        string   `json:"branch"`
		DefaultBranch string   `json:"default_branch"`
		Description   string   `json:"description"`
		Private       bool     `json:"private"`
		Branches      []string `json:"branches"`
	}

	resp := infoResponse{
		Owner:         s.cfg.Owner,
		Repo:          s.cfg.Repo,
		Branch:        s.cfg.Branch,
		DefaultBranch: s.cfg.RepoInfo.DefaultBranch,
		Description:   s.cfg.RepoInfo.Description,
		Private:       s.cfg.RepoInfo.Private,
		Branches:      s.cfg.Branches,
	}
	writeJSON(w, resp)
}

// Query param: ?branch=<name>  (defaults to the configured branch)
func (s *Server) handleTree(w http.ResponseWriter, r *http.Request) {
	branch := r.URL.Query().Get("branch")
	if branch == "" {
		branch = s.cfg.Branch
	}
	if !isSafeBranchName(branch) {
		http.Error(w, "invalid branch name", http.StatusBadRequest)
		return
	}

	tree, err := s.treeForBranch(r.Context(), branch)
	if err != nil {
		log.Printf("[error] GetTree(%s): %v", branch, err)
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	writeJSON(w, s.filteredTree(tree))
}

// Query params: ?sha=<blob_sha>&path=<file_path>&branch=<branch>
func (s *Server) handleBlob(w http.ResponseWriter, r *http.Request) {
	sha := r.URL.Query().Get("sha")
	path := r.URL.Query().Get("path")
	branch := r.URL.Query().Get("branch")

	if sha == "" {
		http.Error(w, "missing sha parameter", http.StatusBadRequest)
		return
	}
	if !isSafeSHA(sha) {
		http.Error(w, "invalid sha parameter", http.StatusBadRequest)
		return
	}
	if s.pathFilterEnabled() {
		if path == "" {
			http.Error(w, "missing path parameter", http.StatusBadRequest)
			return
		}
		if branch == "" {
			http.Error(w, "missing branch parameter", http.StatusBadRequest)
			return
		}
		if !isSafeBranchName(branch) {
			http.Error(w, "invalid branch name", http.StatusBadRequest)
			return
		}
		if !s.cfg.PathFilter.AllowPath(path) {
			http.NotFound(w, r)
			return
		}

		tree, err := s.treeForBranch(r.Context(), branch)
		if err != nil {
			log.Printf("[error] GetTree(%s): %v", branch, err)
			http.Error(w, err.Error(), http.StatusBadGateway)
			return
		}
		if !treeHasBlob(tree, path, sha) {
			http.NotFound(w, r)
			return
		}
	}

	data, err := s.cfg.GitHub.GetBlob(r.Context(), s.cfg.Owner, s.cfg.Repo, sha)
	if err != nil {
		log.Printf("[error] GetBlob(%s): %v", sha, err)
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}

	ct := blobContentType(path)
	w.Header().Set("Content-Type", ct)
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(http.StatusOK)
	w.Write(data) //nolint:errcheck
}

// Query param: ?branch=<name>
func (s *Server) handleCommits(w http.ResponseWriter, r *http.Request) {
	branch := r.URL.Query().Get("branch")
	if branch == "" {
		branch = s.cfg.Branch
	}
	if !isSafeBranchName(branch) {
		http.Error(w, "invalid branch name", http.StatusBadRequest)
		return
	}

	commits, err := s.cfg.GitHub.GetCommits(r.Context(), s.cfg.Owner, s.cfg.Repo, branch)
	if err != nil {
		log.Printf("[error] GetCommits(%s): %v", branch, err)
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	writeJSON(w, commits)
}

func (s *Server) handleSPA(w http.ResponseWriter, r *http.Request) {
	log.Printf("[spa] %s %s", r.Method, r.URL.Path) // ← add this
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if r.URL.Path != "/" && r.URL.Path != "" {
		http.NotFound(w, r)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	w.Write(s.renderedSPA)
}
