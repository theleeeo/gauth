package runner

import (
	"time"

	"github.com/theleeeo/thor/oauth"
	"github.com/theleeeo/thor/repo"
)

type Config struct {
	Addr string `yaml:"addr"`

	AuthCfg AuthConfig `yaml:"auth-tokens"`

	RepoCfg *repo.Config `yaml:"repo"`

	OAuthConfig *oauth.Config `yaml:"oauth"`
}

type AuthConfig struct {
	SecretKey     string        `yaml:"secret-key"`
	ValidDuration time.Duration `yaml:"valid-duration"`
}
