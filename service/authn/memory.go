package authn

import (
	"context"
	"fmt"
	"github.com/IndexStorm/common-go/nanoid"
	"github.com/IndexStorm/common-go/store"
	"io"
	"net/url"
	"time"
)

type InMemoryService interface {
	io.Closer
	Service
}

type inMemoryService struct {
	store *store.ExpirableStore[idpAuthenticationCtx]
	idp   IdentityProvider
}

func NewInMemoryService(idp IdentityProvider) InMemoryService {
	return &inMemoryService{
		store: store.NewExpiringStore[idpAuthenticationCtx](),
		idp:   idp,
	}
}

func (s *inMemoryService) Close() error {
	if err := s.store.Close(); err != nil {
		return fmt.Errorf("close store: %w", err)
	}
	return nil
}

func (s *inMemoryService) LoginWithProvider(
	ctx context.Context, req LoginWithProviderRequest,
) (*LoginWithProviderResponse, error) {
	redirectURL, err := url.Parse(req.RedirectURL)
	if err != nil {
		return nil, fmt.Errorf("parse redirect URL: %w", err)
	}
	authCtx := idpAuthenticationCtx{
		Provider:    req.Provider,
		RedirectURL: redirectURL,
		Nonce:       req.Nonce,
		ExpiresAt:   time.Now().Add(time.Minute * 5),
	}
	state := nanoid.RandomLongID()
	authURL, err := s.idp.GetAuthorizationURL(ctx, req.Provider, state)
	if err != nil {
		return nil, fmt.Errorf("get authorization URL: %w", err)
	}
	s.store.Set(state, authCtx)
	return &LoginWithProviderResponse{AuthenticationURL: authURL}, nil
}

func (s *inMemoryService) ConsumeProvider(ctx context.Context, req ConsumeProviderRequest) (
	*ConsumeProviderResponse, error,
) {
	authCtx, ok := s.store.Pop(req.State)
	if !ok {
		return nil, fmt.Errorf("state not found")
	}
	oauthToken, err := s.idp.ExchangeCode(ctx, authCtx.Provider, req.Code)
	if err != nil {
		return nil, fmt.Errorf("exchange code: %w", err)
	}
	userInfo, err := s.idp.GetUserInfo(ctx, authCtx.Provider, oauthToken.AccessToken)
	if err != nil {
		return nil, fmt.Errorf("get user info: %w", err)
	}
	return &ConsumeProviderResponse{
		Provider:    authCtx.Provider,
		UserInfo:    userInfo,
		RedirectURL: authCtx.RedirectURL,
		Nonce:       authCtx.Nonce,
	}, nil
}

func (s *inMemoryService) ExchangeCode(ctx context.Context, req ExchangeCodeRequest) (*ExchangeCodeResponse, error) {
	oauthToken, err := s.idp.ExchangeCode(ctx, req.Provider, req.Code)
	if err != nil {
		return nil, fmt.Errorf("exchange code: %w", err)
	}
	userInfo, err := s.idp.GetUserInfo(ctx, req.Provider, oauthToken.AccessToken)
	if err != nil {
		return nil, fmt.Errorf("get user info: %w", err)
	}
	return &ExchangeCodeResponse{UserInfo: userInfo}, nil
}

func (s *inMemoryService) VerifyIdToken(ctx context.Context, req VerifyIdTokenRequest) (*VerifyIdTokenResponse, error) {
	userInfo, err := s.idp.VerifyIdToken(ctx, req.Provider, req.Token)
	if err != nil {
		return nil, fmt.Errorf("verify google id token: %w", err)
	}
	return &VerifyIdTokenResponse{UserInfo: userInfo}, nil
}
