package session

import (
	"testing"
	"time"
)

func TestManager_Issue(t *testing.T) {
	done := make(chan struct{})
	defer close(done)

	m := NewManager(time.Hour, done)

	token, err := m.Issue()
	if err != nil {
		t.Fatalf("Issue() error = %v", err)
	}
	if token == "" {
		t.Fatal("expected non-empty token")
	}

	if !m.Valid(token) {
		t.Fatal("expected issued token to be valid")
	}
}

func TestManager_IssueMultipleTokens(t *testing.T) {
	done := make(chan struct{})
	defer close(done)

	m := NewManager(time.Hour, done)

	tokens := make(map[string]bool)
	for i := 0; i < 100; i++ {
		tok, err := m.Issue()
		if err != nil {
			t.Fatalf("Issue() error = %v", err)
		}
		tokens[tok] = true
	}

	if len(tokens) != 100 {
		t.Fatalf("expected 100 unique tokens, got %d", len(tokens))
	}
}

func TestManager_Valid_EmptyToken(t *testing.T) {
	done := make(chan struct{})
	defer close(done)

	m := NewManager(time.Hour, done)

	if m.Valid("") {
		t.Fatal("expected empty token to be invalid")
	}
}

func TestManager_Valid_UnknownToken(t *testing.T) {
	done := make(chan struct{})
	defer close(done)

	m := NewManager(time.Hour, done)

	if m.Valid("unknown-token") {
		t.Fatal("expected unknown token to be invalid")
	}
}

func TestManager_Revoke(t *testing.T) {
	done := make(chan struct{})
	defer close(done)

	m := NewManager(time.Hour, done)

	token, _ := m.Issue()
	if !m.Valid(token) {
		t.Fatal("expected token to be valid before revoke")
	}

	m.Revoke(token)

	if m.Valid(token) {
		t.Fatal("expected token to be invalid after revoke")
	}
}

func TestManager_Count(t *testing.T) {
	done := make(chan struct{})
	defer close(done)

	m := NewManager(time.Hour, done)

	if m.Count() != 0 {
		t.Fatalf("expected 0 tokens initially, got %d", m.Count())
	}

	m.Issue()
	m.Issue()
	m.Issue()

	if m.Count() != 3 {
		t.Fatalf("expected 3 tokens, got %d", m.Count())
	}
}

func TestManager_Valid_ExpiredToken(t *testing.T) {
	done := make(chan struct{})
	defer close(done)

	m := NewManager(time.Nanosecond, done)

	token, _ := m.Issue()

	time.Sleep(time.Millisecond)

	if m.Valid(token) {
		t.Fatal("expected expired token to be invalid")
	}
}

func TestManager_Count_ExcludesExpired(t *testing.T) {
	done := make(chan struct{})
	defer close(done)

	m := NewManager(time.Nanosecond, done)

	m.Issue()
	token, _ := m.Issue()
	m.Issue()

	time.Sleep(time.Millisecond)

	m.Revoke(token)

	if m.Count() != 0 {
		t.Fatalf("expected 0 active tokens after expiry, got %d", m.Count())
	}
}