package cmd

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.ibm.com/soub4i/gh-relay/internal/github"
	"github.ibm.com/soub4i/gh-relay/internal/logo"
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

	done := ctx.Done()
	sessionTTL := 24 * time.Hour
	if f.expire > 0 {
		sessionTTL = f.expire
	}
	sessions := session.NewManager(sessionTTL, done)

	cfg := server.Config{
		Owner:    owner,
		Repo:     repo,
		Branch:   branch,
		RepoInfo: repoInfo,
		Branches: branches,
		GitHub:   gh,
		Sessions: sessions,
		Port:     f.port,
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
	case err := <-serverErr:
		return fmt.Errorf("server error: %w", err)
	}

	return nil
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
