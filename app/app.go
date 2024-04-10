package app

import (
	"context"
	"errors"
	"fmt"

	"github.com/theleeeo/thor/authorizer"
	"github.com/theleeeo/thor/models"
	"github.com/theleeeo/thor/repo"
	"github.com/theleeeo/thor/sdk"
	"github.com/theleeeo/thor/user"
)

type App struct {
	auth  *authorizer.Authorizer
	users *user.Service
}

func New(authSrv *authorizer.Authorizer, userSrv *user.Service) *App {
	return &App{
		auth:  authSrv,
		users: userSrv,
	}
}

func (a *App) PublicKey() []byte {
	return a.auth.PublicKey()
}

func (a *App) CreateToken(ctx context.Context, user *models.User) (string, error) {
	return a.auth.CreateToken(user)
}

func (a *App) DecodeToken(ctx context.Context, token string) (*authorizer.Claims, error) {
	return a.auth.Decode(token)
}

func (a *App) WhoAmI(ctx context.Context, token string) (*models.User, error) {
	t, err := a.auth.Decode(token)
	if err != nil {
		return nil, err
	}

	user, err := a.GetUserByID(ctx, t.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return user, nil
}

func (a *App) AddUserProvider(ctx context.Context, userID string, provider models.UserProvider) error {
	return a.users.AddProvider(ctx, userID, provider)
}

func (a *App) GetUserByID(ctx context.Context, id string) (*models.User, error) {
	if !sdk.UserIsRole(ctx, models.RoleAdmin) && !sdk.UserIs(ctx, id) {
		return nil, errors.New("forbidden")
	}

	u, err := a.users.Get(ctx, repo.GetUserParams{ID: &id})
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return &u.User, nil
}

func (a *App) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	if !sdk.UserIsRole(ctx, models.RoleAdmin) {
		return nil, errors.New("forbidden")
	}

	u, err := a.users.Get(ctx, repo.GetUserParams{Email: &email})
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return &u.User, nil
}

func (a *App) GetUserByProviderID(ctx context.Context, providerID string) (*models.User, error) {
	if !sdk.UserIsRole(ctx, models.RoleAdmin) {
		return nil, errors.New("forbidden")
	}

	u, err := a.users.GetByProviderID(ctx, providerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return &u.User, nil
}

func (a *App) CreateUser(ctx context.Context, user *models.User) (*models.User, error) {
	if !sdk.UserIsRole(ctx, models.RoleAdmin) {
		return nil, errors.New("forbidden")
	}

	u, err := a.users.Create(ctx, user)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return &u.User, nil
}
