package oauth

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
