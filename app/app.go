package app

import (
	"github.com/theleeeo/thor/authorizer"
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

func (a *App) CreateToken(userID string, role authorizer.Role) (string, error) {
	return a.auth.CreateToken(userID, role)
}

func (a *App) DecodeToken(token string) (*authorizer.Claims, error) {
	return a.auth.Decode(token)
}
