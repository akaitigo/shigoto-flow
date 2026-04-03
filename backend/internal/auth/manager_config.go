package auth

type OAuthManagerConfig struct {
	Google       OAuthClientConfig
	Slack        OAuthClientConfig
	GitHub       OAuthClientConfig
	RedirectBase string
}

type OAuthClientConfig struct {
	ClientID     string
	ClientSecret string
}
