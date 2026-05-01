package server

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"strings"
	"time"

	"github.ibm.com/soub4i/gh-relay/internal/session"
)

func New(cfg Config) *Server {
	s := &Server{cfg: cfg, mux: http.NewServeMux()}
	s.registerRoutes()
	s.srv = &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Port),
		Handler:      s.mux,
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
	// Auth.
	s.mux.HandleFunc("/auth/login", s.handleLogin)

	// All /api/* routes require a valid session.
	s.mux.HandleFunc("/api/info", s.auth(s.handleInfo))
	s.mux.HandleFunc("/api/tree", s.auth(s.handleTree))
	s.mux.HandleFunc("/api/blob", s.auth(s.handleBlob))
	s.mux.HandleFunc("/api/commits", s.auth(s.handleCommits))

	// Serve the SPA for everything else.
	s.mux.HandleFunc("/", s.auth(s.handleSPA))
}

// TODO: add rate limiting to prevent abuse and DoS attacks.
// TODO: add a /auth/logout endpoint to allow manual session revocation.
// TODO: add one-time passwords or other mechanisms to increase security if needed.

// auth is middleware that checks for a valid session cookie and redirects to /auth/login if not present.
func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if cookie, err := r.Cookie(session.CookieName()); err == nil && s.cfg.Sessions.Valid(cookie.Value) {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}
	token, err := s.cfg.Sessions.Issue()
	if err != nil {
		http.Error(w, "could not create session", http.StatusInternalServerError)
		log.Printf("[error] issuing session: %v", err)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     session.CookieName(),
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	})

	log.Printf("[session] new session issued; active sessions: %d", s.cfg.Sessions.Count())
	http.Redirect(w, r, "/", http.StatusFound)
}

func (s *Server) auth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		cookie, err := r.Cookie(session.CookieName())
		if err != nil || !s.cfg.Sessions.Valid(cookie.Value) {
			if strings.HasPrefix(r.URL.Path, "/api/") {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				w.Write([]byte(`{"error":"session expired, please refresh"}`))
				return
			}
			http.Redirect(w, r, "/auth/login", http.StatusFound)
			return
		}
		next(w, r)
	}
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

// handleTree returns the recursive file tree for a branch.
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

	tree, err := s.cfg.GitHub.GetTree(r.Context(), s.cfg.Owner, s.cfg.Repo, branch)
	if err != nil {
		log.Printf("[error] GetTree(%s): %v", branch, err)
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	writeJSON(w, tree)
}

// Query params: ?sha=<blob_sha>&path=<file_path>
// path is used only to set the Content-Type correctly.
func (s *Server) handleBlob(w http.ResponseWriter, r *http.Request) {
	sha := r.URL.Query().Get("sha")
	path := r.URL.Query().Get("path")

	if sha == "" {
		http.Error(w, "missing sha parameter", http.StatusBadRequest)
		return
	}
	if !isSafeSHA(sha) {
		http.Error(w, "invalid sha parameter", http.StatusBadRequest)
		return
	}

	data, err := s.cfg.GitHub.GetBlob(r.Context(), s.cfg.Owner, s.cfg.Repo, sha)
	if err != nil {
		log.Printf("[error] GetBlob(%s): %v", sha, err)
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}

	ct := mimeForPath(path)
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
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(spaHTML)) //nolint:errcheck
}
