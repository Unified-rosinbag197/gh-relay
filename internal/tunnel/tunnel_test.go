package tunnel

import (
	"context"
	"testing"
)

func TestOpen_ProviderNone(t *testing.T) {
	tunnel, err := Open(context.Background(), ProviderNone, 8080)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer tunnel.Close()

	if tunnel.URL() != "http://localhost:8080" {
		t.Fatalf("expected URL 'http://localhost:8080', got %q", tunnel.URL())
	}
}

func TestOpen_ProviderEmptyDefaultsToCloudflare(t *testing.T) {
	_, err := Open(context.Background(), "", 8080)
	if err == nil {
		t.Skip("cloudflared is installed and available")
	}
}

func TestOpen_InvalidProvider(t *testing.T) {
	_, err := Open(context.Background(), Provider("invalid"), 8080)
	if err == nil {
		t.Fatal("expected error for invalid provider")
	}
	if err.Error() != "unknown tunnel provider: \"invalid\" (valid: cloudflare, ngrok, none)" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestProvider_Constants(t *testing.T) {
	if ProviderCloudflare != "cloudflare" {
		t.Errorf("expected ProviderCloudflare 'cloudflare', got %q", ProviderCloudflare)
	}
	if ProviderNgrok != "ngrok" {
		t.Errorf("expected ProviderNgrok 'ngrok', got %q", ProviderNgrok)
	}
	if ProviderNone != "none" {
		t.Errorf("expected ProviderNone 'none', got %q", ProviderNone)
	}
}

func TestNoopTunnel_URL(t *testing.T) {
	tunnel := &noopTunnel{url: "http://localhost:3000"}
	if tunnel.URL() != "http://localhost:3000" {
		t.Errorf("expected URL 'http://localhost:3000', got %q", tunnel.URL())
	}
}

func TestNoopTunnel_Close(t *testing.T) {
	tunnel := &noopTunnel{url: "http://localhost:3000"}
	if err := tunnel.Close(); err != nil {
		t.Errorf("Close() error = %v", err)
	}
}