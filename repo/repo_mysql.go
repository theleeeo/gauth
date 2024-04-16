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
func (r *mySqlRepo) CreateUser(ctx context.Context, user models.User, provider models.UserProvider) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	userQuery := "INSERT INTO users (id, name, email) VALUES(?, ?, ?);"
	_, err = tx.ExecContext(ctx, userQuery, user.ID, user.Name, user.Email)
	if err != nil {
		tx.Rollback()
		if e, ok := err.(*mysql.MySQLError); ok && (e.Number == mysqlErrDuplicateEntry || e.Number == mysqlErrDuplicateKey) {
			return ErrAlreadyExists
		}
		return err
	}

	userProviderQuery := "INSERT INTO user_providers (user_id, provider, provider_id) VALUES(?, ?, ?);"
	_, err = tx.ExecContext(ctx, userProviderQuery, user.ID, provider.Type, provider.UserID)
	if err != nil {
		tx.Rollback()
		if e, ok := err.(*mysql.MySQLError); ok && (e.Number == mysqlErrDuplicateEntry || e.Number == mysqlErrDuplicateKey) {
			return ErrAlreadyExists
		}
		return err
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

func (r *mySqlRepo) GetProvidersOfUser(ctx context.Context, userID string) ([]models.UserProvider, error) {
	query := "SELECT provider, provider_id FROM user_providers WHERE user_id = ?;"
	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}

	providers := make([]models.UserProvider, 0)
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

func (r *mySqlRepo) GetUserByProviderID(ctx context.Context, providerID string) (models.User, error) {
	query := `SELECT u.id, u.name, u.email
              FROM users u 
              JOIN user_providers up ON u.id = up.user_id 
              WHERE up.provider_id = ?;`
	row := r.db.QueryRowContext(ctx, query, providerID)

	var user models.User
	err := row.Scan(&user.ID, &user.Name, &user.Email)
	if err != nil {
		if err == sql.ErrNoRows {
			return models.User{}, ErrNotFound
		}
		return models.User{}, err
	}

	// p, err := r.getProvidersOfUser(ctx, user.ID)
	// if err != nil {
	// 	return nil, err
	// }
	// user.Providers = p

	// roles, err := r.GetUserRoles(ctx, user.ID)
	// if err != nil {
	// 	return nil, err
	// }
	// user.Roles = roles

	// permissions, err := r.GetUserPermissions(ctx, user.ID)
	// if err != nil {
	// 	return nil, err
	// }
	// user.Permissions = permissions

	return user, nil
}

func (r *mySqlRepo) GetUser(ctx context.Context, params GetUserParams) (models.User, error) {
	query := "SELECT id, name, email FROM users WHERE"
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
	err := row.Scan(&user.ID, &user.Name, &user.Email)
	if err != nil {
		if err == sql.ErrNoRows {
			return models.User{}, ErrNotFound
		}
		return models.User{}, err
	}

	// p, err := r.getProvidersOfUser(ctx, user.ID)
	// if err != nil {
	// 	return nil, err
	// }
	// user.Providers = p

	// roles, err := r.GetUserRoles(ctx, user.ID)
	// if err != nil {
	// 	return nil, err
	// }
	// user.Roles = roles

	// permissions, err := r.GetUserPermissions(ctx, user.ID)
	// if err != nil {
	// 	return nil, err
	// }
	// user.Permissions = permissions

	return user, nil
}

func (r *mySqlRepo) AddProvider(ctx context.Context, userID string, provider models.UserProvider) error {
	query := "INSERT INTO user_providers (user_id, provider, provider_id) VALUES(?, ?, ?);"
	_, err := r.db.ExecContext(ctx, query, userID, provider.Type, provider.UserID)
	if err != nil {
		if e, ok := err.(*mysql.MySQLError); ok && (e.Number == mysqlErrDuplicateEntry || e.Number == mysqlErrDuplicateKey) {
			return ErrAlreadyExists
		}
		return err
	}

	return nil
}

func (r *mySqlRepo) GetUserRoles(ctx context.Context, userID string) ([]models.Role, error) {
	query := `
		SELECT r.id, r.name
		FROM user_roles ur
		JOIN roles r ON ur.role_id = r.id
		WHERE ur.user_id = ?;
	`
	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	roles := make([]models.Role, 0)
	for rows.Next() {
		var role models.Role
		if err := rows.Scan(&role.ID, &role.Name); err != nil {
			return nil, err
		}

		// permissions, err := r.getRolePermissions(ctx, role.ID)
		// if err != nil {
		// 	return nil, err
		// }
		// role.Permissions = permissions

		roles = append(roles, role)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return roles, nil
}

func (r *mySqlRepo) ListUsers(ctx context.Context, _ ListUsersParams) ([]models.User, error) {
	query := "SELECT id, name, email FROM users;"
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	users := make([]models.User, 0)
	for rows.Next() {
		var user models.User
		if err := rows.Scan(&user.ID, &user.Name, &user.Email); err != nil {
			return nil, err
		}

		// p, err := r.getProvidersOfUser(ctx, user.ID)
		// if err != nil {
		// 	return nil, err
		// }
		// user.Providers = p

		// roles, err := r.GetUserRoles(ctx, user.ID)
		// if err != nil {
		// 	return nil, err
		// }
		// user.Roles = roles

		// permissions, err := r.GetUserPermissions(ctx, user.ID)
		// if err != nil {
		// 	return nil, err
		// }
		// user.Permissions = permissions

		users = append(users, user)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
}

func (r *mySqlRepo) GetPermissionsOfUser(ctx context.Context, userID string) ([]models.Permission, error) {
	query := `
			SELECT p_key, p_val FROM user_roles ur
			JOIN role_permissions rk ON ur.role_id = rk.role_id
			WHERE ur.user_id = ?;
			`
	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	permissions := make([]models.Permission, 0)
	for rows.Next() {
		var k, v string
		err = rows.Scan(&k, &v)
		if err != nil {
			return nil, err
		}
		permissions = append(permissions, models.Permission{Key: k, Val: v})
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return permissions, nil
}

func (r *mySqlRepo) CreateRole(ctx context.Context, role models.Role, permissions []models.Permission) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	roleQuery := "INSERT INTO roles (id, name) VALUES(?, ?);"
	_, err = tx.ExecContext(ctx, roleQuery, role.ID, role.Name)
	if err != nil {
		tx.Rollback()
		if e, ok := err.(*mysql.MySQLError); ok && (e.Number == mysqlErrDuplicateEntry || e.Number == mysqlErrDuplicateKey) {
			return ErrAlreadyExists
		}
		return err
	}

	for _, p := range permissions {
		roleKVQuery := "INSERT INTO role_permissions (role_id, p_key, p_val) VALUES(?, ?, ?);"
		_, err = tx.ExecContext(ctx, roleKVQuery, role.ID, p.Key, p.Val)
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

func (r *mySqlRepo) AssignRole(ctx context.Context, userID string, roleID string) error {
	query := "INSERT INTO user_roles (user_id, role_id) VALUES(?, ?);"
	_, err := r.db.ExecContext(ctx, query, userID, roleID)
	if err != nil {
		if e, ok := err.(*mysql.MySQLError); ok && (e.Number == mysqlErrDuplicateEntry || e.Number == mysqlErrDuplicateKey) {
			return ErrAlreadyExists
		}
		return err
	}

	return nil
}

func (r *mySqlRepo) RemoveRole(ctx context.Context, userID string, roleID string) error {
	query := "DELETE FROM user_roles WHERE user_id = ? AND role_id = ?;"
	_, err := r.db.ExecContext(ctx, query, userID, roleID)
	if err != nil {
		return err
	}

	return nil
}

func (r *mySqlRepo) GetRole(ctx context.Context, roleID string) (models.Role, error) {
	query := "SELECT id, name FROM roles WHERE id = ?;"
	row := r.db.QueryRowContext(ctx, query, roleID)

	var role models.Role
	err := row.Scan(&role.ID, &role.Name)
	if err != nil {
		if err == sql.ErrNoRows {
			return models.Role{}, ErrNotFound
		}
		return models.Role{}, err
	}

	// permissions, err := r.getRolePermissions(ctx, role.ID)
	// if err != nil {
	// 	return nil, err
	// }
	// role.Permissions = permissions

	return role, nil
}

func (r *mySqlRepo) GetPermissionsOfRole(ctx context.Context, roleID string) ([]models.Permission, error) {
	query := "SELECT p_key, p_val FROM role_permissions WHERE role_id = ?;"
	rows, err := r.db.QueryContext(ctx, query, roleID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	permissions := make([]models.Permission, 0)
	for rows.Next() {
		var k, v string
		err = rows.Scan(&k, &v)
		if err != nil {
			return nil, err
		}
		permissions = append(permissions, models.Permission{Key: k, Val: v})
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return permissions, nil
}

func (r *mySqlRepo) GetRolesOfUser(ctx context.Context, userID string) ([]models.Role, error) {
	query := `
		SELECT r.id, r.name
		FROM user_roles ur
		JOIN roles r ON ur.role_id = r.id
		WHERE ur.user_id = ?;
	`
	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	roles := make([]models.Role, 0)
	for rows.Next() {
		var role models.Role
		if err := rows.Scan(&role.ID, &role.Name); err != nil {
			return nil, err
		}

		// permissions, err := r.getRolePermissions(ctx, role.ID)
		// if err != nil {
		// 	return nil, err
		// }
		// role.Permissions = permissions

		roles = append(roles, role)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return roles, nil
}

func (r *mySqlRepo) ListRoles(ctx context.Context, _ ListRolesParams) ([]models.Role, error) {
	query := "SELECT id, name FROM roles;"
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	roles := make([]models.Role, 0)
	for rows.Next() {
		var role models.Role
		if err := rows.Scan(&role.ID, &role.Name); err != nil {
			return nil, err
		}

		// permissions, err := r.getRolePermissions(ctx, role.ID)
		// if err != nil {
		// 	return nil, err
		// }
		// role.Permissions = permissions

		roles = append(roles, role)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return roles, nil
}
