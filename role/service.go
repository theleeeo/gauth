package role

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/theleeeo/thor/models"
	"github.com/theleeeo/thor/repo"
)

type Service struct {
	repo repo.Repo
}

func NewService(repo repo.Repo) *Service {
	return &Service{
		repo: repo,
	}
}

func (s *Service) Create(ctx context.Context, role models.Role, permissions []models.Permission) (Role, error) {
	if role.Name == "" {
		return Role{}, fmt.Errorf("missing role name")
	}

	if len(permissions) == 0 {
		return Role{}, fmt.Errorf("missing role permissions")
	}

	for i, p := range permissions {
		if p.Key == "" {
			return Role{}, fmt.Errorf("missing permission key on permission %d", i+1)
		}

		if p.Val == "" {
			return Role{}, fmt.Errorf("missing permission value on permission %d", i+1)
		}
	}

	role.ID = uuid.NewString()

	if err := s.repo.CreateRole(ctx, role, permissions); err != nil {
		return Role{}, err
	}

	return Role{
		Role: role,
		repo: s.repo,
	}, nil
}

func (s *Service) Get(ctx context.Context, id string) (Role, error) {
	role, err := s.repo.GetRole(ctx, id)
	if err != nil {
		return Role{}, err
	}

	return Role{
		Role: role,
		repo: s.repo,
	}, nil
}

func (s *Service) GetRolesOfUser(ctx context.Context, userID string) ([]Role, error) {
	roleModels, err := s.repo.GetRolesOfUser(ctx, userID)
	if err != nil {
		return []Role{}, err
	}

	roles := make([]Role, len(roleModels))
	for i, r := range roleModels {
		roles[i] = Role{
			Role: r,
			repo: s.repo,
		}
	}

	return roles, nil
}

func (s *Service) List(ctx context.Context, params repo.ListRolesParams) ([]Role, error) {
	roleModels, err := s.repo.ListRoles(ctx, params)
	if err != nil {
		return []Role{}, err
	}

	roles := make([]Role, len(roleModels))
	for i, r := range roleModels {
		roles[i] = Role{
			Role: r,
			repo: s.repo,
		}
	}

	return roles, nil
}

func (s *Service) GetPermissionsOfRole(ctx context.Context, id string) ([]models.Permission, error) {
	permissions, err := s.repo.GetPermissionsOfRole(ctx, id)
	if err != nil {
		return []models.Permission{}, err
	}

	return permissions, nil
}
