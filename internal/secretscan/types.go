// Package secretscan implements lightweight pre-share secret-risk checks.
package secretscan

// Severity describes the risk level of a finding.
type Severity string

const (
	SeverityHigh   Severity = "HIGH"
	SeverityMedium Severity = "MEDIUM"
)

// FindingType describes where a finding came from.
type FindingType string

const (
	FindingTypePath    FindingType = "path"
	FindingTypeContent FindingType = "content"
)

// Finding is a sanitized secret-risk result.
//
// It intentionally does not include matched content or secret values.
type Finding struct {
	Severity Severity
	Type     FindingType
	Path     string
	Rule     string
}

// Rule describes one scanner rule.
type Rule struct {
	Name     string
	Severity Severity
}

// Entry is the tree metadata needed by the scanner.
type Entry struct {
	Path string
	Type string
	Size int
}

// Options controls scanner behavior.
type Options struct {
	Disabled        bool
	ScanContent     bool
	MaxContentBytes int
}

// Scanner scans repository tree metadata and optional file content.
type Scanner struct {
	disabled        bool
	scanContent     bool
	maxContentBytes int
	pathRules       []pathRule
	contentRules    []contentRule
}
