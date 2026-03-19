// Package commands provides the Discord slash command implementations for the GuildMessageProxy bot.
package commands

import (
	"sync"
	"time"
)

// DraftService encapsulates draft storage with thread-safe operations.
type DraftService struct {
	mu    sync.RWMutex
	store map[string]*Draft
	ttl   time.Duration
}

// NewDraftService creates a DraftService with default TTL.
func NewDraftService() *DraftService {
	return &DraftService{
		store: make(map[string]*Draft),
		ttl:   DraftTTL,
	}
}

// getDraftKey generates a unique key for user's draft in a guild.
func getDraftKey(userID, guildID string) string {
	return userID + ":" + guildID
}

// Get retrieves a draft by userID and guildID.
// Returns the draft and a bool indicating if it was found.
func (s *DraftService) Get(userID, guildID string) (*Draft, bool) {
	var key string = getDraftKey(userID, guildID)
	s.mu.RLock()
	defer s.mu.RUnlock()
	var draft, exists = s.store[key]
	return draft, exists
}

// Save stores a draft.
func (s *DraftService) Save(draft *Draft) {
	var key string = getDraftKey(draft.UserID, draft.GuildID)
	s.mu.Lock()
	defer s.mu.Unlock()
	s.store[key] = draft
}

// Delete removes a draft.
func (s *DraftService) Delete(userID, guildID string) {
	var key string = getDraftKey(userID, guildID)
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.store, key)
}

// CleanupExpired removes expired drafts and returns count cleaned.
func (s *DraftService) CleanupExpired() int {
	var now = time.Now()
	var cleaned int

	s.mu.Lock()
	defer s.mu.Unlock()
	for key, draft := range s.store {
		if now.After(draft.ExpiresAt) {
			delete(s.store, key)
			cleaned++
		}
	}

	return cleaned
}
