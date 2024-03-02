package user

import (
	"context"

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

func (s *Service) CreateUser(ctx context.Context, user *models.User) (*User, error) {
	err := s.repo.CreateUser(ctx, user)
	if err != nil {
		return nil, err
	}

	return &User{
		User: *user,
	}, nil
}

func (s *Service) GetUser(ctx context.Context, id string) (*User, error) {
	u, err := s.repo.GetUser(ctx, id)
	if err != nil {
		return nil, err
	}

	return &User{
		User: *u,
	}, nil
}
