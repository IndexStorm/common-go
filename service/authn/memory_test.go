package authn_test

import (
	"context"
	"github.com/IndexStorm/common-go/service/authn"
	"github.com/IndexStorm/common-go/service/authn/idp"
	"github.com/stretchr/testify/require"
	"net/url"
	"strings"
	"testing"
	"time"
)

func TestInMemoryService_LoginWithProvider_Google(t *testing.T) {
	mainCtx, cancel := context.WithTimeout(context.Background(), time.Second)
	t.Cleanup(cancel)

	const authURL = "https://fake.url/path"
	const redirectURL = "https://app.fake.url/path"
	const clientID = "client_id"
	const accessType = "offline"
	scopes := []string{"scope1", "scope2"}

	service := authn.NewInMemoryService(authn.NewOAuthIDP(&authn.GoogleOAuthConfig{
		ClientID:     "client_id",
		ClientSecret: "client_secret",
		RedirectURL:  authURL,
		Scopes:       scopes,
	}))

	response, err := service.LoginWithProvider(mainCtx, authn.LoginWithProviderRequest{
		Provider:    idp.GoogleID,
		RedirectURL: redirectURL,
		Nonce:       "test",
	})
	require.NoError(t, err)
	require.NotNil(t, response)

	parsedAuthURL, err := url.Parse(response.AuthenticationURL)
	require.NoError(t, err)
	authURLQuery := parsedAuthURL.Query()
	require.NotEmpty(t, authURLQuery.Get("state"))
	require.Equal(t, clientID, authURLQuery.Get("client_id"))
	require.Equal(t, authURL, authURLQuery.Get("redirect_uri"))
	require.Equal(t, accessType, authURLQuery.Get("access_type"))
	require.Equal(t, strings.Join(scopes, " "), authURLQuery.Get("scope"))
}
