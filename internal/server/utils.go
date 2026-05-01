package server

import (
	"encoding/json"
	"log"
	"mime"
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

// mimeForPath returns an appropriate Content-Type for the given file path.
func mimeForPath(path string) string {
	if path == "" {
		return "text/plain; charset=utf-8"
	}
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".go":
		return "text/plain; charset=utf-8"
	case ".js", ".mjs", ".cjs":
		return "application/javascript"
	case ".ts", ".tsx":
		return "application/typescript"
	case ".json":
		return "application/json"
	case ".html", ".htm":
		return "text/html; charset=utf-8"
	case ".css":
		return "text/css"
	case ".md", ".markdown":
		return "text/markdown; charset=utf-8"
	case ".sh", ".bash":
		return "text/x-shellscript"
	case ".py":
		return "text/x-python"
	case ".rs":
		return "text/x-rustsrc"
	case ".yaml", ".yml":
		return "text/yaml"
	case ".toml":
		return "text/toml"
	case ".xml":
		return "application/xml"
	case ".svg":
		return "image/svg+xml"
	case ".png":
		return "image/png"
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".gif":
		return "image/gif"
	case ".pdf":
		return "application/pdf"
	default:
		if t := mime.TypeByExtension(ext); t != "" {
			return t
		}
		return "application/octet-stream"
	}
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
		// Every other file — source code, config, markdown, etc. — is sent as
		// plain text. The frontend does all syntax highlighting itself.
		return "text/plain; charset=utf-8"
	}
}
