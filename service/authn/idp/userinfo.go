package idp

type UserInfo struct {
	ID            string
	Email         string
	EmailVerified bool

	ProviderID       ID
	ProviderUserInfo interface{}
}

type GoogleUserInfo struct {
	Sub           string `json:"sub"`
	Name          string `json:"name"`
	GivenName     string `json:"given_name"`
	FamilyName    string `json:"family_name"`
	Picture       string `json:"picture"`
	Email         string `json:"email"`
	EmailVerified bool   `json:"email_verified"`
}
