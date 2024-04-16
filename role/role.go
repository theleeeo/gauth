package role

import (
	"context"
	"fmt"

	"github.com/theleeeo/thor/models"
	"github.com/theleeeo/thor/repo"
)

type Role struct {
	models.Role
	repo repo.Repo

	permissions []models.Permission
}

func (r *Role) Permissions(ctx context.Context) ([]models.Permission, error) {
	if r.permissions != nil {
		return r.permissions, nil
	}

	permissions, err := r.repo.GetPermissionsOfRole(ctx, r.ID)
	if err != nil {
		return nil, fmt.Errorf("error getting the key-value pairs of the user: %w", err)
	}

	r.permissions = permissions

	return permissions, err
}
