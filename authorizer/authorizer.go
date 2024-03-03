package authorizer

import (
	"crypto"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/theleeeo/thor/models"
)

type Authorizer struct {
	privateKey    crypto.PrivateKey
	publicKey     crypto.PublicKey
	rawPublicKey  []byte
	validDuration time.Duration

	parser *jwt.Parser
}

func New(privateKey, publicKey []byte, validDuration time.Duration) (*Authorizer, error) {
	pub, err := jwt.ParseEdPublicKeyFromPEM(publicKey)
	if err != nil {
		return nil, fmt.Errorf("failed to parse public key: %w", err)
	}

	priv, err := jwt.ParseEdPrivateKeyFromPEM(privateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key: %w", err)
	}

	return &Authorizer{
		privateKey:    priv,
		publicKey:     pub,
		rawPublicKey:  publicKey,
		validDuration: validDuration,
		parser:        jwt.NewParser(jwt.WithValidMethods([]string{jwt.SigningMethodEdDSA.Alg()}), jwt.WithExpirationRequired()),
	}, nil
}

func (a *Authorizer) PublicKey() []byte {
	return a.rawPublicKey
}

func (a *Authorizer) Decode(token string) (*Claims, error) {
	t, err := a.parser.Parse(token, func(token *jwt.Token) (interface{}, error) {
		return a.publicKey, nil
	})

	if err != nil {
		return nil, err
	}

	claims, ok := t.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("invalid claims")
	}

	return &Claims{
		UserID:     claims["sub"].(string),
		Expiration: time.Unix(int64(claims["exp"].(float64)), 0),
		Role:       Role(claims["role"].(string)),
	}, nil
}

func (a *Authorizer) CreateToken(user *models.User) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodEdDSA,
		jwt.MapClaims{
			"sub":  user.ID,
			"exp":  time.Now().Add(a.validDuration).Unix(),
			"role": user.Role,
		})

	tokenString, err := token.SignedString(a.privateKey)
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return tokenString, nil
}
