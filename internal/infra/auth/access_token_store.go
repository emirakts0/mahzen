package auth

import (
	"crypto/sha256"
	"encoding/hex"
	"sync"
	"time"
)

type tokenEntry struct {
	userID    string
	expiresAt time.Time
	revoked   bool
}

// AccessTokenStore implements domain.AccessTokenStore with an in-memory map.
type AccessTokenStore struct {
	mu     sync.RWMutex
	tokens map[string]*tokenEntry // key: token hash
}

// NewAccessTokenStore creates a new empty in-memory token store.
func NewAccessTokenStore() *AccessTokenStore {
	return &AccessTokenStore{
		tokens: make(map[string]*tokenEntry),
	}
}

// Lookup checks if a raw token is valid and returns the associated user ID.
func (s *AccessTokenStore) Lookup(rawToken string) (string, bool) {
	h := sha256.Sum256([]byte(rawToken))
	hash := hex.EncodeToString(h[:])

	s.mu.RLock()
	entry, ok := s.tokens[hash]
	s.mu.RUnlock()

	if !ok {
		return "", false
	}
	if entry.revoked || time.Now().After(entry.expiresAt) {
		return "", false
	}
	return entry.userID, true
}

// Add inserts a token entry into the in-memory store.
func (s *AccessTokenStore) Add(tokenHash, userID string, expiresAt time.Time) {
	s.mu.Lock()
	s.tokens[tokenHash] = &tokenEntry{
		userID:    userID,
		expiresAt: expiresAt,
	}
	s.mu.Unlock()
}

// Revoke marks a token as revoked in the in-memory store.
func (s *AccessTokenStore) Revoke(tokenHash string) {
	s.mu.Lock()
	if entry, ok := s.tokens[tokenHash]; ok {
		entry.revoked = true
	}
	s.mu.Unlock()
}

// RemoveExpired purges expired entries from the in-memory store.
func (s *AccessTokenStore) RemoveExpired() {
	now := time.Now()
	s.mu.Lock()
	for hash, entry := range s.tokens {
		if now.After(entry.expiresAt) || entry.revoked {
			delete(s.tokens, hash)
		}
	}
	s.mu.Unlock()
}
