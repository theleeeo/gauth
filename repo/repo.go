package repo

import (
	"context"
	"fmt"
	"log"

	"github.com/theleeeo/thor/models"
)

type Repo interface {
	// Perform the necessary clean up before the repo is discarded.
	Close() error
	Ping() error
	// Migrate() error

	// User
	CreateUser(ctx context.Context, user models.User, provider models.UserProvider) error
	GetUser(ctx context.Context, params GetUserParams) (models.User, error)
	ListUsers(ctx context.Context, params ListUsersParams) ([]models.User, error)
	GetUserByProviderID(ctx context.Context, providerID string) (models.User, error)
	AddProvider(ctx context.Context, userID string, provider models.UserProvider) error
	AssignRole(ctx context.Context, userID string, roleID string) error
	RemoveRole(ctx context.Context, userID string, roleID string) error
	GetProvidersOfUser(ctx context.Context, userID string) ([]models.UserProvider, error)
	GetPermissionsOfUser(ctx context.Context, userID string) ([]models.Permission, error)

	// Role
	CreateRole(ctx context.Context, role models.Role, permissions []models.Permission) error
	GetRole(ctx context.Context, id string) (models.Role, error)
	ListRoles(ctx context.Context, params ListRolesParams) ([]models.Role, error)
	GetRolesOfUser(ctx context.Context, userID string) ([]models.Role, error)
	GetPermissionsOfRole(ctx context.Context, roleID string) ([]models.Permission, error)
}

type GetUserParams struct {
	ID    *string
	Email *string
}

type ListRolesParams struct {
}

type ListUsersParams struct {
}

type Config struct {
	MySql *MySqlConfig `yaml:"mysql"`
}

func New(cfg *Config) (Repo, error) {
	repo, err := createRepo(cfg)
	if err != nil {
		return nil, fmt.Errorf("create repo failed: %w", err)
	}

	// if err := repo.Migrate(); err != nil {
	// 	return err
	// }

	if err := repo.Ping(); err != nil {
		return nil, fmt.Errorf("repo ping failed: %w", err)
	}

	return repo, nil
}

func createRepo(cfg *Config) (Repo, error) {
	if cfg == nil {
		return nil, fmt.Errorf("no repo configuration found")
	}

	if cfg.MySql != nil {
		log.Println("using mysql as repo")
		return NewMySql(cfg.MySql), nil
	}

	return nil, fmt.Errorf("no repo configuration found")
}
