package oauth

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var _ jwt.Claims = &googleClaims{}

type googleClaims struct {
	Issuer              string `json:"iss"`
	AuthorizedPresenter string `json:"azp"`
	Audience            string `json:"aud"`
	Subject             string `json:"sub"`
	AccessTokenHash     string `json:"at_hash"`
	HostedDomain        string `json:"hd"`
	Email               string `json:"email"`
	EmailVerified       bool   `json:"email_verified"`
	IssuedAt            int    `json:"iat"`
	ExpiresAt           int    `json:"exp"`
	Nonce               string `json:"nonce"`
	// First name
	Family_name string `json:"family_name"`
	// Last name
	Given_name string `json:"given_name"`
	// Picture URL
	Picture string `json:"picture"`
}

func (g *googleClaims) GetAudience() (jwt.ClaimStrings, error) {
	return jwt.ClaimStrings{g.Audience}, nil
}

func (g *googleClaims) GetExpirationTime() (*jwt.NumericDate, error) {
	return jwt.NewNumericDate(time.Unix(int64(g.ExpiresAt), 0)), nil
}

func (g *googleClaims) GetIssuedAt() (*jwt.NumericDate, error) {
	return jwt.NewNumericDate(time.Unix(int64(g.IssuedAt), 0)), nil
}

func (g *googleClaims) GetIssuer() (string, error) {
	return g.Issuer, nil
}

func (g *googleClaims) GetNotBefore() (*jwt.NumericDate, error) {
	panic("unimplemented")
}

func (g *googleClaims) GetSubject() (string, error) {
	return g.Subject, nil
}
