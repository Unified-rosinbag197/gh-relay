package filter

import "testing"

func TestPolicyEmptyAllowsEverything(t *testing.T) {
	p := mustPolicy(t, "", "")

	if !p.AllowPath("src/main.go") {
		t.Fatal("expected empty policy to allow normal path")
	}
}

func TestPolicyAllowOnly(t *testing.T) {
	p := mustPolicy(t, "src/**, README.md", "")

	if !p.AllowPath("src/main.go") {
		t.Fatal("expected allowed src file")
	}
	if !p.AllowPath("README.md") {
		t.Fatal("expected exact README.md to be allowed")
	}
	if p.AllowPath("docs/intro.md") {
		t.Fatal("expected non-matching path to be hidden")
	}
}

func TestPolicyDenyOnly(t *testing.T) {
	p := mustPolicy(t, "", ".env,secrets/**")

	if p.AllowPath(".env") {
		t.Fatal("expected .env to be denied")
	}
	if p.AllowPath("secrets/prod.yml") {
		t.Fatal("expected secrets descendant to be denied")
	}
	if !p.AllowPath("src/main.go") {
		t.Fatal("expected unrelated path to be allowed")
	}
}

func TestPolicyDenyWinsOverAllow(t *testing.T) {
	p := mustPolicy(t, "src/**", "src/private/**")

	if !p.AllowPath("src/main.go") {
		t.Fatal("expected public src file to be allowed")
	}
	if p.AllowPath("src/private/key.txt") {
		t.Fatal("expected deny rule to win over allow rule")
	}
}

func TestPolicyDoubleStarDirectoryPattern(t *testing.T) {
	p := mustPolicy(t, "src/**", "")

	if !p.AllowPath("src/pkg/main.go") {
		t.Fatal("expected src/** to match nested files")
	}
	if p.AllowPath("docs/src/main.go") {
		t.Fatal("expected src/** not to match other directories")
	}
}

func TestPolicyBasenameGlobMatchesAtAnyDepth(t *testing.T) {
	p := mustPolicy(t, "", "*.pem")

	if p.AllowPath("deploy/prod.pem") {
		t.Fatal("expected basename glob to deny nested pem file")
	}
	if p.AllowPath("prod.pem") {
		t.Fatal("expected basename glob to deny root pem file")
	}
	if !p.AllowPath("deploy/prod.key") {
		t.Fatal("expected non-matching basename to be allowed")
	}
}

func TestPolicyDotEnvBasenameMatchesAtAnyDepth(t *testing.T) {
	p := mustPolicy(t, "", ".env")

	if p.AllowPath(".env") {
		t.Fatal("expected root .env to be denied")
	}
	if p.AllowPath("services/api/.env") {
		t.Fatal("expected nested .env to be denied")
	}
	if !p.AllowPath("services/api/.env.example") {
		t.Fatal("expected different basename to be allowed")
	}
}

func TestPolicyRejectsTraversalPatterns(t *testing.T) {
	if _, err := NewPolicy("../secret", ""); err == nil {
		t.Fatal("expected traversal-like allow pattern to be rejected")
	}
	if _, err := NewPolicy("", "safe/../secret"); err == nil {
		t.Fatal("expected traversal-like deny pattern to be rejected")
	}
}

func TestPolicyRejectsTraversalPaths(t *testing.T) {
	p := mustPolicy(t, "src/**", "")

	if p.AllowPath("../secret") {
		t.Fatal("expected traversal-like path to be rejected")
	}
	if p.AllowPath("src/../secret") {
		t.Fatal("expected normalized traversal-like path to be rejected")
	}
}

func mustPolicy(t *testing.T, allow, deny string) *Policy {
	t.Helper()
	p, err := NewPolicy(allow, deny)
	if err != nil {
		t.Fatalf("NewPolicy() error = %v", err)
	}
	return p
}
