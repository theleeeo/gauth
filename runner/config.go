package runner

import (
	"time"

	"github.com/theleeeo/thor/oauth"
	"github.com/theleeeo/thor/repo"
)

type Config struct {
	Addr string `yaml:"addr"`

	// The URL of the app, used for redirecting after OAuth login
	AppURL string `yaml:"app-url"`

	AuthCfg AuthConfig `yaml:"auth-tokens"`

	RepoCfg *repo.Config `yaml:"repo"`

	OAuthProviders []oauth.ProviderConfig `yaml:"oauth-providers"`
}

type AuthConfig struct {
	SecretKey     string        `yaml:"secret-key"`
	ValidDuration time.Duration `yaml:"valid-duration"`
}
