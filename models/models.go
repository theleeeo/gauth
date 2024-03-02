package models

type User struct {
	// The user's ID in the system
	ID string `json:"id"`
	// The user's prefered nickname
	Nickname string `json:"nickname"`

	Role Role `json:"role"`

	// Through which provider the user is authenticated
	Provider UserProvider `json:"providers"`
}

type Role string

const (
	RoleUser  Role = "user"
	RoleAdmin Role = "admin"
)

type UserProviderType string

const (
	UserProviderTypeGithub UserProviderType = "github"
)

type UserProvider struct {
	// By what provider the user is authenticated
	Type UserProviderType `json:"user-provider-type"`
	// The user's ID in the provider's system
	UserID string `json:"user-id"`
}
