package github

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestValidateToken_Unauthorized(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/user", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	client := &Client{
		token:      "test-token",
		httpClient: server.Client(),
	}

	err := client.ValidateToken(context.Background())
	if err == nil {
		t.Fatal("expected error for unauthorized token")
	}
}

func TestGetRepo_NotFound(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/repos/owner/repo", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	client := &Client{
		token:      "test-token",
		httpClient: server.Client(),
	}

	_, err := client.GetRepo(context.Background(), "owner", "repo")
	if err == nil {
		t.Fatal("expected error for not found repo")
	}
}

func TestGetBlob_NotFound(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/repos/owner/repo/git/blobs/sha123", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	client := &Client{
		token:      "test-token",
		httpClient: server.Client(),
	}

	_, err := client.GetBlob(context.Background(), "owner", "repo", "sha123")
	if err == nil {
		t.Fatal("expected error for not found blob")
	}
}

func TestGetCommits_Non200(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/repos/owner/repo/commits", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	client := &Client{
		token:      "test-token",
		httpClient: server.Client(),
	}

	_, err := client.GetCommits(context.Background(), "owner", "repo", "main")
	if err == nil {
		t.Fatal("expected error for non-200 status")
	}
}

func TestGetBranches_Non200(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/repos/owner/repo/branches", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	client := &Client{
		token:      "test-token",
		httpClient: server.Client(),
	}

	_, err := client.GetBranches(context.Background(), "owner", "repo")
	if err == nil {
		t.Fatal("expected error for non-200 status")
	}
}

func TestResolveRef_NotFound(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/repos/owner/repo/commits/nonexistent", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	client := &Client{
		token:      "test-token",
		httpClient: server.Client(),
	}

	_, err := client.resolveRef(context.Background(), "owner", "repo", "nonexistent")
	if err == nil {
		t.Fatal("expected error for not found ref")
	}
}

func TestGetRepo_JSONDecodeError(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/repos/owner/repo", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte("not valid json"))
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	client := &Client{
		token:      "test-token",
		httpClient: server.Client(),
	}

	_, err := client.GetRepo(context.Background(), "owner", "repo")
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestGetTree_ResolveRefError(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/repos/owner/repo/commits/main", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(struct {
			SHA string `json:"sha"`
		}{SHA: ""})
	})
	mux.HandleFunc("/repos/owner/repo/git/trees/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	client := &Client{
		token:      "test-token",
		httpClient: server.Client(),
	}

	_, err := client.GetTree(context.Background(), "owner", "repo", "main")
	if err == nil {
		t.Fatal("expected error when ref not found")
	}
}

func TestGetTree_TreeFetchError(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/repos/owner/repo/commits/main", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(struct {
			SHA string `json:"sha"`
		}{SHA: "abc123"})
	})
	mux.HandleFunc("/repos/owner/repo/git/trees/abc123", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	client := &Client{
		token:      "test-token",
		httpClient: server.Client(),
	}

	_, err := client.GetTree(context.Background(), "owner", "repo", "main")
	if err == nil {
		t.Fatal("expected error when tree fetch fails")
	}
}

func TestGetBlob_DecodeError(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/repos/owner/repo/git/blobs/sha123", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte("not valid json"))
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	client := &Client{
		token:      "test-token",
		httpClient: server.Client(),
	}

	_, err := client.GetBlob(context.Background(), "owner", "repo", "sha123")
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

