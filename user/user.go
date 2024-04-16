package user

import (
	"context"
	"fmt"

	"github.com/theleeeo/thor/models"
	"github.com/theleeeo/thor/repo"
)

type User struct {
	models.User
	repo repo.Repo

	// Through which providers the user is authenticated
	providers []models.UserProvider

	roles []models.Role

	// Permission key-value pairs of the user
	permissions []models.Permission
}

func (u *User) Providers(ctx context.Context) ([]models.UserProvider, error) {
	if u.providers != nil {
		return u.providers, nil
	}

	providers, err := u.repo.GetProvidersOfUser(ctx, u.ID)
	if err != nil {
		return nil, fmt.Errorf("error getting the providers of the user: %w", err)
	}

	u.providers = providers

	return providers, err
}

func (u *User) Roles(ctx context.Context) ([]models.Role, error) {
	if u.roles != nil {
		return u.roles, nil
	}

	roles, err := u.repo.GetRolesOfUser(ctx, u.ID)
	if err != nil {
		return nil, fmt.Errorf("error getting the roles of the user: %w", err)
	}

	u.roles = roles

	return roles, err
}

func (u *User) Permissions(ctx context.Context) ([]models.Permission, error) {
	if u.permissions != nil {
		return u.permissions, nil
	}

	permissions, err := u.repo.GetPermissionsOfUser(ctx, u.ID)
	if err != nil {
		return nil, fmt.Errorf("error getting the key-value pairs of the user: %w", err)
	}

	u.permissions = permissions

	return permissions, err
}

func (u *User) AddProvider(ctx context.Context, provider models.UserProvider) error {
	if err := u.repo.AddProvider(ctx, u.ID, provider); err != nil {
		return fmt.Errorf("error adding provider to the user: %w", err)
	}

	u.providers = append(u.providers, provider)

	return nil
}

func (u *User) AssignRole(ctx context.Context, roleID string) error {
	if err := u.repo.AssignRole(ctx, u.ID, roleID); err != nil {
		return fmt.Errorf("error assigning role to the user: %w", err)
	}

	u.roles = append(u.roles, models.Role{ID: roleID})

	return nil
}

func (u *User) RemoveRole(ctx context.Context, roleID string) error {
	if err := u.repo.RemoveRole(ctx, u.ID, roleID); err != nil {
		return fmt.Errorf("error removing role from the user: %w", err)
	}

	for i, r := range u.roles {
		if r.ID == roleID {
			u.roles = append(u.roles[:i], u.roles[i+1:]...)
			break
		}
	}

	return nil
}
