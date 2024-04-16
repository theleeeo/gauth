package app

import (
	"context"
	"fmt"

	"github.com/theleeeo/thor/authorizer"
	"github.com/theleeeo/thor/models"
	"github.com/theleeeo/thor/repo"
	"github.com/theleeeo/thor/role"
	"github.com/theleeeo/thor/user"
)

type App struct {
	auth        *authorizer.Authorizer
	userService *user.Service
	roleService *role.Service
}

func New(auth *authorizer.Authorizer, userService *user.Service, roleService *role.Service) *App {
	return &App{
		auth:        auth,
		userService: userService,
		roleService: roleService,
	}
}

func (a *App) PublicKey() []byte {
	return a.auth.PublicKey()
}

func (a *App) DecodeToken(ctx context.Context, token string) (*authorizer.Claims, error) {
	return a.auth.Decode(token)
}

func (a *App) WhoAmI(ctx context.Context, token string) (user.User, error) {
	t, err := a.auth.Decode(token)
	if err != nil {
		return user.User{}, err
	}

	u, err := a.GetUserByID(ctx, t.UserID)
	if err != nil {
		return user.User{}, fmt.Errorf("failed to get user: %w", err)
	}

	return u, nil
}

func (a *App) GetUserByID(ctx context.Context, id string) (user.User, error) {
	// if !sdk.UserIsRole(ctx, models.RoleAdmin) && !sdk.UserIs(ctx, id) {
	// 	return nil, errors.New("forbidden")
	// }

	u, err := a.userService.Get(ctx, repo.GetUserParams{ID: &id})
	if err != nil {
		return user.User{}, fmt.Errorf("failed to get user: %w", err)
	}

	return u, nil
}

func (a *App) GetUserByEmail(ctx context.Context, email string) (user.User, error) {
	// if !sdk.UserIsRole(ctx, models.RoleAdmin) {
	// 	return nil, errors.New("forbidden")
	// }

	u, err := a.userService.Get(ctx, repo.GetUserParams{Email: &email})
	if err != nil {
		return user.User{}, fmt.Errorf("failed to get user: %w", err)
	}

	return u, nil
}

func (a *App) GetUserByProviderID(ctx context.Context, providerID string) (user.User, error) {
	// if !sdk.UserIsRole(ctx, models.RoleAdmin) {
	// 	return nil, errors.New("forbidden")
	// }

	u, err := a.userService.GetByProviderID(ctx, providerID)
	if err != nil {
		return user.User{}, fmt.Errorf("failed to get user: %w", err)
	}

	return u, nil
}

func (a *App) ListUsers(ctx context.Context, params repo.ListUsersParams) ([]user.User, error) {
	// if !sdk.UserIsRole(ctx, models.RoleAdmin) {
	// 	return nil, errors.New("forbidden")
	// }

	users, err := a.userService.List(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}

	return users, nil
}

func (a *App) CreateUser(ctx context.Context, userModel models.User, provider models.UserProvider) (user.User, error) {
	// if !sdk.UserIsRole(ctx, models.RoleAdmin) {
	// 	return nil, errors.New("forbidden")
	// }

	u, err := a.userService.Create(ctx, userModel, provider)
	if err != nil {
		return user.User{}, fmt.Errorf("failed to create user: %w", err)
	}

	return u, nil
}

func (a *App) CreateRole(ctx context.Context, roleModel models.Role, permissions []models.Permission) (role.Role, error) {
	// if !sdk.UserIsRole(ctx, models.RoleAdmin) {
	// 	return nil, errors.New("forbidden")
	// }

	r, err := a.roleService.Create(ctx, roleModel, permissions)
	if err != nil {
		return role.Role{}, fmt.Errorf("failed to create role: %w", err)
	}

	return r, nil
}

func (a *App) GetRoleByID(ctx context.Context, id string) (role.Role, error) {
	// if !sdk.UserIsRole(ctx, models.RoleAdmin) {
	// 	return nil, errors.New("forbidden")
	// }

	r, err := a.roleService.Get(ctx, id)
	if err != nil {
		return role.Role{}, fmt.Errorf("failed to get role: %w", err)
	}

	return r, nil
}

func (a *App) ListRoles(ctx context.Context, params repo.ListRolesParams) ([]role.Role, error) {
	// if !sdk.UserIsRole(ctx, models.RoleAdmin) {
	// 	return nil, errors.New("forbidden")
	// }

	roles, err := a.roleService.List(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to list roles: %w", err)
	}

	return roles, nil
}

func (a *App) AssignRole(ctx context.Context, userID, roleID string) error {
	// if !sdk.UserIsRole(ctx, models.RoleAdmin) {
	// 	return errors.New("forbidden")
	// }

	u, err := a.userService.Get(ctx, repo.GetUserParams{ID: &userID})
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}

	if err := u.AssignRole(ctx, roleID); err != nil {
		return fmt.Errorf("failed to assign role: %w", err)
	}

	return nil
}

func (a *App) RemoveRole(ctx context.Context, userID, roleID string) error {
	// if !sdk.UserIsRole(ctx, models.RoleAdmin) {
	// 	return errors.New("forbidden")
	// }

	u, err := a.userService.Get(ctx, repo.GetUserParams{ID: &userID})
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}

	if err := u.RemoveRole(ctx, roleID); err != nil {
		return fmt.Errorf("failed to remove role: %w", err)
	}

	return nil
}

func (a *App) GetRolesOfUser(ctx context.Context, userID string) ([]role.Role, error) {
	// if !sdk.UserIsRole(ctx, models.RoleAdmin) {
	// 	return nil, errors.New("forbidden")
	// }

	roles, err := a.roleService.GetRolesOfUser(ctx, userID)
	if err != nil {
		return []role.Role{}, fmt.Errorf("failed to remove role: %w", err)
	}

	return roles, nil
}

func (a *App) GetPermissionsOfRole(ctx context.Context, roleID string) ([]models.Permission, error) {
	// if !sdk.UserIsRole(ctx, models.RoleAdmin) {
	// 	return nil, errors.New("forbidden")
	// }

	permissions, err := a.roleService.GetPermissionsOfRole(ctx, roleID)
	if err != nil {
		return []models.Permission{}, fmt.Errorf("failed to get permissions of role: %w", err)
	}

	return permissions, nil
}

func (a *App) GetPermissionsOfUser(ctx context.Context, userID string) ([]models.Permission, error) {
	// if !sdk.UserIsRole(ctx, models.RoleAdmin) {
	// 	return nil, errors.New("forbidden")
	// }

	permissions, err := a.userService.GetPermissionsOfUser(ctx, userID)
	if err != nil {
		return []models.Permission{}, fmt.Errorf("failed to get permissions of user: %w", err)
	}

	return permissions, nil
}
