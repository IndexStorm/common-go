package authn

import (
	"context"
	"fmt"
	"github.com/IndexStorm/common-go/service/authn/idp"
	"github.com/imroc/req/v3"
	"golang.org/x/oauth2"
	oauthgoogle "golang.org/x/oauth2/google"
)

type oauthIDP struct {
	client *req.Client
	google *oauth2.Config
}

type GoogleOAuthConfig struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
	Scopes       []string
}

type OAuthIDPConfig struct {
	Google          GoogleOAuthConfig
	ClientUserAgent string
}

func NewOAuthIDP(config OAuthIDPConfig) IdentityProvider {
	return &oauthIDP{
		client: req.C().SetUserAgent(config.ClientUserAgent),
		google: &oauth2.Config{
			RedirectURL:  config.Google.RedirectURL,
			ClientID:     config.Google.ClientID,
			ClientSecret: config.Google.ClientSecret,
			Scopes:       config.Google.Scopes,
			Endpoint:     oauthgoogle.Endpoint,
		},
	}
}

func (p *oauthIDP) GetAuthorizationURL(
	ctx context.Context, provider idp.ID, state string,
) (string, error) {
	switch provider {
	case idp.GoogleID:
		return p.google.AuthCodeURL(state, oauth2.AccessTypeOffline), nil
	default:
		return "", fmt.Errorf("unsupported provider: %s", provider)
	}
}

func (p *oauthIDP) ExchangeCode(
	ctx context.Context, provider idp.ID, code string,
) (*oauth2.Token, error) {
	switch provider {
	case idp.GoogleID:
		oauthToken, err := p.google.Exchange(ctx, code)
		if err != nil {
			return nil, fmt.Errorf("exchange code: %w", err)
		}
		return oauthToken, nil
	default:
		return nil, fmt.Errorf("unsupported provider: %s", provider)
	}
}

func (p *oauthIDP) GetUserInfo(
	ctx context.Context, provider idp.ID, token string,
) (*idp.UserInfo, error) {
	switch provider {
	case idp.GoogleID:
		var response idp.GoogleUserInfo
		resp, err := p.client.R().
			SetContext(ctx).
			SetQueryParam("access_token", token).
			SetSuccessResult(&response).
			Post("https://www.googleapis.com/oauth2/v3/userinfo")
		if err != nil {
			return nil, err
		}
		if resp.StatusCode != 200 {
			return nil, fmt.Errorf("unexpected status: %d. %s", resp.StatusCode, resp.String())
		}
		return &idp.UserInfo{
			ID:               response.Sub,
			Email:            response.Email,
			EmailVerified:    response.EmailVerified,
			ProviderID:       provider,
			ProviderUserInfo: response,
		}, nil
	default:
		return nil, fmt.Errorf("unsupported provider: %s", provider)
	}
}

func (p *oauthIDP) VerifyIdToken(ctx context.Context, provider idp.ID, token string) (*idp.UserInfo, error) {
	return nil, fmt.Errorf("unsupported provider: %s", provider)
}
