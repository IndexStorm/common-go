package authn

import (
	"context"
	"github.com/IndexStorm/common-go/service/authn/idp"
	"golang.org/x/oauth2"
	"net/url"
	"time"
)

type IdentityProvider interface {
	GetAuthorizationURL(ctx context.Context, provider idp.ID, state string) (string, error)
	ExchangeCode(ctx context.Context, provider idp.ID, code string) (*oauth2.Token, error)
	GetUserInfo(ctx context.Context, provider idp.ID, token string) (*idp.UserInfo, error)
	VerifyIdToken(ctx context.Context, provider idp.ID, token string) (*idp.UserInfo, error)
}

type idpAuthenticationCtx struct {
	Provider    idp.ID
	RedirectURL *url.URL
	Nonce       string
	ExpiresAt   time.Time
}

func (c idpAuthenticationCtx) IsExpired(now time.Time) bool {
	return !c.ExpiresAt.After(now)
}
