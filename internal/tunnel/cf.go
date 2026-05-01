package tunnel

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os/exec"
	"regexp"
	"time"
)

type cloudflareTunnel struct {
	cmd    *exec.Cmd
	url    string
	stderr io.ReadCloser
}

var cfURLPattern = regexp.MustCompile(`https://[a-z0-9\-]+\.trycloudflare\.com`)

func openCloudflare(ctx context.Context, port int) (*cloudflareTunnel, error) {
	bin, err := exec.LookPath("cloudflared")
	if err != nil {
		return nil, fmt.Errorf(
			"cloudflared not found in PATH: install from https://developers.cloudflare.com/cloudflare-one/connections/connect-networks/downloads/",
		)
	}

	cmd := exec.CommandContext(ctx, bin, "tunnel", "--url", fmt.Sprintf("http://localhost:%d", port))

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, fmt.Errorf("creating cloudflared stderr pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("starting cloudflared: %w", err)
	}

	t := &cloudflareTunnel{cmd: cmd, stderr: stderr}

	url, err := t.waitForURL(30 * time.Second)
	if err != nil {
		_ = cmd.Process.Kill()
		return nil, err
	}
	t.url = url

	// Drain stderr in background to prevent pipe blocking.
	go io.Copy(io.Discard, stderr) //nolint:errcheck

	return t, nil
}

func (t *cloudflareTunnel) waitForURL(timeout time.Duration) (string, error) {
	urlCh := make(chan string, 1)
	errCh := make(chan error, 1)

	go func() {
		scanner := bufio.NewScanner(t.stderr)
		for scanner.Scan() {
			line := scanner.Text()
			if match := cfURLPattern.FindString(line); match != "" {
				urlCh <- match
				return
			}
		}
		if err := scanner.Err(); err != nil {
			errCh <- fmt.Errorf("reading cloudflared output: %w", err)
		} else {
			errCh <- fmt.Errorf("cloudflared exited before producing a URL")
		}
	}()

	select {
	case url := <-urlCh:
		return url, nil
	case err := <-errCh:
		return "", err
	case <-time.After(timeout):
		return "", fmt.Errorf("timed out waiting for cloudflared URL after %s", timeout)
	}
}

func (t *cloudflareTunnel) URL() string  { return t.url }
func (t *cloudflareTunnel) Close() error { return t.cmd.Process.Kill() }
