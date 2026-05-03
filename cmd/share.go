package cmd

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.ibm.com/soub4i/gh-relay/internal/github"
	"github.ibm.com/soub4i/gh-relay/internal/logo"
	"github.ibm.com/soub4i/gh-relay/internal/secretscan"
	"github.ibm.com/soub4i/gh-relay/internal/server"
	"github.ibm.com/soub4i/gh-relay/internal/session"
	"github.ibm.com/soub4i/gh-relay/internal/tunnel"
)

type shareFlags struct {
	token  string
	repo   string
	branch string
	port   int
	expire time.Duration
	tunnel string

	scanSecrets   bool
	scanContent   bool
	failOnSecrets bool
	audit         bool
}

func RunShareSession(ctx context.Context, f shareFlags) error {
	logger := log.New(os.Stderr, "", 0)

	printBanner(logger)

	logger.Print("Validating GitHub token…")
	gh := github.NewClient(f.token)
	if err := gh.ValidateToken(ctx); err != nil {
		return fmt.Errorf("Error:  %w", err)
	}
	logger.Print("  Token valid")

	owner, repo, err := github.ParseOwnerRepo(f.repo)
	if err != nil {
		return err
	}

	logger.Printf("  Fetching repository info for %s/%s…", owner, repo)
	repoInfo, err := gh.GetRepo(ctx, owner, repo)
	if err != nil {
		return fmt.Errorf("Error:  %w", err)
	}
	logger.Printf("  Repository: %s (%s)", repoInfo.FullName, visibilityLabel(repoInfo.Private))

	branch := f.branch
	if branch == "main" && repoInfo.DefaultBranch != "" && repoInfo.DefaultBranch != "main" {
		branch = repoInfo.DefaultBranch
		logger.Printf("Using default branch: %s", branch)
	}

	logger.Print("  Fetching branch list…")
	branches, err := gh.GetBranches(ctx, owner, repo)
	if err != nil {
		logger.Printf(" Could not fetch branches: %v - proceeding with %q only", err, branch)
		branches = []string{branch}
	} else {
		logger.Printf("Found %d branch(es)", len(branches))
	}

	if !containsString(branches, branch) {
		return fmt.Errorf("branch %q not found in %s/%s", branch, owner, repo)
	}

	initialTree, err := runSecretPreflight(ctx, logger, gh, owner, repo, branch, f)
	if err != nil {
		return err
	}

	done := ctx.Done()
	sessionTTL := 24 * time.Hour
	if f.expire > 0 {
		sessionTTL = f.expire
	}
	sessions := session.NewManager(sessionTTL, done)
	var auditLog *server.AuditLog
	if f.audit {
		auditLog = &server.AuditLog{}
	}

	cfg := server.Config{
		Owner:    owner,
		Repo:     repo,
		Branch:   branch,
		RepoInfo: repoInfo,
		Branches: branches,
		GitHub:   gh,
		Sessions: sessions,
		Port:     f.port,
		Tree:     initialTree,
		AuditLog: auditLog,
	}
	serverErr := startServer(ctx, cfg)

	logger.Printf("Opening %s tunnel…", f.tunnel)
	tun, err := tunnel.Open(ctx, tunnel.Provider(f.tunnel), f.port)
	if err != nil {
		return fmt.Errorf("Error:  %w", err)
	}
	defer tun.Close()
	logger.Printf("Tunnel active")

	printShareInfo(logger, f, owner, repo, branch, repoInfo, tun.URL())

	select {
	case <-ctx.Done():
		logger.Println("\nSession ended, tunnel closed, server shut down.")
		logger.Println("The shared URL is now inactive. All session cookies are invalid.")
		if auditLog != nil {
			auditLog.Summary(logger)
		}
	case err := <-serverErr:
		return fmt.Errorf("server error: %w", err)
	}

	return nil
}

func runSecretPreflight(ctx context.Context, logger *log.Logger, gh *github.Client, owner, repo, branch string, f shareFlags) (*github.Tree, error) {
	if !f.scanSecrets {
		return nil, nil
	}

	logger.Print("  Scanning repository tree for sensitive paths…")
	tree, err := gh.GetTree(ctx, owner, repo, branch)
	if err != nil {
		return nil, fmt.Errorf("secret scan could not fetch repository tree: %w", err)
	}
	if tree.Truncated {
		logger.Print("  GitHub returned a truncated tree; the secret-risk scan may be incomplete.")
	}

	scanner := secretscan.New(secretscan.Options{ScanContent: f.scanContent})
	findings := scanner.ScanEntries(secretScanEntries(tree.Tree))

	if f.scanContent {
		logger.Print("  Scanning small text blobs for secret patterns…")
		contentFindings, skipped, err := scanSecretContent(ctx, scanner, gh, owner, repo, tree)
		if err != nil {
			return tree, err
		}
		findings = append(findings, contentFindings...)
		if skipped > 0 {
			logger.Printf("  Secret content scan skipped %d file(s) that could not be fetched.", skipped)
		}
	}

	secretscan.SortFindings(findings)
	if len(findings) == 0 {
		logger.Print("  No potential sensitive files or secrets detected.")
		return tree, nil
	}

	if err := handleSecretFindings(logger, findings, f); err != nil {
		return tree, err
	}
	return tree, nil
}

