package session

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"
)

const (
	cookieName   = "gh_relay_session"
	tokenByteLen = 32
)

func NewManager(ttl time.Duration, done <-chan struct{}) *Manager {
	m := &Manager{
		tokens: make(map[string]time.Time),
		ttl:    ttl,
	}
	go m.reap(done)
	return m
}

// Issue creates a new session token and returns it.
func (m *Manager) Issue() (string, error) {
	raw := make([]byte, tokenByteLen)
	if _, err := rand.Read(raw); err != nil {
		return "", fmt.Errorf("generating session token: %w", err)
	}
	token := hex.EncodeToString(raw)

	m.mu.Lock()
	m.tokens[token] = time.Now().Add(m.ttl)
	m.mu.Unlock()

	return token, nil
}

func (m *Manager) Valid(token string) bool {
	if token == "" {
		return false
	}
	// I don't really think the lock contention here is worth optimizing, but we can always revisit if needed.
	m.mu.RLock()
	expiry, ok := m.tokens[token]
	m.mu.RUnlock()
	return ok && time.Now().Before(expiry)
}

func (m *Manager) Revoke(token string) {
	m.mu.Lock()
	delete(m.tokens, token)
	m.mu.Unlock()
}

func (m *Manager) Count() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	n := 0
	now := time.Now()
	for _, exp := range m.tokens {
		if now.Before(exp) {
			n++
		}
	}
	return n
}

// reap removes expired tokens on a ticker until done is closed.
func (m *Manager) reap(done <-chan struct{}) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-done:
			return
		case <-ticker.C:
			m.mu.Lock()
			now := time.Now()
			for tok, exp := range m.tokens {
				if now.After(exp) {
					delete(m.tokens, tok)
				}
			}
			m.mu.Unlock()
		}
	}
}
