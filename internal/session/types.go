package session

import (
	"sync"
	"time"
)

type Manager struct {
	mu     sync.RWMutex
	tokens map[string]time.Time // token:expiry
	ttl    time.Duration
}
