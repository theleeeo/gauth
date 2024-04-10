package repo

import (
	"context"
	"database/sql"
	"errors"
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
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	userQuery := "INSERT INTO users (id, first_name, last_name, email, role) VALUES(?, ?, ?, ?, ?)"
	_, err = tx.ExecContext(ctx, userQuery, user.ID, user.FirstName, user.LastName, user.Email, user.Role)
	if err != nil {
		tx.Rollback()
		if e, ok := err.(*mysql.MySQLError); ok && (e.Number == mysqlErrDuplicateEntry || e.Number == mysqlErrDuplicateKey) {
			return ErrAlreadyExists
		}
		return err
	}

	for _, p := range user.Providers {
		userProviderQuery := "INSERT INTO user_providers (user_id, provider, provider_id) VALUES(?, ?, ?)"
		_, err = tx.ExecContext(ctx, userProviderQuery, user.ID, p.Type, p.UserID)
		if err != nil {
			tx.Rollback()
			if e, ok := err.(*mysql.MySQLError); ok && (e.Number == mysqlErrDuplicateEntry || e.Number == mysqlErrDuplicateKey) {
				return ErrAlreadyExists
			}
			return err
		}
	}

	err = tx.Commit()
	if err != nil {
		rollbackErr := tx.Rollback()
		if rollbackErr != nil {
			return errors.Join(err, rollbackErr)
		}
		return err
	}

	return nil
}

func (r *mySqlRepo) getProvidersOfUser(ctx context.Context, userID string) ([]models.UserProvider, error) {
	query := "SELECT provider, provider_id FROM user_providers WHERE user_id = ?"
	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}

	var providers []models.UserProvider
	for rows.Next() {
		var p models.UserProvider
		err = rows.Scan(&p.Type, &p.UserID)
		if err != nil {
			rows.Close()
			return nil, err
		}
		providers = append(providers, p)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return providers, nil
}

func (r *mySqlRepo) GetUserByProviderID(ctx context.Context, providerID string) (*models.User, error) {
	query := `SELECT u.id, u.first_name, u.last_name, u.email, u.role
              FROM users u 
              INNER JOIN user_providers up ON u.id = up.user_id 
              WHERE up.provider_id = ?`
	row := r.db.QueryRowContext(ctx, query, providerID)

	var user models.User
	err := row.Scan(&user.ID, &user.FirstName, &user.LastName, &user.Email, &user.Role)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, err
	}

	p, err := r.getProvidersOfUser(ctx, user.ID)
	if err != nil {
		return nil, err
	}
	user.Providers = p

	return &user, nil
}

func (r *mySqlRepo) GetUser(ctx context.Context, params GetUserParams) (*models.User, error) {
	query := "SELECT id, first_name, last_name, email, role FROM users WHERE"
	var args []interface{}

	if params.ID != nil {
		query += " id = ?"
		args = append(args, *params.ID)
	}

	if params.Email != nil {
		query += " email = ?"
		args = append(args, *params.Email)
	}

	row := r.db.QueryRowContext(ctx, query, args...)

	var user models.User
	err := row.Scan(&user.ID, &user.FirstName, &user.LastName, &user.Email, &user.Role)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, err
	}

	p, err := r.getProvidersOfUser(ctx, user.ID)
	if err != nil {
		return nil, err
	}
	user.Providers = p

	return &user, nil
}
