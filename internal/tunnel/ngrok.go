package tunnel

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os/exec"
	"regexp"
	"strings"
	"time"
)

type ngrokTunnel struct {
	cmd *exec.Cmd
	url string
}
type noopTunnel struct{ url string }

func openNgrok(ctx context.Context, port int) (*ngrokTunnel, error) {
	bin, err := exec.LookPath("ngrok")
	if err != nil {
		return nil, fmt.Errorf(
			"ngrok not found in PATH: install from https://ngrok.com/download",
		)
	}

	cmd := exec.CommandContext(ctx, bin, "http", fmt.Sprintf("%d", port))

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, fmt.Errorf("creating ngrok stderr pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("starting ngrok: %w", err)
	}

	t := &ngrokTunnel{cmd: cmd}

	url, err := waitForNgrokURL(stderr, 30*time.Second)
	if err != nil {
		_ = cmd.Process.Kill()
		return nil, err
	}
	t.url = url

	go io.Copy(io.Discard, stderr) //nolint:errcheck
	return t, nil
}

var ngrokURLPattern = regexp.MustCompile(`https://[a-z0-9]+\.ngrok[-a-z]*.io`)

func waitForNgrokURL(r io.Reader, timeout time.Duration) (string, error) {
	urlCh := make(chan string, 1)
	errCh := make(chan error, 1)

	go func() {
		scanner := bufio.NewScanner(r)
		for scanner.Scan() {
			line := scanner.Text()
			if strings.Contains(line, "ngrok") {
				if match := ngrokURLPattern.FindString(line); match != "" {
					urlCh <- match
					return
				}
			}
		}
		errCh <- fmt.Errorf("ngrok exited before producing a URL")
	}()

	select {
	case url := <-urlCh:
		return url, nil
	case err := <-errCh:
		return "", err
	case <-time.After(timeout):
		return "", fmt.Errorf("timed out waiting for ngrok URL after %s", timeout)
	}
}

func (t *ngrokTunnel) URL() string  { return t.url }
func (t *ngrokTunnel) Close() error { return t.cmd.Process.Kill() }
func (t *noopTunnel) URL() string   { return t.url }
func (t *noopTunnel) Close() error  { return nil }
