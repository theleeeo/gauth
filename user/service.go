package user

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

func (s *Service) Create(ctx context.Context, user models.User, provider models.UserProvider) (User, error) {
	if user.Email == "" {
		return User{}, fmt.Errorf("missing user email")
	}

	if provider.Type == "" {
		return User{}, fmt.Errorf("missing user provider type")
	}

	if provider.UserID == "" {
		return User{}, fmt.Errorf("missing user provider id")
	}

	user.ID = uuid.NewString()

	if err := s.repo.CreateUser(ctx, user, provider); err != nil {
		return User{}, err
	}

	return User{
		User: user,
		repo: s.repo,
	}, nil
}

func (s *Service) Get(ctx context.Context, params repo.GetUserParams) (User, error) {
	user, err := s.repo.GetUser(ctx, params)
	if err != nil {
		return User{}, err
	}

	return User{
		User: user,
		repo: s.repo,
	}, nil
}

func (s *Service) GetByProviderID(ctx context.Context, providerID string) (User, error) {
	user, err := s.repo.GetUserByProviderID(ctx, providerID)
	if err != nil {
		return User{}, err
	}

	return User{
		User: user,
		repo: s.repo,
	}, nil
}

func (s *Service) List(ctx context.Context, params repo.ListUsersParams) ([]User, error) {
	userModels, err := s.repo.ListUsers(ctx, params)
	if err != nil {
		return nil, err
	}

	users := make([]User, len(userModels))
	for i, u := range userModels {
		users[i] = User{
			User: u,
			repo: s.repo,
		}
	}

	return users, nil
}

func (s *Service) GetPermissionsOfUser(ctx context.Context, userID string) ([]models.Permission, error) {
	permissions, err := s.repo.GetPermissionsOfUser(ctx, userID)
	if err != nil {
		return []models.Permission{}, err
	}

	return permissions, nil
}
