package repo

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/go-sql-driver/mysql"
	"github.com/theleeeo/thor/models"
)

const (
	mysqlErrDuplicateEntry = 1062
	mysqlErrDuplicateKey   = 1169
)

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

func (r *mySqlRepo) Ping() error {
	return r.db.Ping()
}

// CreateUser implements Repo.
func (r *mySqlRepo) CreateUser(ctx context.Context, user *models.User) error {
	query := "INSERT INTO users (id, nickname, role, provider, provider_id) VALUES(?, ?, ?, ?, ?)"
	_, err := r.db.ExecContext(ctx, query, user.ID, user.Nickname, user.Role, user.Provider.Type, user.Provider.UserID)
	if err != nil {
		if e, ok := err.(*mysql.MySQLError); ok && (e.Number == mysqlErrDuplicateEntry || e.Number == mysqlErrDuplicateKey) {
			return ErrAlreadyExists
		}
		return err
	}

	return nil
}

// GetUserByID implements Repo.
func (r *mySqlRepo) GetUserByID(ctx context.Context, id string) (*models.User, error) {
	query := "SELECT id, nickname, role, provider, provider_id FROM users WHERE id = ?"
	row := r.db.QueryRowContext(ctx, query, id)

	var user models.User
	err := row.Scan(&user.ID, &user.Nickname, &user.Role, &user.Provider.Type, &user.Provider.UserID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, err
	}

	return &user, nil
}

func (r *mySqlRepo) GetUserByProviderID(ctx context.Context, providerID string) (*models.User, error) {
	query := "SELECT id, nickname, role, provider, provider_id FROM users WHERE provider_id = ?"
	row := r.db.QueryRowContext(ctx, query, providerID)

	var user models.User
	err := row.Scan(&user.ID, &user.Nickname, &user.Role, &user.Provider.Type, &user.Provider.UserID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, err
	}

	return &user, nil
}
