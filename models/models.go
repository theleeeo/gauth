package models

type User struct {
	// The user's ID in the system
	ID   string `json:"id"`
	Name string `json:"name"`
	// The user's email
	Email string `json:"email"`
}

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

type Role struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type Permission struct {
	Key string
	Val string
}
