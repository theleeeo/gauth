package models

type User struct {
	// The user's ID in the system
	ID   string `json:"id"`
	Name string `json:"name"`
	// The user's email
	Email string `json:"email"`

	Role Role `json:"role"`

	// Through which providers the user is authenticated
	Providers []UserProvider `json:"providers"`
}

type Role string

const (
	RoleUser  Role = "user"
	RoleAdmin Role = "admin"
)

type UserProviderType string

const (
	UserProviderTypeGithub UserProviderType = "github"
	UserProviderTypeGoogle UserProviderType = "google"
)

type UserProvider struct {
	// By what provider the user is authenticated
	Type UserProviderType `json:"user-provider-type"`
	// The user's ID in the provider's system
	UserID string `json:"user-id"`
}
