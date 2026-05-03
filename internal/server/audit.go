package server

import (
	"log"
	"net/http"
	"strings"
	"sync"
	"time"
)

type AuditRecord struct {
	timestamp time.Time
	endpoint  string
	filePath  string
	branch    string
	ipAddress string
}

type AuditLog struct {
	records []AuditRecord
	mu      sync.Mutex
}

func (al *AuditLog) add(record AuditRecord) {
	al.mu.Lock()
	al.records = append(al.records, record)
	al.mu.Unlock()
}

func (al *AuditLog) Summary(l *log.Logger) {
	al.mu.Lock()
	defer al.mu.Unlock()

	if len(al.records) == 0 {
		l.Println("  [audit] No guest activity recorded.")
		return
	}

	blobCount := 0
	uniqueFiles := make(map[string]struct{})
	uniqueIPs := make(map[string]struct{})
	branches := make(map[string]struct{})

	for _, r := range al.records {
		uniqueIPs[r.ipAddress] = struct{}{}
		if r.branch != "" {
			branches[r.branch] = struct{}{}
		}
		if r.endpoint == "/api/blob" {
			blobCount++
			uniqueFiles[r.filePath] = struct{}{}
		}
	}

	first := al.records[0].timestamp
	last := al.records[len(al.records)-1].timestamp
	duration := last.Sub(first)

	l.Println()
	l.Println(strings.Repeat("-", 54))
	l.Println("  SESSION AUDIT SUMMARY")
	l.Println(strings.Repeat("-", 54))
	l.Printf("  Files viewed  : %d (%d unique)", blobCount, len(uniqueFiles))
	l.Printf("  Total requests: %d", len(al.records))
	l.Printf("  Unique IPs    : %d", len(uniqueIPs))
	if len(branches) > 0 {
		branchList := make([]string, 0, len(branches))
		for b := range branches {
			branchList = append(branchList, b)
		}
		l.Printf("  Branches      : %s", strings.Join(branchList, ", "))
	}
	l.Printf("  Duration      : %s", formatDuration(duration))
	l.Println(strings.Repeat("-", 54))
}

func formatDuration(d time.Duration) string {
	return d.Truncate(time.Second).String()
}

// statusRecorder wraps http.ResponseWriter to capture the status code.
type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(code int) {
	r.status = code
	r.ResponseWriter.WriteHeader(code)
}

func (r *statusRecorder) Flush() {
	if f, ok := r.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}
