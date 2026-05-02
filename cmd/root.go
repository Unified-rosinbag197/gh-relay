// Package cmd implements the gh-relay command-line interface.
package cmd

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.ibm.com/soub4i/gh-relay/internal/version"
)

// rootCmd is the top-level command set.
var rootCmd = &command{
	name:  "gh-relay",
	short: "Ephemeral, read-only code sharing for private GitHub repositories.",
	long: `gh-relay lets a repository maintainer share read-only access to a
private GitHub repository with a guest without adding them as a collaborator.

It starts a local proxy server and exposes it through a secure tunnel.
The guest receives a temporary URL and can browse the code in their browser.

Usage:
  gh-relay <command> [flags]

Commands:
  share     Start a sharing session for a repository
  version   Print version information

Run "gh-relay <command> --help" for more information about a command.`,
}

// Execute is the entry point called from main.
func Execute() error {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, rootCmd.long)
		return nil
	}

	switch os.Args[1] {
	case "share":
		return runShare(os.Args[2:])
	case "version":
		return runVersion()
	case "--help", "-h", "help":
		fmt.Fprintln(os.Stderr, rootCmd.long)
		return nil
	default:
		return fmt.Errorf("unknown command %q - run \"gh-relay --help\"", os.Args[1])
	}
}

// command holds metadata about a subcommand.
type command struct {
	name  string
	short string
	long  string
}

func runShare(args []string) error {
	fs := flag.NewFlagSet("share", flag.ContinueOnError)

	var f shareFlags
	fs.StringVar(&f.token, "token", "", "GitHub Personal Access Token (required)")
	fs.StringVar(&f.repo, "repo", "", "Target repository, e.g. owner/repo (required)")
	fs.StringVar(&f.branch, "branch", "main", "Branch to share (default: main)")
	fs.IntVar(&f.port, "port", 8080, "Local port for the proxy server")
	fs.DurationVar(&f.expire, "expire", 0, "Session duration, e.g. 30m or 1h (default: unlimited)")
	fs.StringVar(&f.tunnel, "tunnel", "cloudflare", "Tunnel provider: cloudflare, ngrok, or none")
	fs.BoolVar(&f.scanSecrets, "scan-secrets", true, "Scan repository paths for sensitive files before sharing")
	fs.Var(negatedBoolFlag{target: &f.scanSecrets}, "no-scan-secrets", "Disable pre-share sensitive file scanning")
	fs.BoolVar(&f.scanContent, "scan-content", false, "Also scan small text blobs for common secret patterns")
	fs.BoolVar(&f.failOnSecrets, "fail-on-secrets", false, "Exit non-zero if the pre-share scan finds potential secrets")

	fs.Usage = func() {
		fmt.Fprintln(os.Stderr, `Usage: gh-relay share [flags]

Start an ephemeral sharing session. The command validates your GitHub token,
fetches repository metadata, opens a tunnel, and prints a URL for your guest.
Press Ctrl+C or let --expire elapse to end the session.

Flags:`)
		fs.PrintDefaults()
		fmt.Fprintln(os.Stderr, `
Examples:
  gh-relay share --token ghp_abc123 --repo my-org/private-app --expire 1h
  gh-relay share --token ghp_abc123 --repo my-org/private-app --tunnel ngrok --port 9000
  gh-relay share --token ghp_abc123 --repo my-org/private-app --scan-content
  gh-relay share --token ghp_abc123 --repo my-org/private-app --fail-on-secrets
  gh-relay share --token ghp_abc123 --repo my-org/private-app --no-scan-secrets --tunnel none`)
	}

	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return err
	}

	if err := ValidateShareFlags(f); err != nil {
		return err
	}
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	// If --expire is set, also cancel after that duration.
	if f.expire > 0 {
		var expireCancel context.CancelFunc
		ctx, expireCancel = context.WithTimeout(ctx, f.expire)
		defer expireCancel()
	}

	return RunShareSession(ctx, f)
}

type negatedBoolFlag struct {
	target *bool
}

func (f negatedBoolFlag) Set(value string) error {
	disabled, err := strconv.ParseBool(value)
	if err != nil {
		return err
	}
	if f.target != nil {
		*f.target = !disabled
	}
	return nil
}

func (f negatedBoolFlag) String() string {
	if f.target == nil {
		return "false"
	}
	return strconv.FormatBool(!*f.target)
}

func (f negatedBoolFlag) IsBoolFlag() bool {
	return true
}

func runVersion() error {
	fmt.Printf("gh-relay version %s\n", version.Version)
	return nil
}

func ValidateShareFlags(f shareFlags) error {
	if f.token == "" {
		return fmt.Errorf("--token is required")
	}
	if f.repo == "" {
		return fmt.Errorf("--repo is required")
	}
	if f.port < 1 || f.port > 65535 {
		return fmt.Errorf("--port must be between 1 and 65535")
	}
	return nil
}
