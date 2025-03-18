package authn

import (
	"context"
	"github.com/IndexStorm/common-go/service/authn/idp"
	"net/url"
)

type Service interface {
	LoginWithProvider(ctx context.Context, req LoginWithProviderRequest) (*LoginWithProviderResponse, error)
	ConsumeProvider(ctx context.Context, req ConsumeProviderRequest) (*ConsumeProviderResponse, error)
	ExchangeCode(ctx context.Context, req ExchangeCodeRequest) (*ExchangeCodeResponse, error)
	VerifyIdToken(ctx context.Context, req VerifyIdTokenRequest) (*VerifyIdTokenResponse, error)
}

type LoginWithProviderRequest struct {
	Provider    idp.ID
	RedirectURL string
	Nonce       string
}

type LoginWithProviderResponse struct {
	AuthenticationURL string
}

type ConsumeProviderRequest struct {
	State string
	Code  string
}

type ConsumeProviderResponse struct {
	Provider    idp.ID
	UserInfo    *idp.UserInfo
	RedirectURL *url.URL
	Nonce       string
}

type ExchangeCodeRequest struct {
	Provider idp.ID
	Code     string
}

type ExchangeCodeResponse struct {
	UserInfo *idp.UserInfo
}

type VerifyIdTokenRequest struct {
	Provider idp.ID
	Token    string
}

type VerifyIdTokenResponse struct {
	UserInfo *idp.UserInfo
}
