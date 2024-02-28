package runner

import "time"

type Config struct {
	Addr string `yaml:"addr"`

	AuthCfg AuthConfig `yaml:"auth-tokens"`

	OAuthProviders OauthProviderConfig `yaml:"oauth-providers"`
}

type AuthConfig struct {
	SecretKey     string        `yaml:"secret-key"`
	ValidDuration time.Duration `yaml:"valid-duration"`
}

type OauthProviderConfig struct {
	Github *GithubOauthConfig `yaml:"github"`
}

type GithubOauthConfig struct {
	ClientID     string `yaml:"client-id"`
	ClientSecret string `yaml:"client-secret"`
}
