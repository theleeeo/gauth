package repo

import (
	"context"
	"database/sql"
	"fmt"

	_ "github.com/go-sql-driver/mysql"
	"github.com/theleeeo/thor/models"
)

var _ Repo = &mySqlRepo{}

type MySqlConfig struct {
	Addr     string `yaml:"addr"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	Database string `yaml:"database"`
}

type mySqlRepo struct {
	db *sql.DB
}

// NewMySql creates a repo implementation for MariaDB.
// The repo must be closed after use.
func NewMySql(cfg *MySqlConfig) *mySqlRepo {
	db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s)/%s", cfg.User, cfg.Password, cfg.Addr, cfg.Database))
	if err != nil {
		panic(err.Error())
	}

	return &mySqlRepo{
		db: db,
	}
}

func (r *mySqlRepo) Close() error {
	return r.db.Close()
}

// CreateUser implements Repo.
func (r *mySqlRepo) CreateUser(ctx context.Context, user *models.User) error {
	query := "INSERT INTO users (id, nickname, role, provider, provider_id) VALUES(?, ?, ?, ?, ?)"
	_, err := r.db.ExecContext(ctx, query, user.ID, user.Nickname, user.Role, user.Provider.Type, user.Provider.UserID)
	if err != nil {
		return err
	}

	return nil
}

// GetUser implements Repo.
func (r *mySqlRepo) GetUser(ctx context.Context, id string) (*models.User, error) {
	query := "SELECT id, nickname, role, provider, provider_id FROM users WHERE id = ?"
	row := r.db.QueryRowContext(ctx, query, id)

	var user models.User
	err := row.Scan(&user.ID, &user.Nickname, &user.Role, &user.Provider.Type, &user.Provider.UserID)
	if err != nil {
		return nil, err
	}

	return &user, nil
}
