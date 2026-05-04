package secretscan

import (
	"bytes"
	"path"
	"regexp"
	"strings"
	"unicode/utf8"
)

func defaultPathRules() []pathRule {
	return []pathRule{
		{
			rule: Rule{Name: "dotenv file", Severity: SeverityHigh},
			match: func(entry Entry) bool {
				base := baseName(entry.Path)
				return base == ".env" || strings.HasPrefix(base, ".env.")
			},
		},
		{
			rule: Rule{Name: "private key/certificate file", Severity: SeverityHigh},
			match: func(entry Entry) bool {
				ext := pathExt(entry.Path)
				return ext == ".pem" || ext == ".key"
			},
		},
		{
			rule: Rule{Name: "SSH private key filename", Severity: SeverityHigh},
			match: func(entry Entry) bool {
				base := baseName(entry.Path)
				return base == "id_rsa" || base == "id_ed25519"
			},
		},
		{
			rule: Rule{Name: "secrets directory", Severity: SeverityHigh},
			match: func(entry Entry) bool {
				return hasPathSegment(entry.Path, "secrets")
			},
		},
		{
			rule: Rule{Name: "credentials file", Severity: SeverityHigh},
			match: func(entry Entry) bool {
				base := baseName(entry.Path)
				return base == "credentials" || base == "credentials.yml"
			},
		},
		{
			rule: Rule{Name: "PKCS12 certificate bundle", Severity: SeverityHigh},
			match: func(entry Entry) bool {
				ext := pathExt(entry.Path)
				return ext == ".p12" || ext == ".pfx"
			},
		},
		{
			rule: Rule{Name: "kubeconfig file", Severity: SeverityHigh},
			match: func(entry Entry) bool {
				return baseName(entry.Path) == "kubeconfig"
			},
		},
		{
			rule: Rule{Name: "npm credentials file", Severity: SeverityHigh},
			match: func(entry Entry) bool {
				return baseName(entry.Path) == ".npmrc"
			},
		},
		{
			rule: Rule{Name: "PyPI credentials file", Severity: SeverityHigh},
			match: func(entry Entry) bool {
				return baseName(entry.Path) == ".pypirc"
			},
		},
		{
			rule: Rule{Name: "Terraform variable secrets file", Severity: SeverityHigh},
			match: func(entry Entry) bool {
				return baseName(entry.Path) == "terraform.tfvars"
			},
		},
	}
}

func defaultContentRules() []contentRule {
	return []contentRule{
		{
			rule: Rule{Name: "AWS access key ID", Severity: SeverityHigh},
			re:   regexp.MustCompile(`AKIA[0-9A-Z]{16}`),
		},
		{
			rule: Rule{Name: "GitHub token prefix", Severity: SeverityHigh},
			re:   regexp.MustCompile(`(?:ghp_|github_pat_|gho_|ghu_|ghs_|ghr_)`),
		},
		{
			rule: Rule{Name: "private key block header", Severity: SeverityHigh},
			re:   regexp.MustCompile(`-----BEGIN [A-Z0-9 ]*PRIVATE KEY-----`),
		},
		{
			rule: Rule{Name: "generic token assignment", Severity: SeverityMedium},
			re:   regexp.MustCompile(`(?i)\b(password|secret|api_key|token)\s*=`),
		},
	}
}

func severityRank(severity Severity) int {
	switch severity {
	case SeverityHigh:
		return 0
	case SeverityMedium:
		return 1
	default:
		return 2
	}
}

func cleanPath(filePath string) string {
	return strings.TrimPrefix(path.Clean(strings.ReplaceAll(filePath, `\`, `/`)), "./")
}

func baseName(filePath string) string {
	return strings.ToLower(path.Base(cleanPath(filePath)))
}

func pathExt(filePath string) string {
	return strings.ToLower(path.Ext(cleanPath(filePath)))
}

func hasPathSegment(filePath, segment string) bool {
	segment = strings.ToLower(segment)
	for _, part := range strings.Split(cleanPath(filePath), "/") {
		if strings.ToLower(part) == segment {
			return true
		}
	}
	return false
}

func hasBinaryExtension(filePath string) bool {
	switch pathExt(filePath) {
	case ".7z", ".a", ".avi", ".bin", ".bmp", ".bz2", ".class", ".dll", ".dylib",
		".eot", ".exe", ".gif", ".gz", ".ico", ".jar", ".jpeg", ".jpg", ".mov",
		".mp3", ".mp4", ".o", ".otf", ".pdf", ".png", ".pyc", ".pyo", ".so",
		".tar", ".ttf", ".wasm", ".webp", ".woff", ".woff2", ".xz", ".zip":
		return true
	default:
		return false
	}
}

func isLikelyText(data []byte) bool {
	if len(data) == 0 {
		return true
	}
	if bytes.IndexByte(data, 0) >= 0 || !utf8.Valid(data) {
		return false
	}

	control := 0
	for _, b := range data {
		if b < 0x20 && b != '\n' && b != '\r' && b != '\t' {
			control++
		}
	}
	return control*100/len(data) < 5
}
