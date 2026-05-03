package cmd

import (
	"context"
	"fmt"
	"io"
	"log"
	"testing"

	"github.ibm.com/soub4i/gh-relay/internal/filter"
	"github.ibm.com/soub4i/gh-relay/internal/github"
)

const (
	testSHAEnv    = "1111111111111111111111111111111111111111"
	testSHAPem    = "2222222222222222222222222222222222222222"
	testSHASecret = "3333333333333333333333333333333333333333"
)

func TestSecretPreflightIgnoresDeniedPaths(t *testing.T) {
	tree := &github.Tree{Tree: []github.TreeEntry{
		{Path: ".env", Type: "blob", SHA: testSHAEnv, Size: 16},
		{Path: "deploy/prod.pem", Type: "blob", SHA: testSHAPem, Size: 16},
		{Path: "secrets", Type: "tree"},
		{Path: "secrets/prod.yml", Type: "blob", SHA: testSHASecret, Size: 16},
	}}
	fake := &secretPreflightFakeGitHub{tree: tree}
	policy := mustSharePolicy(t, "", ".env,*.pem,secrets/**")

	_, err := runSecretPreflight(
		context.Background(),
		log.New(io.Discard, "", 0),
		fake,
		"owner",
		"repo",
		"main",
		shareFlags{scanSecrets: true, scanContent: true, failOnSecrets: true},
		policy,
	)
	if err != nil {
		t.Fatalf("runSecretPreflight() error = %v", err)
	}
	if fake.getBlobCalls != 0 {
		t.Fatalf("GetBlob calls = %d, want 0 for denied content paths", fake.getBlobCalls)
	}
}

func mustSharePolicy(t *testing.T, allow, deny string) *filter.Policy {
	t.Helper()
	policy, err := filter.NewPolicy(allow, deny)
	if err != nil {
		t.Fatalf("NewPolicy() error = %v", err)
	}
	return policy
}

type secretPreflightFakeGitHub struct {
	tree         *github.Tree
	getBlobCalls int
}

func (f *secretPreflightFakeGitHub) GetTree(_ context.Context, _, _, _ string) (*github.Tree, error) {
	return f.tree, nil
}

func (f *secretPreflightFakeGitHub) GetBlob(_ context.Context, _, _, sha string) ([]byte, error) {
	f.getBlobCalls++
	return nil, fmt.Errorf("denied blob %s should not be fetched", sha)
}
