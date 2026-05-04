package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.ibm.com/soub4i/gh-relay/internal/filter"
	"github.ibm.com/soub4i/gh-relay/internal/github"
	"github.ibm.com/soub4i/gh-relay/internal/session"
)

const (
	shaMain   = "1111111111111111111111111111111111111111"
	shaEnv    = "2222222222222222222222222222222222222222"
	shaDocs   = "3333333333333333333333333333333333333333"
	shaSecret = "4444444444444444444444444444444444444444"
)

func TestTreeExcludesDeniedFiles(t *testing.T) {
	tree := testTree()
	srv, _ := newTestServer(t, mustFilter(t, "", ".env,secrets/**"), tree)

	rr := apiRequest(srv, "/api/tree?branch=main")
	if rr.Code != http.StatusOK {
		t.Fatalf("GET /api/tree status = %d, want %d: %s", rr.Code, http.StatusOK, rr.Body.String())
	}

	got := decodeTree(t, rr)
	assertPathPresent(t, got, "src/main.go")
	assertPathAbsent(t, got, ".env")
	assertPathAbsent(t, got, "secrets")
	assertPathAbsent(t, got, "secrets/prod.yml")
}

func TestTreeKeepsParentDirectoriesForAllowedDescendants(t *testing.T) {
	tree := &github.Tree{Tree: []github.TreeEntry{
		{Path: "docs", Type: "tree"},
		{Path: "docs/guide", Type: "tree"},
		{Path: "docs/guide/intro.md", Type: "blob", SHA: shaDocs},
		{Path: "docs/private.md", Type: "blob", SHA: shaSecret},
		{Path: "src", Type: "tree"},
		{Path: "src/main.go", Type: "blob", SHA: shaMain},
	}}
	srv, _ := newTestServer(t, mustFilter(t, "docs/guide/intro.md", ""), tree)

	rr := apiRequest(srv, "/api/tree?branch=main")
	if rr.Code != http.StatusOK {
		t.Fatalf("GET /api/tree status = %d, want %d: %s", rr.Code, http.StatusOK, rr.Body.String())
	}

	got := decodeTree(t, rr)
	assertPathPresent(t, got, "docs")
	assertPathPresent(t, got, "docs/guide")
	assertPathPresent(t, got, "docs/guide/intro.md")
	assertPathAbsent(t, got, "docs/private.md")
	assertPathAbsent(t, got, "src")
	assertPathAbsent(t, got, "src/main.go")
}

func TestBlobRejectsDeniedPathWhenFiltered(t *testing.T) {
	srv, fake := newTestServer(t, mustFilter(t, "", ".env"), testTree())

	rr := apiRequest(srv, fmt.Sprintf("/api/blob?branch=main&sha=%s&path=.env", shaEnv))
	if rr.Code != http.StatusNotFound {
		t.Fatalf("GET /api/blob status = %d, want %d", rr.Code, http.StatusNotFound)
	}
	if fake.getBlobCalls != 0 {
		t.Fatalf("GetBlob calls = %d, want 0", fake.getBlobCalls)
	}
}

func TestBlobRejectsSHAPathMismatchWhenFiltered(t *testing.T) {
	srv, fake := newTestServer(t, mustFilter(t, "src/**", ""), testTree())

	rr := apiRequest(srv, fmt.Sprintf("/api/blob?branch=main&sha=%s&path=src/main.go", shaDocs))
	if rr.Code != http.StatusNotFound {
		t.Fatalf("GET /api/blob status = %d, want %d", rr.Code, http.StatusNotFound)
	}
	if fake.getBlobCalls != 0 {
		t.Fatalf("GetBlob calls = %d, want 0", fake.getBlobCalls)
	}
}

func TestBlobRequiresPathAndBranchWhenFiltered(t *testing.T) {
	tests := []struct {
		name string
		url  string
	}{
		{name: "missing path", url: fmt.Sprintf("/api/blob?branch=main&sha=%s", shaMain)},
		{name: "missing branch", url: fmt.Sprintf("/api/blob?sha=%s&path=src/main.go", shaMain)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv, fake := newTestServer(t, mustFilter(t, "src/**", ""), testTree())

			rr := apiRequest(srv, tt.url)
			if rr.Code != http.StatusBadRequest {
				t.Fatalf("GET /api/blob status = %d, want %d", rr.Code, http.StatusBadRequest)
			}
			if fake.getBlobCalls != 0 {
				t.Fatalf("GetBlob calls = %d, want 0", fake.getBlobCalls)
			}
		})
	}
}

