package user

import (
	"context"

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

func (s *Service) Create(ctx context.Context, user *models.User) (*models.User, error) {
	user.ID = uuid.NewString()
	user.Role = "user"

	if err := s.repo.CreateUser(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *Service) Get(ctx context.Context, params repo.GetUserParams) (*models.User, error) {
	return s.repo.GetUser(ctx, params)
}

func (s *Service) GetByProviderID(ctx context.Context, providerID string) (*models.User, error) {
	return s.repo.GetUserByProviderID(ctx, providerID)
}

func (s *Service) AddProvider(ctx context.Context, userID string, provider models.UserProvider) error {
	return s.repo.AddProvider(ctx, userID, provider)
}
