package authz

import (
	"context"
	"time"
)

type Service interface {
	IssueAuthorizationToken(context.Context, IssueAuthorizationTokenRequest) (*IssueAuthorizationTokenResponse, error)
	ParseAuthorizationToken(context.Context, ParseAuthorizationTokenRequest) error
	RevokeAuthorizationToken(context.Context, RevokeAuthorizationTokenRequest) error
}

type IssueAuthorizationTokenRequest struct {
	TokenType TokenType
	Claims    interface{}
	UserID    string
}

type IssueAuthorizationTokenResponse struct {
	Token     string
	TokenID   string
	ExpiresAt time.Time
}

type RevokeAuthorizationTokenRequest struct {
	TokenID string
}

type RevokeAuthorizationTokenResponse struct {
}

type ParseAuthorizationTokenRequest struct {
	Token  string
	Claims interface{}
}
