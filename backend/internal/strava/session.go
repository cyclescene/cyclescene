package strava

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"sync"
	"time"
)

// stateContext holds the state token expiry and associated city code
type stateContext struct {
	ExpiresAt time.Time
	CityCode  string
}

// SessionStore manages ephemeral OAuth sessions in-memory
type SessionStore struct {
	sessions sync.Map // map[string]*Session
	states   sync.Map // map[string]*stateContext for CSRF protection with city context
}

// NewSessionStore creates a new session store
func NewSessionStore() *SessionStore {
	store := &SessionStore{}

	// Start background cleanup goroutine
	go store.cleanupExpiredSessions()

	return store
}

// GenerateState creates a secure random state token for CSRF protection
// This is a convenience method that calls GenerateStateWithCity with empty cityCode
func (s *SessionStore) GenerateState() (string, error) {
	return s.GenerateStateWithCity("")
}

// GenerateStateWithCity creates a secure random state token with associated city code
func (s *SessionStore) GenerateStateWithCity(cityCode string) (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}

	state := base64.URLEncoding.EncodeToString(b)

	// Store state with 10 minute expiration and city context
	s.states.Store(state, &stateContext{
		ExpiresAt: time.Now().Add(10 * time.Minute),
		CityCode:  cityCode,
	})

	return state, nil
}

// ValidateState checks if a state token is valid and removes it
// This is a convenience method that ignores the city code
func (s *SessionStore) ValidateState(state string) bool {
	_, valid := s.ValidateStateAndGetCity(state)
	return valid
}

// ValidateStateAndGetCity validates the state token and returns the associated city code
// Returns (cityCode, valid) - cityCode may be empty if not set during state generation
func (s *SessionStore) ValidateStateAndGetCity(state string) (string, bool) {
	value, ok := s.states.LoadAndDelete(state)
	if !ok {
		return "", false
	}

	// Handle both old format (time.Time) and new format (*stateContext)
	switch ctx := value.(type) {
	case *stateContext:
		if time.Now().After(ctx.ExpiresAt) {
			return "", false
		}
		return ctx.CityCode, true
	case time.Time:
		// Backwards compatibility with old format
		if time.Now().After(ctx) {
			return "", false
		}
		return "", true
	default:
		return "", false
	}
}

// CreateSession creates a new session and returns the session ID
func (s *SessionStore) CreateSession(session *Session) (string, error) {
	// Generate session ID
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}

	sessionID := base64.URLEncoding.EncodeToString(b)
	session.SessionID = sessionID
	session.CreatedAt = time.Now()

	// Store session
	s.sessions.Store(sessionID, session)

	// Auto-cleanup after 1 hour
	time.AfterFunc(1*time.Hour, func() {
		s.sessions.Delete(sessionID)
	})

	return sessionID, nil
}

// GetSession retrieves a session by ID
func (s *SessionStore) GetSession(sessionID string) (*Session, bool) {
	value, ok := s.sessions.Load(sessionID)
	if !ok {
		return nil, false
	}

	session, ok := value.(*Session)
	if !ok {
		return nil, false
	}

	// Check if token expired
	if time.Now().After(session.ExpiresAt) {
		s.sessions.Delete(sessionID)
		return nil, false
	}

	return session, true
}

// DeleteSession removes a session
func (s *SessionStore) DeleteSession(sessionID string) {
	s.sessions.Delete(sessionID)
}

// cleanupExpiredSessions runs periodically to remove expired sessions
func (s *SessionStore) cleanupExpiredSessions() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		now := time.Now()

		// Cleanup expired sessions
		s.sessions.Range(func(key, value interface{}) bool {
			session, ok := value.(*Session)
			if ok && now.After(session.ExpiresAt.Add(1*time.Hour)) {
				s.sessions.Delete(key)
			}
			return true
		})

		// Cleanup expired states
		s.states.Range(func(key, value interface{}) bool {
			switch ctx := value.(type) {
			case *stateContext:
				if now.After(ctx.ExpiresAt) {
					s.states.Delete(key)
				}
			case time.Time:
				if now.After(ctx) {
					s.states.Delete(key)
				}
			}
			return true
		})
	}
}

// GetStats returns statistics about the session store
func (s *SessionStore) GetStats() map[string]int {
	sessionCount := 0
	stateCount := 0

	s.sessions.Range(func(key, value interface{}) bool {
		sessionCount++
		return true
	})

	s.states.Range(func(key, value interface{}) bool {
		stateCount++
		return true
	})

	return map[string]int{
		"sessions": sessionCount,
		"states":   stateCount,
	}
}

// Debug helper
func (s *SessionStore) String() string {
	stats := s.GetStats()
	return fmt.Sprintf("SessionStore{sessions: %d, states: %d}", stats["sessions"], stats["states"])
}
