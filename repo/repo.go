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
	CreateUser(ctx context.Context, user *models.User) error
	GetUser(ctx context.Context, params GetUserParams) (*models.User, error)
	GetUserByProviderID(ctx context.Context, providerID string) (*models.User, error)
}

type GetUserParams struct {
	Email *string
	ID    *string
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