func secretScanEntries(entries []github.TreeEntry) []secretscan.Entry {
	out := make([]secretscan.Entry, 0, len(entries))
	for _, entry := range entries {
		out = append(out, secretscan.Entry{
			Path: entry.Path,
			Type: entry.Type,
			Size: entry.Size,
		})
	}
	return out
}

func scanSecretContent(ctx context.Context, scanner *secretscan.Scanner, gh *github.Client, owner, repo string, tree *github.Tree) ([]secretscan.Finding, int, error) {
	var findings []secretscan.Finding
	skipped := 0

	for _, entry := range tree.Tree {
		if entry.Type != "blob" || entry.SHA == "" || !scanner.ShouldScanContent(entry.Path, entry.Size) {
			continue
		}

		data, err := gh.GetBlob(ctx, owner, repo, entry.SHA)
		if err != nil {
			if ctx.Err() != nil {
				return nil, skipped, ctx.Err()
			}
			skipped++
			continue
		}

		findings = append(findings, scanner.ScanContent(entry.Path, data)...)
	}

	return findings, skipped, nil
}

func handleSecretFindings(logger *log.Logger, findings []secretscan.Finding, f shareFlags) error {
	printSecretFindings(logger, findings)

	if f.failOnSecrets {
		return fmt.Errorf("secret-risk scan found %d finding(s); sharing stopped by --fail-on-secrets", len(findings))
	}

	if !stdinIsTerminal() {
		logger.Print("Non-interactive input detected; continuing without confirmation. Use --fail-on-secrets to stop on findings.")
		return nil
	}

	ok, err := promptContinue()
	if err != nil {
		return fmt.Errorf("reading confirmation: %w", err)
	}
	if !ok {
		return fmt.Errorf("sharing canceled after secret-risk warning")
	}
	return nil
}

func printSecretFindings(logger *log.Logger, findings []secretscan.Finding) {
	logger.Println()
	logger.Println("⚠️  Potential sensitive files or secrets detected before sharing:")
	logger.Println()
	for _, finding := range findings {
		logger.Printf("%-7s %-8s %-28s matched rule: %s", finding.Severity, finding.Type, finding.Path, finding.Rule)
	}
	logger.Println()
	logger.Println("gh-relay does not print secret values. Review these files before exposing this repo.")
	logger.Println()
}

func stdinIsTerminal() bool {
	info, err := os.Stdin.Stat()
	if err != nil {
		return false
	}
	return info.Mode()&os.ModeCharDevice != 0
}

func promptContinue() (bool, error) {
	fmt.Fprint(os.Stderr, "Continue sharing? [y/N]: ")
	reader := bufio.NewReader(os.Stdin)
	answer, err := reader.ReadString('\n')
	if err != nil && len(answer) == 0 {
		return false, err
	}
	answer = strings.TrimSpace(strings.ToLower(answer))
	return answer == "y" || answer == "yes", nil
}

func printBanner(l *log.Logger) {
	l.Println(strings.Repeat("-", 54))
	for _, line := range strings.Split(logo.ASCIILOGO, "\n") {
		l.Printf("  %s\n", line)
	}
	l.Println(strings.Repeat("-", 54))
}

func printShareInfo(l *log.Logger, f shareFlags, owner, repo, branch string, info *github.RepoInfo, url string) {
	l.Println()
	l.Println(strings.Repeat("-", 54))
	l.Println("  SESSION ACTIVE")
	l.Println(strings.Repeat("-", 54))
	l.Printf("  Repository : %s/%s", owner, repo)
	l.Printf("  Branch     : %s", branch)
	l.Printf("  Visibility : %s", visibilityLabel(info.Private))
	if info.Description != "" {
		l.Printf("  Description: %s", info.Description)
	}
	l.Println()
	l.Printf("  Share this URL with your guest:")
	l.Printf("    %s", url)
	l.Println()
	if f.expire > 0 {
		l.Printf("  Session expires in: %s (%s)", f.expire, time.Now().Add(f.expire).Format(time.RFC1123))
	} else {
		l.Print("  Session expires: when you press Ctrl+C")
	}
	l.Println()
	l.Println("  Press Ctrl+C to end the session immediately.")
	l.Println(strings.Repeat("-", 54))
}

func visibilityLabel(private bool) string {
	if private {
		return "private"
	}
	return "public"
}

func containsString(ss []string, s string) bool {
	for _, v := range ss {
		if v == s {
			return true
		}
	}
	return false
}

func startServer(ctx context.Context, cfg server.Config) chan error {
	srv := server.New(cfg)

	serverErr := make(chan error, 1)
	go func() {
		if err := srv.Start(ctx); err != nil {
			serverErr <- err
		}
	}()

	time.Sleep(150 * time.Millisecond)
	return serverErr
}
