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
	GetUserByID(ctx context.Context, id string) (*models.User, error)
	GetUserByProviderID(ctx context.Context, providerID string) (*models.User, error)
}

type Config struct {
	MySql *MySqlConfig `yaml:"mysql"`
}

func New(cfg *Config) (Repo, error) {
	if cfg == nil {
		return nil, fmt.Errorf("no repo configuration found")
	}

	if cfg.MySql != nil {
		log.Println("using mysql as repo")
		return NewMySql(cfg.MySql), nil
	}

	return nil, fmt.Errorf("no repo configuration found")
}
