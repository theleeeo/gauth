package oauth

type Config struct {
	// The URL of the app, used for redirecting after OAuth login
	AppURL         string           `yaml:"app-url"`
	CookieName     string           `yaml:"cookie-name"`
	SessionName    string           `yaml:"session-name"`
	CookieSecret   string           `yaml:"cookie-secret"`
	AllowedReturns []string         `yaml:"allowed-returns"`
	Providers      []ProviderConfig `yaml:"providers"`
}

type ProviderType string

const (
	GithubProviderType ProviderType = "github"
	GoogleProviderType ProviderType = "google"
)

type ProviderConfig struct {
	Type         ProviderType `yaml:"type"`
	Name         string       `yaml:"name"`
	ClientID     string       `yaml:"client-id"`
	ClientSecret string       `yaml:"client-secret"`
}
