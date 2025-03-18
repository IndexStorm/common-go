package session

import (
	"context"
	"time"
)

type Service interface {
	InitializeSession(ctx context.Context, request InitializeSessionRequest) error
	ValidateSession(ctx context.Context, request ValidateSessionRequest) error
	InvalidateSession(ctx context.Context, request InvalidateSessionRequest) error
}

type InitializeSessionRequest struct {
	SessionID string
	UserID    string
	ExpiresAt time.Time
}

type ValidateSessionRequest struct {
	SessionID string
}

type InvalidateSessionRequest struct {
	SessionID string
}
