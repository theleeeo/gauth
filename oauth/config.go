package oauth

type Config struct {
	// The URL of the app, used for redirecting after OAuth login
	AppURL       string           `yaml:"app-url"`
	CookieSecret string           `yaml:"cookie-secret"`
	Providers    []ProviderConfig `yaml:"providers"`
}

type ProviderType string

const (
	GithubProviderType ProviderType = "github"
)

type ProviderConfig struct {
	Type         ProviderType `yaml:"type"`
	Name         string       `yaml:"name"`
	ClientID     string       `yaml:"client-id"`
	ClientSecret string       `yaml:"client-secret"`
}
