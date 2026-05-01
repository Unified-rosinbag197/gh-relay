package server

import (
	"encoding/json"
	"log"
	"net/http"
	"path/filepath"
	"strings"
)

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	if err := enc.Encode(v); err != nil {
		log.Printf("[error] encoding JSON response: %v", err)
	}
}

// isSafeSHA validates that sha is a 40-char hex string.
func isSafeSHA(sha string) bool {
	if len(sha) != 40 {
		return false
	}
	for _, c := range sha {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
			return false
		}
	}
	return true
}

// isSafeBranchName rejects branch names with shell-unsafe characters.
func isSafeBranchName(name string) bool {
	if name == "" || len(name) > 255 {
		return false
	}
	for _, c := range name {
		if c == '\x00' || c == '\\' || c == '^' || c == ':' || c == '?' || c == '[' || c == ' ' {
			return false
		}
	}
	return true
}

func blobContentType(path string) string {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".png":
		return "image/png"
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".gif":
		return "image/gif"
	case ".webp":
		return "image/webp"
	case ".svg":
		return "image/svg+xml"
	case ".ico":
		return "image/x-icon"
	case ".pdf":
		return "application/pdf"
	case ".zip", ".tar", ".gz", ".bz2", ".xz", ".7z",
		".exe", ".bin", ".dll", ".so", ".dylib",
		".wasm", ".pyc", ".pyo", ".class":
		return "application/octet-stream"
	default:
		return "text/plain; charset=utf-8"
	}
}

func getClientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		if i := strings.Index(xff, ","); i != -1 {
			return strings.TrimSpace(xff[:i])
		}
		return xff
	}
	if ip := r.Header.Get("CF-Connecting-IP"); ip != "" {
		return ip
	}
	return r.RemoteAddr
}
