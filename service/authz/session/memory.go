package session

import (
	"context"
	"github.com/IndexStorm/common-go/store"
	"time"
)

type memoryService struct {
	sessions *store.ExpirableStore[expirableSession]
}

type expirableSession struct {
	SessionID string
	UserID    string
	ExpiresAt time.Time
}

func NewMemoryService() Service {
	return &memoryService{
		sessions: store.NewExpiringStore[expirableSession](),
	}
}

func (s expirableSession) IsExpired(now time.Time) bool {
	return !s.ExpiresAt.After(now)
}

func (s *memoryService) InitializeSession(ctx context.Context, request InitializeSessionRequest) error {
	s.sessions.Set(request.SessionID, expirableSession{
		SessionID: request.SessionID,
		UserID:    request.UserID,
		ExpiresAt: request.ExpiresAt,
	})
	return nil
}

func (s *memoryService) ValidateSession(ctx context.Context, request ValidateSessionRequest) error {
	session, found := s.sessions.Get(request.SessionID)
	if !found {
		return ErrSessionNotFound
	}
	if time.Now().After(session.ExpiresAt) {
		return ErrSessionExpired
	}
	return nil
}

func (s *memoryService) InvalidateSession(ctx context.Context, request InvalidateSessionRequest) error {
	s.sessions.Delete(request.SessionID)
	return nil
}