func newTestServer(t *testing.T, policy *filter.Policy, tree *github.Tree) (*Server, *fakeGitHub) {
	t.Helper()
	done := make(chan struct{})
	t.Cleanup(func() { close(done) })

	fake := &fakeGitHub{
		trees: map[string]*github.Tree{"main": tree},
		blobs: map[string][]byte{
			shaMain:   []byte("package main\n"),
			shaEnv:    []byte("SECRET=placeholder\n"),
			shaDocs:   []byte("# Intro\n"),
			shaSecret: []byte("secret\n"),
		},
	}
	srv := New(Config{
		Owner:      "owner",
		Repo:       "repo",
		Branch:     "main",
		RepoInfo:   &github.RepoInfo{FullName: "owner/repo", DefaultBranch: "main", Private: true},
		Branches:   []string{"main"},
		GitHub:     fake,
		Sessions:   session.NewManager(time.Hour, done),
		Tree:       tree,
		PathFilter: policy,
	})
	return srv, fake
}

func apiRequest(srv *Server, target string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodGet, target, nil)
	req.Header.Set("X-Relay-Token", srv.token)
	rr := httptest.NewRecorder()
	srv.srv.Handler.ServeHTTP(rr, req)
	return rr
}

func decodeTree(t *testing.T, rr *httptest.ResponseRecorder) *github.Tree {
	t.Helper()
	var tree github.Tree
	if err := json.NewDecoder(rr.Body).Decode(&tree); err != nil {
		t.Fatalf("decoding tree response: %v", err)
	}
	return &tree
}

func assertPathPresent(t *testing.T, tree *github.Tree, repoPath string) {
	t.Helper()
	if !treeContainsPath(tree, repoPath) {
		t.Fatalf("expected path %q in tree; got %#v", repoPath, tree.Tree)
	}
}

func assertPathAbsent(t *testing.T, tree *github.Tree, repoPath string) {
	t.Helper()
	if treeContainsPath(tree, repoPath) {
		t.Fatalf("expected path %q to be absent; got %#v", repoPath, tree.Tree)
	}
}

func treeContainsPath(tree *github.Tree, repoPath string) bool {
	for _, entry := range tree.Tree {
		if entry.Path == repoPath {
			return true
		}
	}
	return false
}

func mustFilter(t *testing.T, allow, deny string) *filter.Policy {
	t.Helper()
	policy, err := filter.NewPolicy(allow, deny)
	if err != nil {
		t.Fatalf("NewPolicy() error = %v", err)
	}
	return policy
}

func testTree() *github.Tree {
	return &github.Tree{Tree: []github.TreeEntry{
		{Path: "src", Type: "tree"},
		{Path: "src/main.go", Type: "blob", SHA: shaMain},
		{Path: "docs", Type: "tree"},
		{Path: "docs/intro.md", Type: "blob", SHA: shaDocs},
		{Path: ".env", Type: "blob", SHA: shaEnv},
		{Path: "secrets", Type: "tree"},
		{Path: "secrets/prod.yml", Type: "blob", SHA: shaSecret},
	}}
}

type fakeGitHub struct {
	trees        map[string]*github.Tree
	blobs        map[string][]byte
	getBlobCalls int
}

func (f *fakeGitHub) GetTree(_ context.Context, _, _, ref string) (*github.Tree, error) {
	tree, ok := f.trees[ref]
	if !ok {
		return nil, fmt.Errorf("missing tree for ref %q", ref)
	}
	return tree, nil
}

func (f *fakeGitHub) GetBlob(_ context.Context, _, _, sha string) ([]byte, error) {
	f.getBlobCalls++
	data, ok := f.blobs[sha]
	if !ok {
		return nil, fmt.Errorf("missing blob %q", sha)
	}
	return data, nil
}

func (f *fakeGitHub) GetCommits(_ context.Context, _, _, _ string) ([]github.CommitInfo, error) {
	return nil, nil
}
