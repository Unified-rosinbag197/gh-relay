package tunnel

import (
	"context"
	"fmt"
)

// Provider identifies the tunneling backend.
type Provider string

const (
	ProviderCloudflare Provider = "cloudflare"
	ProviderNgrok      Provider = "ngrok"
	ProviderNone       Provider = "none"
)

type Tunnel interface {
	URL() string
	Close() error
}

func Open(ctx context.Context, provider Provider, port int) (Tunnel, error) {
	switch provider {
	case ProviderCloudflare, "":
		return openCloudflare(ctx, port)
	case ProviderNgrok:
		return openNgrok(ctx, port)
	case ProviderNone:
		return &noopTunnel{url: fmt.Sprintf("http://localhost:%d", port)}, nil
	default:
		return nil, fmt.Errorf("unknown tunnel provider: %q (valid: cloudflare, ngrok, none)", provider)
	}
}
